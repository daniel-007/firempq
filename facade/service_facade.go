package facade

import (
	"firempq/common"
	"firempq/db"
	"firempq/svcerr"
	"github.com/op/go-logging"
	"sync"
)

var log = logging.MustGetLogger("firempq")

type ServiceFacade struct {
	allSvcs  map[string]common.ISvc
	lock     sync.Mutex
	database *db.DataStorage
}

func NewFacade(database *db.DataStorage) *ServiceFacade {
	f := ServiceFacade{
		database: database,
		allSvcs:  make(map[string]common.ISvc),
	}
	f.loadAllServices()
	return &f
}

func (s *ServiceFacade) loadAllServices() {
	for _, sm := range s.database.GetAllServiceMeta() {
		log.Info("Loading service data for: %s", sm.Name)
		objLoader, ok := SVC_LOADER[sm.Stype]
		if !ok {
			log.Error("Unknown service '%s' type: %s", sm.Name, sm.Stype)
			continue
		}
		svcInstance, err := objLoader(s.database, sm.Name)
		if err != nil {
			log.Error("Service '%s' was not loaded because of: %s", sm.Name, err)
		} else {
			s.allSvcs[sm.Name] = svcInstance
		}
	}
}

func (s *ServiceFacade) CreateService(svcType string, svcName string, params map[string]string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.allSvcs[svcName]; ok {
		return svcerr.ERR_SVC_ALREADY_EXISTS
	}
	svcCrt, ok := SVC_CREATOR[svcType]
	if !ok {
		return svcerr.ERR_SVC_UNKNOWN_TYPE
	}
	smeta := common.NewServiceMetaInfo(svcType, 0, svcName)
	s.database.SaveServiceMeta(smeta)

	s.allSvcs[svcName] = svcCrt(svcName, params)

	return nil
}

func (s *ServiceFacade) DropService(svcName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.allSvcs[svcName]
	if !ok {
		return svcerr.ERR_NO_SVC
	}
	delete(s.allSvcs, svcName)
	return nil
}

func (s *ServiceFacade) GetService(name string) (common.ISvc, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	svc, ok := s.allSvcs[name]
	return svc, ok
}

func (s *ServiceFacade) Close() {
	for _, svc := range s.allSvcs {
		svc.Close()
	}
	s.database.FlushCache()
	s.database.Close()
}
