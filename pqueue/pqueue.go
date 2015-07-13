package pqueue

import (
	"firempq/common"
	"firempq/db"
	"firempq/defs"
	"firempq/qerrors"
	"firempq/queues/priority_first"
	"firempq/structs"
	"firempq/util"
	"github.com/op/go-logging"
	"sort"
	"strconv"
	"sync"
	"time"
)

var log = logging.MustGetLogger("firempq")

const (
	ACTION_UNLOCK_BY_ID        = "UNLOCK"
	ACTION_DELETE_LOCKED_BY_ID = "DELLOCKED"
	ACTION_SET_LOCK_TIMEOUT    = "SETLOCKTIMEOUT"
)

type InMemCache struct {
	Msg      []byte
	Payload  string
	ToDelete bool
}

type PQueue struct {
	queueName string
	// Messages which are waiting to be picked up
	availableMsgs *priority_first.PriorityFirstQueue

	// All messages with the ticking counters except those which are inFlight.
	expireHeap *structs.IndexHeap
	// All locked messages
	inFlightHeap *structs.IndexHeap
	// Just a message map message id to the full message data.
	allMessagesMap map[string]*PQMessage

	lock sync.Mutex
	// Used to stop internal goroutine.
	workDone bool
	// Action handlers which are out of standard queue interface.
	actionHandlers map[string](func(map[string]string) error)
	// Instance of the database.
	database *db.DataStorage
	// Queue settings.
	settings *PQueueSettings

	newMsgNotification chan bool
}

func initPQueue(database *db.DataStorage, queueName string, settings *PQueueSettings) *PQueue {
	pq := PQueue{
		allMessagesMap:     make(map[string]*PQMessage),
		availableMsgs:      priority_first.NewActiveQueues(settings.MaxPriority),
		database:           database,
		expireHeap:         structs.NewIndexHeap(),
		inFlightHeap:       structs.NewIndexHeap(),
		queueName:          queueName,
		settings:           settings,
		workDone:           false,
		newMsgNotification: make(chan bool),
	}

	pq.actionHandlers = map[string](func(map[string]string) error){
		ACTION_DELETE_LOCKED_BY_ID: pq.DeleteLockedById,
		ACTION_SET_LOCK_TIMEOUT:    pq.SetLockTimeout,
		ACTION_UNLOCK_BY_ID:        pq.UnlockMessageById,
	}

	pq.loadAllMessages()
	go pq.periodicCleanUp()

	return &pq
}

func NewPQueue(database *db.DataStorage, queueName string, priorities int64, size int64) *PQueue {
	queue := initPQueue(database, queueName, NewPQueueSettings(priorities, size))
	queue.SavePQueue()
	return queue
}

func LoadPQueue(database *db.DataStorage, queueName string) (common.IQueue, error) {
	settings := new(PQueueSettings)
	err := database.GetQueueSettings(settings, queueName)
	if err != nil {
		return nil, err
	}
	pq := initPQueue(database, queueName, settings)
	return pq, nil
}

func (pq *PQueue) SavePQueue() {
	pq.database.SaveQueueSettings(pq.settings, pq.queueName)
}

func (pq *PQueue) GetStatus() map[string]interface{} {
	res := pq.settings.ToMap()
	res["TotalMessages"] = len(pq.allMessagesMap)
	res["InFlightSize"] = pq.inFlightHeap.Len()
	return res
}

func (pq *PQueue) GetQueueType() string {
	return common.QTYPE_PRIORITY_QUEUE
}

// Push message to the queue.
// Pushing message automatically enables auto expiration.
func (pq *PQueue) Push(msg *PQMessage, payload string) error {
	pq.settings.LastPushTs = util.Uts()

	pq.lock.Lock()
	defer pq.lock.Unlock()
	return pq.storeMessage(msg, payload)
}

func (pq *PQueue) PushMessage(msgData map[string]string, payload string) error {
	msg, err := MessageFromMap(msgData)
	if err != nil {
		return err
	}
	return pq.Push(msg, payload)
}

func (pq *PQueue) storeMessage(msg *PQMessage, payload string) error {
	if _, ok := pq.allMessagesMap[msg.Id]; ok {
		return qerrors.ERR_ITEM_ALREADY_EXISTS
	}
	queueLen := pq.availableMsgs.Len()
	pq.allMessagesMap[msg.Id] = msg
	pq.trackExpiration(msg)
	pq.availableMsgs.Push(msg.Id, msg.Priority)
	if 0 == queueLen {
		select {
		case pq.newMsgNotification <- true:
		default: // allows non blocking channel usage if there are no users awaiting wor the message
		}
	}

	pq.database.StoreItem(pq.queueName, msg, payload)
	return nil
}

// Pop first available message.
// Will return nil if there are no messages available.
func (pq *PQueue) Pop() *PQMessage {

	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.availableMsgs.Empty() {
		return nil
	}

	msgId := pq.availableMsgs.Pop()
	msg, ok := pq.allMessagesMap[msgId]

	if !ok {
		return nil
	}
	msg.PopCount += 1
	pq.lockMessage(msg)
	pq.database.UpdateItem(pq.queueName, msg)
	return msg
}

func (pq *PQueue) PopMessage() (common.IMessage, error) {
	msg := pq.Pop()
	if msg == nil {
		return nil, qerrors.ERR_QUEUE_EMPTY
	}
	return msg, nil
}

// Size of one batch of messages for popMessageBatch function
const MESSAGES_BATCH_SIZE_LIMIT = 10

// Will pop 'limit' messages within 'timeoutMsec' time interval. Function will exit on timeout
// even if queue has enough messages.
func (pq *PQueue) PopWait(timeoutMsec, limit int) []common.IMessage {
	if limit > MESSAGES_BATCH_SIZE_LIMIT {
		limit = MESSAGES_BATCH_SIZE_LIMIT
	}
	msgs := make([]common.IMessage, 0, limit)

	popLimitItems := func() {
		for len(msgs) < limit {
			msg := pq.Pop()
			if msg == nil {
				break
			}
			msgs = append(msgs, msg)
		}
	}
	// Try to pop items first time and return them if number of popped items is greater than 0.
	popLimitItems()
	if len(msgs) > 0 {
		return msgs
	}
	// If items not founds lets wait for any incoming.
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Duration(timeoutMsec) * time.Millisecond)
		timeout <- true
	}()

	needExit := false
	for !needExit && len(msgs) == 0 {
		select {
		case t := <-pq.newMsgNotification:
			popLimitItems()
			needExit = len(msgs) == 0 && t
		case t := <-timeout:
			needExit = true || t
			popLimitItems()
		}
	}
	close(timeout)
	return msgs
}

// Returns message payload.
// This method doesn't lock anything. Actually lock happens withing loadPayload structure.
func (pq *PQueue) GetMessagePayload(msgId string) string {
	return pq.database.GetPayload(pq.queueName, msgId)
}

func (pq *PQueue) lockMessage(msg *PQMessage) {
	nowTs := util.Uts()
	pq.settings.LastPopTs = nowTs
	// Increase number of pop attempts.

	msg.UnlockTs = nowTs + pq.settings.PopLockTimeout
	pq.expireHeap.PopById(msg.Id)
	pq.inFlightHeap.PushItem(msg.Id, msg.UnlockTs)
}

// Remove message id from In Flight message heap.
func (pq *PQueue) unflightMessage(msgId string) (*PQMessage, error) {
	msg, ok := pq.allMessagesMap[msgId]
	if !ok {
		return nil, qerrors.ERR_MSG_NOT_EXIST
	}

	hi := pq.inFlightHeap.PopById(msgId)
	if hi == structs.EMPTY_HEAP_ITEM {
		return nil, qerrors.ERR_MSG_NOT_LOCKED
	}

	return msg, nil
}

func (pq *PQueue) DeleteById(msgId string) error {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.inFlightHeap.ContainsId(msgId) {
		return qerrors.ERR_MSG_IS_LOCKED
	}

	if !pq.deleteMessage(msgId) {
		return qerrors.ERR_MSG_NOT_EXIST
	}

	return nil
}

// Queue custom specific handler for the queue type specific features.
func (pq *PQueue) CustomHandler(action string, params map[string]string) error {
	handler, ok := pq.actionHandlers[action]
	if !ok {
		return qerrors.InvalidRequest("Unknown action: " + action)
	}
	return handler(params)
}

// Set a user defined message lock timeout. Only locked message timeout can be set.
func (pq *PQueue) SetLockTimeout(params map[string]string) error {
	msgId, ok := params[defs.PARAM_MSG_ID]
	if !ok {
		return qerrors.ERR_MSG_NOT_DEFINED
	}

	timeoutStr, ok := params[defs.PARAM_MSG_TIMEOUT]

	if !ok {
		return qerrors.ERR_MSG_TIMEOUT_NOT_DEFINED
	}

	timeout, terr := strconv.Atoi(timeoutStr)
	if terr != nil || timeout < 0 || timeout > defs.TIMEOUT_MAX_LOCK {
		return qerrors.ERR_MSG_TIMEOUT_IS_WRONG
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	msg, err := pq.unflightMessage(msgId)
	if err != nil {
		return err
	}

	msg.UnlockTs = util.Uts() + int64(timeout)

	pq.inFlightHeap.PushItem(msgId, msg.UnlockTs)
	pq.database.UpdateItem(pq.queueName, msg)

	return nil
}

// Delete locked message by id.
func (pq *PQueue) DeleteLockedById(params map[string]string) error {
	msgId, ok := params[defs.PARAM_MSG_ID]
	if !ok {
		return qerrors.ERR_MSG_NOT_DEFINED
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	_, err := pq.unflightMessage(msgId)
	if err != nil {
		return err
	}
	pq.deleteMessage(msgId)
	return nil
}

func (pq *PQueue) UnlockMessageById(params map[string]string) error {

	msgId, ok := params[defs.PARAM_MSG_ID]
	if !ok {
		return qerrors.ERR_MSG_NOT_DEFINED
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	// Make sure message exists.
	msg, err := pq.unflightMessage(msgId)
	if err != nil {
		return err
	}
	// Message exists, push it into the front of the queue.
	return pq.returnToFront(msg)
}

func (pq *PQueue) deleteMessage(msgId string) bool {
	if msg, ok := pq.allMessagesMap[msgId]; ok {
		delete(pq.allMessagesMap, msgId)
		pq.availableMsgs.RemoveItem(msgId, msg.Priority)
		pq.database.DeleteItem(pq.queueName, msgId)
		pq.expireHeap.PopById(msgId)
		return true
	}
	return false
}

// Adds message into expiration heap. Not thread safe!
func (pq *PQueue) trackExpiration(msg *PQMessage) {
	ok := pq.expireHeap.PushItem(msg.Id, msg.CreatedTs+int64(pq.settings.MsgTTL))
	if !ok {
		log.Error("Error! Item already exists in the expire heap: %s", msg.Id)
	}
}

// Attempts to return a message into the front of the queue.
// If a number of POP attempts has exceeded, message will be deleted.
func (pq *PQueue) returnToFront(msg *PQMessage) error {
	lim := pq.settings.PopCountLimit
	if lim > 0 {
		if lim <= msg.PopCount {
			pq.deleteMessage(msg.Id)
			return qerrors.ERR_MSG_POP_ATTEMPTS_EXCEEDED
		}
	}
	msg.UnlockTs = 0
	pq.availableMsgs.PushFront(msg.Id)
	pq.trackExpiration(msg)
	pq.database.UpdateItem(pq.queueName, msg)
	return nil
}

// Unlocks all items which exceeded their lock time.
func (pq *PQueue) releaseInFlight() int {
	cur_ts := util.Uts()
	ifHeap := pq.inFlightHeap

	i := 0
	for !(ifHeap.Empty()) && ifHeap.MinElement() < cur_ts {
		i++
		hi := ifHeap.PopItem()
		pqmsg := pq.allMessagesMap[hi.Id]
		pq.returnToFront(pqmsg)
		if i >= MAX_CLEANS_PER_ATTEMPT {
			break
		}
	}
	return i
}

// Remove all items which are completely expired.
func (pq *PQueue) cleanExpiredItems() int {
	cur_ts := util.Uts()
	eh := pq.expireHeap

	i := 0
	for !(eh.Empty()) && eh.MinElement() < cur_ts {
		i++
		hi := eh.PopItem()
		// There are two places to remove expired item:
		// 1. Map of all items - all items map.
		// 2. Available message.
		msg := pq.allMessagesMap[hi.Id]
		pq.availableMsgs.RemoveItem(msg.Id, msg.Priority)
		pq.deleteMessage(msg.Id)
		if i >= MAX_CLEANS_PER_ATTEMPT {
			break
		}
	}
	return i
}

// 1 milliseconds.
const SLEEP_INTERVAL_IF_ITEMS = 1000000

// 1000 items should be enough to not create long locks. In the most good cases clean ups are rare.
const MAX_CLEANS_PER_ATTEMPT = 1000

// How frequently loop should run.
const DEFAULT_UNLOCK_INTERVAL = 100 * 1000000 // 0.1 second

// Remove expired items. Should be running as a thread.
func (pq *PQueue) periodicCleanUp() {
	for !(pq.workDone) {
		var sleepTime time.Duration = DEFAULT_UNLOCK_INTERVAL

		pq.lock.Lock()
		if !(pq.workDone) {
			cleaned := pq.releaseInFlight()
			if cleaned > 0 {
				log.Debug("%d messages returned to the front of the queue.", cleaned)
				sleepTime = SLEEP_INTERVAL_IF_ITEMS
			}
		}
		pq.lock.Unlock()

		pq.lock.Lock()
		if !(pq.workDone) {
			cleaned := pq.cleanExpiredItems()
			if cleaned > 0 {
				log.Debug("%d items expired.", cleaned)
				sleepTime = SLEEP_INTERVAL_IF_ITEMS
			}
		}
		pq.lock.Unlock()
		if !pq.workDone {
			time.Sleep(sleepTime)
		}
	}
}

// Database related data management.
type MessageSlice []*PQMessage

func (p MessageSlice) Len() int           { return len(p) }
func (p MessageSlice) Less(i, j int) bool { return p[i].CreatedTs < p[j].CreatedTs }
func (p MessageSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (pq *PQueue) loadAllMessages() {
	nowTs := util.Uts()
	log.Info("Initializing queue: %s", pq.queueName)
	iter := pq.database.IterQueue(pq.queueName)
	defer iter.Close()

	msgs := MessageSlice{}
	delIds := []string{}

	s := pq.settings
	for iter.Valid() {
		pqmsg := PQMessageFromBinary(string(iter.Key), iter.Value)

		// Store list if message IDs that should be removed.
		if pqmsg.CreatedTs+s.MsgTTL < nowTs ||
			(pqmsg.PopCount >= s.PopCountLimit &&
				s.PopCountLimit > 0) {
			delIds = append(delIds, pqmsg.Id)
		} else {
			msgs = append(msgs, pqmsg)
		}
		iter.Next()
	}
	log.Debug("Loaded %d messages for %s queue", len(msgs), pq.queueName)
	if len(delIds) > 0 {
		log.Debug("%d messages will be removed because of expiration", len(delIds))
		for _, msgId := range delIds {
			pq.database.DeleteItem(pq.queueName, msgId)
		}

	}
	// Sorting data guarantees that messages will be available almost in the same order as they arrived.
	sort.Sort(msgs)

	for _, msg := range msgs {
		pq.allMessagesMap[msg.Id] = msg
		if msg.UnlockTs > nowTs {
			pq.inFlightHeap.PushItem(msg.Id, msg.UnlockTs)
		} else {
			pq.expireHeap.PushItem(msg.Id, msg.CreatedTs+s.MsgTTL)
			pq.availableMsgs.Push(msg.Id, msg.Priority)
		}
	}

	log.Debug("Messages available: %d", pq.expireHeap.Len())
	log.Debug("Messages in flight: %d", pq.inFlightHeap.Len())
}

func (pq *PQueue) DeleteAll() {
	total := 0
	for {
		ids := []string{}
		pq.lock.Lock()
		if len(pq.allMessagesMap) == 0 {
			pq.lock.Unlock()
			break
		}
		for k, _ := range pq.allMessagesMap {
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
	pq.lock.Lock()
	if pq.workDone {
		return
	}
	pq.workDone = true
	pq.lock.Unlock()
}

var _ common.IQueue = &PQueue{}
