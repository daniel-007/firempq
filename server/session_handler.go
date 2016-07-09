package server

import (
	"bytes"
	"net"
	"strconv"
	"sync"

	"github.com/vburenin/firempq/apis"
	"github.com/vburenin/firempq/db"
	"github.com/vburenin/firempq/log"
	"github.com/vburenin/firempq/mpqerr"
	"github.com/vburenin/firempq/mpqproto"
	"github.com/vburenin/firempq/qmgr"
	"github.com/vburenin/firempq/resp"
	"github.com/vburenin/firempq/signals"
	"github.com/vburenin/firempq/utils"
)

const (
	CMD_PING       = "PING"
	CMD_CREATE_SVC = "CRT"
	CMD_DROP_SVC   = "DROP"
	CMD_QUIT       = "QUIT"
	CMD_UNIX_TS    = "TS"
	CMD_LIST       = "LIST"
	CMD_CTX        = "CTX"
	CMD_LOGLEVEL   = "LOGLEVEL"
	CMD_PANIC      = "PANIC"
	CMD_DBSTATS    = "DBSTATS"
)

type FuncHandler func([]string) apis.IResponse

type SessionHandler struct {
	connLock  sync.Mutex
	conn      net.Conn
	active    bool
	ctx       apis.ServiceContext
	stopChan  chan struct{}
	tokenizer *mpqproto.Tokenizer
	svcs      *qmgr.ServiceManager
	buffPool  sync.Pool
}

func NewSessionHandler(conn net.Conn, services *qmgr.ServiceManager) *SessionHandler {
	sh := &SessionHandler{
		conn:      conn,
		tokenizer: mpqproto.NewTokenizer(),
		ctx:       nil,
		active:    true,
		svcs:      services,
		stopChan:  make(chan struct{}),
		buffPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
	sh.QuitListener()
	return sh
}

func (s *SessionHandler) QuitListener() {
	go func() {
		select {
		case <-signals.QuitChan:
			s.Stop()
			s.WriteResponse(mpqerr.ERR_CONN_CLOSING)
			if s.ctx != nil {
				s.ctx.Finish()
			}
			s.conn.Close()
			return
		case <-s.stopChan:
		}
	}()
}

// DispatchConn dispatcher. Entry point to start connection handling.
func (s *SessionHandler) DispatchConn() {
	addr := s.conn.RemoteAddr().String()
	log.Debug("Client connected: %s", addr)
	s.WriteResponse(resp.NewStrResponse("HELLO FIREMPQ-0.1"))
	for s.active {
		cmdTokens, err := s.tokenizer.ReadTokens(s.conn)
		if err == nil {
			resp := s.processCmdTokens(cmdTokens)
			err = s.WriteResponse(resp)
		}
		if err != nil {
			log.LogConnError(err)
			break
		}
	}
	close(s.stopChan)
	if s.ctx != nil {
		s.ctx.Finish()
	}
	s.conn.Close()
	log.Debug("Client disconnected: %s", addr)
}

// Basic token processing that looks for global commands,
// if there is no token match it will look into current context
// to see if there is a processor for the rest of the tokens.
func (s *SessionHandler) processCmdTokens(cmdTokens []string) apis.IResponse {
	if len(cmdTokens) == 0 {
		return resp.OK_RESPONSE
	}

	cmd := cmdTokens[0]
	tokens := cmdTokens[1:]

	switch cmd {
	case CMD_QUIT:
		return s.quitHandler(tokens)
	case CMD_CTX:
		return s.ctxHandler(tokens)
	case CMD_CREATE_SVC:
		return s.createServiceHandler(tokens)
	case CMD_DROP_SVC:
		return s.dropServiceHandler(tokens)
	case CMD_LIST:
		return s.listServicesHandler(tokens)
	case CMD_LOGLEVEL:
		return logLevelHandler(tokens)
	case CMD_PING:
		return pingHandler(tokens)
	case CMD_UNIX_TS:
		return tsHandler(tokens)
	case CMD_PANIC:
		return panicHandler(tokens)
	case CMD_DBSTATS:
		return dbstatHandler(tokens)
	default:
		if s.ctx == nil {
			return mpqerr.InvalidRequest("Unknown command: " + cmd)
		} else {
			return s.ctx.Call(cmd, tokens)
		}
	}
}

// WriteResponse writes apis.IResponse into connection writer.
func (s *SessionHandler) WriteResponse(resp apis.IResponse) error {
	s.connLock.Lock()
	defer s.connLock.Unlock()

	b := s.buffPool.Get().(*bytes.Buffer)
	b.Reset()
	defer s.buffPool.Put(b)

	err := resp.WriteResponse(b)
	err = b.WriteByte('\n')
	if err != nil {
		return err
	}
	_, err = s.conn.Write(b.Bytes())
	return err
}

// Handler that creates a service.
func (s *SessionHandler) createServiceHandler(tokens []string) apis.IResponse {
	if len(tokens) < 1 {
		return mpqerr.InvalidRequest("Service name should be provided")
	}
	if len(tokens) > 1 {
		return mpqerr.InvalidRequest("At least service name should be provided")
	}

	svcName := tokens[0]
	if len(svcName) > 256 {
		return mpqerr.InvalidRequest("Service name can not be longer than 256 characters")
	}

	if !mpqproto.ValidateItemId(svcName) {
		return mpqerr.ERR_ID_IS_WRONG
	}

	_, exists := s.svcs.GetService(svcName)
	if exists {
		return mpqerr.ConflictRequest("Service exists already")
	}

	return s.svcs.CreateService(apis.STYPE_PRIORITY_QUEUE, svcName, tokens[1:])
}

// Drop service.
func (s *SessionHandler) dropServiceHandler(tokens []string) apis.IResponse {
	if len(tokens) == 0 {
		return mpqerr.InvalidRequest("Service name must be provided")
	}
	if len(tokens) > 1 {
		return mpqerr.InvalidRequest("DROP accept service name only")
	}
	svcName := tokens[0]
	res := s.svcs.DropService(svcName)
	return res
}

// Context changer.
func (s *SessionHandler) ctxHandler(tokens []string) apis.IResponse {
	if len(tokens) > 1 {
		return mpqerr.InvalidRequest("CTX accept service name only")
	}

	if len(tokens) == 0 {
		return mpqerr.InvalidRequest("Service name must be provided")
	}

	svcName := tokens[0]
	svc, exists := s.svcs.GetService(svcName)
	if !exists {
		return mpqerr.ERR_NO_SVC
	}
	s.ctx = svc.NewContext(s)
	return resp.OK_RESPONSE
}

// Stop the processing loop.
func (s *SessionHandler) Stop() {
	s.active = false
}

// Stops the main loop on QUIT.
func (s *SessionHandler) quitHandler(tokens []string) apis.IResponse {
	if len(tokens) > 0 {
		return mpqerr.ERR_CMD_WITH_NO_PARAMS
	}
	s.Stop()
	return resp.OK_RESPONSE
}

// List all active services.
func (s *SessionHandler) listServicesHandler(tokens []string) apis.IResponse {
	svcPrefix := ""
	if len(tokens) > 1 {
		return mpqerr.InvalidRequest("LIST accept service name prefix only")
	}
	if len(tokens) == 1 {
		svcPrefix = tokens[0]
	}

	return s.svcs.ListServiceNames(svcPrefix)
}

// Ping responder.
func pingHandler(tokens []string) apis.IResponse {
	if len(tokens) > 0 {
		return mpqerr.ERR_CMD_WITH_NO_PARAMS
	}
	return resp.RESP_PONG
}

// Returns current server unix time stamp in milliseconds.
func tsHandler(tokens []string) apis.IResponse {
	if len(tokens) > 0 {
		return mpqerr.ERR_CMD_WITH_NO_PARAMS
	}
	return resp.NewIntResponse(utils.Uts())
}

func logLevelHandler(tokens []string) apis.IResponse {
	if len(tokens) != 1 {
		return mpqerr.InvalidRequest("Log level accept one integer parameter in range [0-5]")
	}
	l, e := strconv.Atoi(tokens[0])
	if e != nil || l < 0 || l > 5 {
		return mpqerr.InvalidRequest("Log level is an integer in range [0-5]")
	}
	log.Warning("Log level changed to: %d", l)
	log.SetLevel(l)
	return resp.OK_RESPONSE
}

func panicHandler(tokens []string) apis.IResponse {
	if len(tokens) > 0 {
		return mpqerr.ERR_CMD_WITH_NO_PARAMS
	}

	log.Critical("Panic requested!")
	panic("Panic requested")
	return resp.OK_RESPONSE
}

func dbstatHandler(tokens []string) apis.IResponse {
	if len(tokens) > 0 {
		return mpqerr.ERR_CMD_WITH_NO_PARAMS
	}
	db := db.DatabaseInstance()
	return resp.NewDictResponse("+DBSTATS", db.GetStats())
}
