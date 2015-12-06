package db

// Level DB wrapper over multiple service data store.
// Each key has a prefix that combines service name as well as type of stored data.
// The following format is used for item keys:
// somename:m:itemid
// This format is used for payload keys:
// somename:p:payload

import (
	"firempq/common"
	"firempq/conf"
	"firempq/log"
	"sync"
	"time"

	"github.com/jmhodges/levigo"
)

// Default LevelDB read options.
var defaultReadOptions = levigo.NewReadOptions()

// Default LevelDB write options.
var defaultWriteOptions = levigo.NewWriteOptions()

// DataStorage A high level structure on top of LevelDB.
// It caches item storing them into database
// as multiple large batches later.
type DataStorage struct {
	db             *levigo.DB        // Pointer the the instance of level db.
	dbName         string            // LevelDB database name.
	itemCache      map[string][]byte // Active cache for item metadata.
	tmpItemCache   map[string][]byte // Active cache during flush operation.
	cacheLock      sync.Mutex        // Used for caches access.
	flushLock      sync.Mutex        // Used to prevent double flush.
	closed         bool
	flushSync      *sync.WaitGroup // Use to wait until flush happens.
	forceFlushChan chan bool
}

// NewDataStorage is a constructor of DataStorage.
func NewDataStorage(dbName string) (*DataStorage, error) {
	ds := DataStorage{
		dbName:         dbName,
		itemCache:      make(map[string][]byte),
		tmpItemCache:   nil,
		closed:         false,
		forceFlushChan: make(chan bool, 1),
		flushSync:      &sync.WaitGroup{},
	}

	// LevelDB write options.
	opts := levigo.NewOptions()
	opts.SetCreateIfMissing(true)
	opts.SetWriteBufferSize(10 * 1024 * 1024)
	opts.SetCompression(levigo.SnappyCompression)

	db, err := levigo.Open(dbName, opts)
	if err != nil {
		return nil, err
	}
	ds.db = db
	ds.flushSync.Add(1)
	go ds.periodicCacheFlush()
	return &ds, nil
}

func (ds *DataStorage) periodicCacheFlush() {
	for !ds.closed {
		select {
		case <-ds.forceFlushChan:
			break
		case <-time.After(conf.CFG.DbFlushInterval * time.Millisecond):
			break
		}
		ds.flushLock.Lock()
		oldFlushSync := ds.flushSync
		ds.flushSync = &sync.WaitGroup{}
		ds.flushSync.Add(1)
		if !ds.closed {
			ds.flushCache()
		}
		oldFlushSync.Done()
		ds.flushLock.Unlock()
	}
	ds.flushSync.Done()
}

// WaitFlush waits until all data is flushed on disk.
func (ds *DataStorage) WaitFlush() {
	ds.flushLock.Lock()
	s := ds.flushSync
	ds.flushLock.Unlock()
	s.Wait()
}

// CachedStoreItem stores data into the cache.
func (ds *DataStorage) FastStoreData(id string, data []byte) {
	ds.cacheLock.Lock()
	ds.itemCache[id] = data
	ds.cacheLock.Unlock()
}

// CachedStoreItemWithPayload stores data into the cache.
func (ds *DataStorage) FastStoreData2(id1 string, data1 []byte, id2 string, data2 []byte) {
	ds.cacheLock.Lock()
	ds.itemCache[id1] = data1
	ds.itemCache[id2] = data2
	ds.cacheLock.Unlock()
}

// DeleteServiceData deletes all service data such as service metadata, items and payloads.
func (ds *DataStorage) DeleteDataWithPrefix(prefix string) int {
	ds.flushLock.Lock()
	defer ds.flushLock.Unlock()
	ds.flushCache()

	limitCounter := 0
	total := 0
	iter := ds.IterData(prefix)
	wb := levigo.NewWriteBatch()

	for iter.Valid() {
		total++
		if limitCounter < 1000 {
			wb.Delete(iter.Key)
			limitCounter++
		} else {
			limitCounter = 0
			ds.db.Write(defaultWriteOptions, wb)
			wb = levigo.NewWriteBatch()
		}
		iter.Next()
	}

	ds.db.Write(defaultWriteOptions, wb)
	return total
}

// FlushCache flushes all cache into database.
func (ds *DataStorage) flushCache() {
	ds.cacheLock.Lock()
	ds.tmpItemCache = ds.itemCache
	ds.itemCache = make(map[string][]byte)
	ds.cacheLock.Unlock()

	wb := levigo.NewWriteBatch()
	for k, v := range ds.tmpItemCache {
		key := common.UnsafeStringToBytes(k)
		if v == nil {
			wb.Delete(key)
		} else {
			wb.Put(key, v)
		}
	}
	ds.db.Write(defaultWriteOptions, wb)
}

// IterServiceItems returns new over all service item metadata.
// Service name used as a prefix to file all service items.
func (ds *DataStorage) IterData(prefix string) *ItemIterator {
	iter := ds.db.NewIterator(defaultReadOptions)
	return makeItemIterator(iter, common.UnsafeStringToBytes(prefix))
}

// SaveServiceMeta stores service metadata into database.
func (ds *DataStorage) StoreData(id string, data []byte) error {
	return ds.db.Put(defaultWriteOptions, common.UnsafeStringToBytes(id), data)
}

func (ds *DataStorage) DeleteData(id string) {
	ds.db.Delete(defaultWriteOptions, common.UnsafeStringToBytes(id))
}

// GetPayload returns item payload. Three places are checked:
// 1. Top level cache.
// 2. Temp cache while data is getting flushed into db.
// 3. If not found in cache, will mane a DB lookup.
func (ds *DataStorage) GetData(id string) []byte {
	ds.cacheLock.Lock()
	data, ok := ds.itemCache[id]
	if ok {
		ds.cacheLock.Unlock()
		return data
	}
	data, ok = ds.tmpItemCache[id]
	if ok {
		ds.cacheLock.Unlock()
		return data
	}
	ds.cacheLock.Unlock()
	value, _ := ds.db.Get(defaultReadOptions, common.UnsafeStringToBytes(id))
	return value
}

// DeleteItem deletes item metadata and payload, affects cache only until flushed.
func (ds *DataStorage) FastDeleteData(id string) {
	ds.cacheLock.Lock()
	ds.itemCache[id] = nil
	ds.cacheLock.Unlock()
}

// DeleteItem deletes item metadata and payload, affects cache only until flushed.
func (ds *DataStorage) FastDeleteData2(id1, id2 string) {
	ds.cacheLock.Lock()
	ds.itemCache[id1] = nil
	ds.itemCache[id2] = nil
	ds.cacheLock.Unlock()
}

// Close flushes data on disk and closes database.
func (ds *DataStorage) Close() {
	ds.flushLock.Lock()
	defer ds.flushLock.Unlock()
	if !ds.closed {
		ds.flushCache()
		ds.closed = true
		ds.db.Close()
	} else {
		log.Error("Attempt to close database more than once!")
	}
}
