package pqueue

import (
	"firempq/common"
	"firempq/conf"
	"firempq/db"
	"firempq/defs"
	"firempq/structs"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("firempq")

// Size of one batch of messages for popMessageBatch function
const MESSAGES_BATCH_SIZE_LIMIT = 20
const MAX_POP_WAIT_TIMEOUT = 30000

const (
	ACTION_UNLOCK_BY_ID        = "UNLOCK"
	ACTION_DELETE_LOCKED_BY_ID = "DELLOCKED"
	ACTION_DELETE_BY_ID        = "DEL"
	ACTION_SET_LOCK_TIMEOUT    = "SETLOCKTIMEOUT"
	ACTION_PUSH                = "PUSH"
	ACTION_POP                 = "POP"
	ACTION_POP_WAIT            = "POPWAIT"
	ACTION_SET_QPARAM          = "SETQP"
	ACTION_SET_MPARAM          = "SETMP"
	ACTION_STATUS              = "STATUS"
	ACTION_RELEASE_IN_FLIGHT   = "RELEASE"
	ACTION_EXPIRE              = "EXPIRE"
)

type PQueue struct {
	queueName string
	// Messages which are waiting to be picked up
	availMsgs *structs.PriorityFirstQueue

	// All messages with the ticking counters except those which are inFlight.
	expireHeap *structs.IndexHeap
	// All locked messages
	inFlightHeap *structs.IndexHeap
	// Just a message map message id to the full message data.
	msgMap map[string]*PQMessage

	lock sync.Mutex
	// Set as True if the service is closed.
	closeFlag common.BoolFlag
	// Action handlers which are out of standard queue interface.
	actionHandlers map[string](common.CallFuncType)
	// Instance of the database.
	database *db.DataStorage
	// Queue settings.
	config *PQConfig

	newMsgNotification chan bool
}

func initPQueue(database *db.DataStorage, queueName string, config *PQConfig) *PQueue {
	pq := PQueue{
		msgMap:             make(map[string]*PQMessage),
		availMsgs:          structs.NewActiveQueues(config.MaxPriority),
		database:           database,
		expireHeap:         structs.NewIndexHeap(),
		inFlightHeap:       structs.NewIndexHeap(),
		queueName:          queueName,
		config:             config,
		newMsgNotification: make(chan bool),
	}

	pq.loadAllMessages()
	return &pq
}

func NewPQueue(database *db.DataStorage, queueName string, priorities int64, size int64) *PQueue {
	defaults := conf.CFG.PQueueConfig
	config := &PQConfig{
		MaxPriority:    priorities,
		MaxSize:        size,
		MsgTtl:         defaults.DefaultMessageTtl,
		DeliveryDelay:  defaults.DefaultDeliveryDelay,
		PopLockTimeout: defaults.DefaultLockTimeout,
		PopCountLimit:  defaults.DefaultPopCountLimit,
		LastPushTs:     common.Uts(),
		LastPopTs:      common.Uts(),
		CreateTs:       common.Uts(),
		InactivityTtl:  0,
	}

	queue := initPQueue(database, queueName, config)
	queue.database.SaveServiceConfig(queueName, config)
	return queue
}

func LoadPQueue(database *db.DataStorage, queueName string) (common.ISvc, error) {
	config := &PQConfig{}
	err := database.LoadServiceConfig(config, queueName)
	if err != nil {
		return nil, err
	}
	pq := initPQueue(database, queueName, config)
	return pq, nil
}

func (pq *PQueue) GetStatus() map[string]interface{} {
	res := make(map[string]interface{})
	res["MaxPriority"] = pq.config.GetMaxPriority()
	res["MaxSize"] = pq.config.GetMaxSize()
	res["MsgTtl"] = pq.config.GetMsgTtl()
	res["DeliveryDelay"] = pq.config.GetDeliveryDelay()
	res["PopLockTimeout"] = pq.config.GetPopLockTimeout()
	res["PopCountLimit"] = pq.config.GetPopCountLimit()
	res["CreateTs"] = pq.config.GetCreateTs()
	res["LastPushTs"] = pq.config.GetLastPushTs()
	res["LastPopTs"] = pq.config.GetLastPopTs()
	res["InactivityTtl"] = pq.config.GetInactivityTtl()
	res["TotalMessages"] = len(pq.msgMap)
	res["InFlightSize"] = pq.inFlightHeap.Len()

	return res
}

func (pq *PQueue) GetCurrentStatus(params []string) common.IResponse {
	if len(params) > 0 {
		return common.ERR_CMD_WITH_NO_PARAMS
	}
	return common.NewDictResponse(pq.GetStatus())
}

func (pq *PQueue) GetType() defs.ServiceType {
	return defs.HT_PRIORITY_QUEUE
}

func (pq *PQueue) GetTypeName() string {
	return common.STYPE_PRIORITY_QUEUE
}

// Call dispatches processing of the command to the appropriate command parser.
func (pq *PQueue) Call(cmd string, params []string) common.IResponse {
	switch cmd {
	case ACTION_POP_WAIT:
		return pq.PopWait(params)
	case ACTION_DELETE_LOCKED_BY_ID:
		return pq.DeleteLockedById(params)
	case ACTION_DELETE_BY_ID:
		return pq.DeleteById(params)
	case ACTION_POP:
		return pq.Pop(params)
	case ACTION_PUSH:
		return pq.Push(params)
	case ACTION_SET_LOCK_TIMEOUT:
		return pq.SetLockTimeout(params)
	case ACTION_UNLOCK_BY_ID:
		return pq.UnlockMessageById(params)
	case ACTION_STATUS:
		return pq.GetCurrentStatus(params)
	case ACTION_RELEASE_IN_FLIGHT:
		return pq.ReleaseInFlight(params)
	case ACTION_EXPIRE:
		return pq.ExpireItems(params)
	}
	return common.InvalidRequest("Unknown action: " + cmd)
}

// Delete all messages in the queue. It includes all type of messages
func (pq *PQueue) Clear() {
	total := 0
	for {
		ids := []string{}
		pq.lock.Lock()
		if len(pq.msgMap) == 0 {
			pq.lock.Unlock()
			break
		}
		for k, _ := range pq.msgMap {
			ids = append(ids, k)
			if len(ids) > 100 {
				break
			}
		}
		total += len(ids)
		for _, id := range ids {
			pq.deleteMessage(id)
		}
		pq.lock.Unlock()
	}
	log.Debug("Removed %d messages.", total)
}

func (pq *PQueue) Close() {
	pq.closeFlag.SetTrue()
}

func (pq *PQueue) IsClosed() bool {
	return pq.closeFlag.Flag
}

func makeUnknownParamResponse(paramName string) *common.ErrorResponse {
	return common.InvalidRequest(fmt.Sprintf("Unknown parameter: %s", paramName))
}

func getMessageIdOnly(params []string) (string, *common.ErrorResponse) {
	var err *common.ErrorResponse
	var msgId string

	for len(params) > 0 {
		switch params[0] {
		case PRM_ID:
			params, msgId, err = common.ParseStringParam(params, 1, 128)
		default:
			return "", makeUnknownParamResponse(params[0])
		}
		if err != nil {
			return "", err
		}
	}

	if len(msgId) == 0 {
		return "", common.ERR_MSG_ID_NOT_DEFINED
	}
	return msgId, nil
}

// Push message to the queue.
// Pushing message automatically enables auto expiration.
func (pq *PQueue) Push(params []string) common.IResponse {
	var err *common.ErrorResponse
	var msgId string
	var priority int64 = pq.config.MaxPriority - 1
	var payload string = ""

	for len(params) > 0 {
		switch params[0] {
		case PRM_ID:
			params, msgId, err = common.ParseStringParam(params, 1, 128)
		case PRM_PRIORITY:
			params, priority, err = common.ParseInt64Params(params, 0, pq.config.MaxPriority-1)
		case PRM_PAYLOAD:
			params, payload, err = common.ParseStringParam(params, 1, 512*1024)
		default:
			return makeUnknownParamResponse(params[0])
		}
		if err != nil {
			return err
		}
	}
	if len(msgId) == 0 {
		msgId = common.GenRandMsgId()
	}

	msg := NewPQMessage(msgId, priority)

	pq.config.LastPushTs = common.Uts()

	pq.lock.Lock()
	defer pq.lock.Unlock()
	return pq.storeMessage(msg, payload)
}

func (pq *PQueue) storeMessage(msg *PQMessage, payload string) common.IResponse {
	if _, ok := pq.msgMap[msg.Id]; ok {
		return common.ERR_ITEM_ALREADY_EXISTS
	}
	queueLen := pq.availMsgs.Len()
	pq.msgMap[msg.Id] = msg
	pq.trackExpiration(msg)
	pq.availMsgs.Push(msg.Id, msg.Priority)
	if 0 == queueLen {
		select {
		case pq.newMsgNotification <- true:
		default: // allows non blocking channel usage if there are no users awaiting wor the message
		}
	}

	pq.database.StoreItemWithPayload(pq.queueName, msg, payload)
	return common.OK200_RESPONSE
}

// Pop first available messages.
// Will return nil if there are no messages available.
func (pq *PQueue) Pop(params []string) common.IResponse {
	var err *common.ErrorResponse
	var limit int64 = 1
	lockTimeout := pq.config.PopLockTimeout

	for len(params) > 0 {
		p := params[0]
		switch p {
		case PRM_LOCK_TIMEOUT:
			params, lockTimeout, err = common.ParseInt64Params(params, 0, 24*1000*3600)
		case PRM_LIMIT:
			params, limit, err = common.ParseInt64Params(params, 1, MESSAGES_BATCH_SIZE_LIMIT)
		default:
			return makeUnknownParamResponse(params[0])
		}
		if err != nil {
			return err
		}
	}
	return common.NewItemsResponse(pq.popMessages(lockTimeout, limit))
}

func (pq *PQueue) tsParamFunc(params []string, f func(int64) int64) common.IResponse {
	var err *common.ErrorResponse
	var ts int64 = -1
	for len(params) > 0 {
		switch params[0] {
		case PRM_TIMESTAMP:
			params, ts, err = common.ParseInt64Params(params, 0, math.MaxInt64)
		default:
			return makeUnknownParamResponse(params[0])
		}
		if err != nil {
			return err
		}
	}
	if ts < 0 {
		return common.ERR_TS_PARAMETER_NEEDED
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()
	return common.NewIntResponse(f(ts))
}

func (pq *PQueue) ExpireItems(params []string) common.IResponse {
	return pq.tsParamFunc(params, pq.cleanExpiredItems)
}

func (pq *PQueue) ReleaseInFlight(params []string) common.IResponse {
	return pq.tsParamFunc(params, pq.releaseInFlight)
}

// PopWait pops up to specified number of messages, if there are no available messages.
// It will wait until either new message is pushed or wait time exceeded.
func (pq *PQueue) PopWait(params []string) common.IResponse {
	var err *common.ErrorResponse
	var limit int64 = 1
	var popWaitTimeout int64 = 1000
	var lockTimeout int64 = pq.config.PopLockTimeout

	for len(params) > 0 {
		switch params[0] {
		case PRM_LOCK_TIMEOUT:
			params, lockTimeout, err = common.ParseInt64Params(params, 0, 24*1000*3600)
		case PRM_LIMIT:
			params, limit, err = common.ParseInt64Params(params, 1, MESSAGES_BATCH_SIZE_LIMIT)
		case PRM_POP_WAIT_TIMEOUT:
			params, popWaitTimeout, err = common.ParseInt64Params(params, 1, MAX_POP_WAIT_TIMEOUT)
		default:
			return makeUnknownParamResponse(params[0])
		}
		if err != nil {
			return err
		}
	}

	return common.NewItemsResponse(pq.popWaitItems(lockTimeout, popWaitTimeout, limit))
}

func (pq *PQueue) popMetaMessage(lockTimeout int64) *PQMessage {
	if pq.availMsgs.Empty() {
		return nil
	}

	msgId := pq.availMsgs.Pop()
	msg, ok := pq.msgMap[msgId]

	if !ok {
		return nil
	}
	msg.PopCount += 1
	pq.lockMessage(msg, lockTimeout)
	pq.database.StoreItem(pq.queueName, msg)

	return msg
}

func (pq *PQueue) popMessages(lockTimeout int64, limit int64) []common.IItem {
	var msgsMeta []*PQMessage
	pq.lock.Lock()
	for int64(len(msgsMeta)) < limit {
		mm := pq.popMetaMessage(lockTimeout)
		if mm == nil {
			break
		} else {
			msgsMeta = append(msgsMeta, mm)
		}
	}
	pq.lock.Unlock()

	var msgs []common.IItem
	for _, mm := range msgsMeta {
		msgs = append(msgs, NewMsgItem(mm, pq.getPayload(mm.Id)))
	}

	return msgs
}

// Will pop 'limit' messages within 'timeout'(milliseconds) time interval.
func (pq *PQueue) popWaitItems(lockTimeout, popWaitTimeout, limit int64) []common.IItem {
	// Try to pop items first time and return them if number of popped items is greater than 0.
	msgItems := pq.popMessages(lockTimeout, limit)
	if len(msgItems) > 0 {
		return msgItems
	}

	// If items not founds lets wait for any incoming.
	timeoutChan := make(chan bool)
	defer close(timeoutChan)

	go func() {
		time.Sleep(time.Duration(popWaitTimeout) * time.Millisecond)
		timeoutChan <- true
	}()

	for {
		select {
		case t := <-pq.newMsgNotification:
			msgItems := pq.popMessages(lockTimeout, limit)
			if len(msgItems) > 0 && t {
				return msgItems
			}
		case t := <-timeoutChan:
			if t {
				return pq.popMessages(lockTimeout, limit)
			}
		}
	}
}

// Returns message payload.
// This method doesn't lock anything. Actually lock happens withing loadPayload structure.
func (pq *PQueue) getPayload(msgId string) string {
	return pq.database.GetPayload(pq.queueName, msgId)
}

func (pq *PQueue) lockMessage(msg *PQMessage, lockTimeout int64) {
	nowTs := common.Uts()
	pq.config.LastPopTs = nowTs
	// Increase number of pop attempts.

	msg.UnlockTs = nowTs + lockTimeout
	pq.expireHeap.PopById(msg.Id)
	pq.inFlightHeap.PushItem(msg.Id, msg.UnlockTs)
}

// Remove message id from In Flight message heap.
func (pq *PQueue) unflightMessage(msgId string) (*PQMessage, *common.ErrorResponse) {
	msg, ok := pq.msgMap[msgId]
	if !ok {
		return nil, common.ERR_MSG_NOT_EXIST
	}

	hi := pq.inFlightHeap.PopById(msgId)
	if hi == structs.EMPTY_HEAP_ITEM {
		return nil, common.ERR_MSG_NOT_LOCKED
	}

	return msg, nil
}

func (pq *PQueue) DeleteById(params []string) common.IResponse {
	msgId, retData := getMessageIdOnly(params)
	if retData != nil {
		return retData
	}
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.inFlightHeap.ContainsId(msgId) {
		return common.ERR_MSG_IS_LOCKED
	}

	if !pq.deleteMessage(msgId) {
		return common.ERR_MSG_NOT_EXIST
	}

	return common.OK200_RESPONSE
}

// Set a user defined message lock timeout. Only locked message timeout can be set.
func (pq *PQueue) SetLockTimeout(params []string) common.IResponse {
	var err *common.ErrorResponse
	var msgId string
	var lockTimeout int64 = -1

	for len(params) > 0 {
		switch params[0] {
		case PRM_ID:
			params, msgId, err = common.ParseStringParam(params, 1, 128)
		case PRM_LOCK_TIMEOUT:
			params, lockTimeout, err = common.ParseInt64Params(params, 0, 24*1000*3600)
		default:
			return makeUnknownParamResponse(params[0])
		}
		if err != nil {
			return err
		}
	}

	if len(msgId) == 0 {
		return common.ERR_MSG_ID_NOT_DEFINED
	}

	if lockTimeout < 0 {
		return common.ERR_MSG_TIMEOUT_NOT_DEFINED
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	msg, err := pq.unflightMessage(msgId)
	if err != nil {
		return err
	}

	msg.UnlockTs = common.Uts() + int64(lockTimeout)

	pq.inFlightHeap.PushItem(msgId, msg.UnlockTs)
	pq.database.StoreItem(pq.queueName, msg)

	return common.OK200_RESPONSE
}

// Delete locked message by id.
func (pq *PQueue) DeleteLockedById(params []string) common.IResponse {
	msgId, retData := getMessageIdOnly(params)
	if retData != nil {
		return retData
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	_, err := pq.unflightMessage(msgId)
	if err != nil {
		return err
	}
	pq.deleteMessage(msgId)
	return common.OK200_RESPONSE
}

func (pq *PQueue) UnlockMessageById(params []string) common.IResponse {
	msgId, retData := getMessageIdOnly(params)
	if retData != nil {
		return retData
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	// Make sure message exists.
	msg, err := pq.unflightMessage(msgId)
	if err != nil {
		return err
	}
	// Message exists, push it into the front of the queue.
	err = pq.returnToFront(msg)
	if err != nil {
		return err
	}
	return common.OK200_RESPONSE
}

func (pq *PQueue) deleteMessage(msgId string) bool {
	if msg, ok := pq.msgMap[msgId]; ok {
		delete(pq.msgMap, msgId)
		pq.availMsgs.RemoveItem(msgId, msg.Priority)
		pq.database.DeleteItem(pq.queueName, msgId)
		pq.expireHeap.PopById(msgId)
		return true
	}
	return false
}

// Adds message into expiration heap. Not thread safe!
func (pq *PQueue) trackExpiration(msg *PQMessage) {
	ok := pq.expireHeap.PushItem(msg.Id, msg.CreatedTs+int64(pq.config.MsgTtl))
	if !ok {
		log.Error("Error! Item already exists in the expire heap: %s", msg.Id)
	}
}

// Attempts to return a message into the front of the queue.
// If a number of POP attempts has exceeded, message will be deleted.
func (pq *PQueue) returnToFront(msg *PQMessage) *common.ErrorResponse {
	lim := pq.config.PopCountLimit
	if lim > 0 {
		if lim <= msg.PopCount {
			pq.deleteMessage(msg.Id)
			return common.ERR_MSG_POP_ATTEMPTS_EXCEEDED
		}
	}
	msg.UnlockTs = 0
	pq.availMsgs.PushFront(msg.Id)
	pq.trackExpiration(msg)
	pq.database.StoreItem(pq.queueName, msg)
	return nil
}

// 1000 items should be enough to not create long locks.
const MAX_CLEANS_PER_ATTEMPT = 1000

// Unlocks all items which exceeded their lock time.
func (pq *PQueue) releaseInFlight(ts int64) int64 {
	ifHeap := pq.inFlightHeap

	var counter int64 = 0
	for !(ifHeap.Empty()) && ifHeap.MinElement() < ts && counter < MAX_CLEANS_PER_ATTEMPT {
		counter++
		unlockedItem := ifHeap.PopItem()
		pqmsg := pq.msgMap[unlockedItem.Id]
		pq.returnToFront(pqmsg)
	}
	if counter > 0 {
		log.Debug("'%d' item(s) returned to the front of the queue.", counter)
	}
	return counter
}

// Remove all items which are completely expired.
func (pq *PQueue) cleanExpiredItems(ts int64) int64 {
	var counter int64 = 0

	eh := pq.expireHeap
	for !(eh.Empty()) && eh.MinElement() < ts && counter < MAX_CLEANS_PER_ATTEMPT {
		counter++
		pq.deleteMessage(eh.PopItem().Id)
	}
	if counter > 0 {
		log.Debug("%d item(s) expired.", counter)
	}
	return counter
}

// How frequently loop should run.
func (pq *PQueue) Update(ts int64) bool {

	params := []string{PRM_TIMESTAMP, strconv.FormatInt(ts, 10)}
	r1, ok1 := pq.Call(ACTION_RELEASE_IN_FLIGHT, params).(*common.IntResponse)
	r2, ok2 := pq.Call(ACTION_EXPIRE, params).(*common.IntResponse)

	return (ok1 && r1.Value > 0) || (ok2 && r2.Value > 0)
}

// Database related data management.
type MessageSlice []*PQMessage

func (p MessageSlice) Len() int           { return len(p) }
func (p MessageSlice) Less(i, j int) bool { return p[i].CreatedTs < p[j].CreatedTs }
func (p MessageSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (pq *PQueue) loadAllMessages() {
	nowTs := common.Uts()
	log.Debug("Initializing queue: %s", pq.queueName)
	iter := pq.database.IterServiceItems(pq.queueName)
	defer iter.Close()

	msgs := MessageSlice{}
	delIds := []string{}

	cfg := pq.config
	for iter.Valid() {
		pqmsg := UnmarshalPQMessage(string(iter.Key), iter.Value)
		// Store list if message IDs that should be removed.
		if pqmsg.CreatedTs+cfg.MsgTtl < nowTs ||
			(pqmsg.PopCount >= cfg.PopCountLimit &&
				cfg.PopCountLimit > 0) {
			delIds = append(delIds, pqmsg.Id)
		} else {
			msgs = append(msgs, pqmsg)
		}
		iter.Next()
	}
	log.Debug("Loaded %d messages for %s queue", len(msgs), pq.queueName)
	if len(delIds) > 0 {
		log.Debug("Deleting %d expired messages", len(delIds))
		for _, msgId := range delIds {
			pq.database.DeleteItem(pq.queueName, msgId)
		}

	}
	// Sorting data guarantees that messages will be available almost in the same order as they arrived.
	sort.Sort(msgs)

	for _, msg := range msgs {
		pq.msgMap[msg.Id] = msg
		if msg.UnlockTs > nowTs {
			pq.inFlightHeap.PushItem(msg.Id, msg.UnlockTs)
		} else {
			pq.expireHeap.PushItem(msg.Id, msg.CreatedTs+cfg.MsgTtl)
			pq.availMsgs.Push(msg.Id, msg.Priority)
		}
	}

	log.Debug("Messages available: %d", pq.expireHeap.Len())
	log.Debug("Messages in flight: %d", pq.inFlightHeap.Len())
}

var _ common.ISvc = &PQueue{}
