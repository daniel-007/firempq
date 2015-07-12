package facade

import (
	"firempq/common"
	"firempq/db"
	"firempq/pqueue"
	"sync"
)

type CreateFactoryFunc func(string, map[string]string) common.IQueue
type DataLoaderFunc func(*db.DataStorage, string) (common.IQueue, error)

var QUEUE_CREATER = map[string](CreateFactoryFunc){
	common.QTYPE_PRIORITY_QUEUE: pqueue.CreatePQueue,
}

var QUEUE_LOADER = map[string](DataLoaderFunc){
	common.QTYPE_PRIORITY_QUEUE: pqueue.LoadPQueue,
}

var facade *QFacade
var lock sync.Mutex

func CreateFacade() *QFacade {
	lock.Lock()
	defer lock.Unlock()
	if facade == nil {
		facade = NewFacade(db.GetDatabase())
	}
	return facade
}
