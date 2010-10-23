package supervisor

import "log"

type Supervisor struct {
	serviceSpec map [string] *ServiceSpec
}

type Service interface {
	// returns true when successfull
	// and a ping channel to check for life
	Start() (chan bool, bool)
	Stop() // return true when successfull
}

type Policy int

const (
	ALWAYS = iota // 0 value and therefore the default for restart policies
	NEVER
)

type ServiceSpec struct {
	service Service
	restartPolicy Policy
	ping chan bool
}

func (sup *Supervisor) RegisterService(name string, s *ServiceSpec) {
	if !(s.service == nil) {
		sup.serviceSpec[name] = s
	} else {
		log.Panicf("No Service is service spec: %s", s)
	}
	
}

func (sup *Supervisor) UnregisterService(name string) bool {
	// check if the key exists
	if _, exists := sup.serviceSpec[name]; !exists {
		return false // return false if it didn't
	}
	sup.serviceSpec[name] = &ServiceSpec{}, false // delete the key
	return true
}

func (sup *Supervisor) Start() bool {
	return sup.doForServices(func(s *ServiceSpec) bool {
	        ch, result := s.service.Start()
		s.ping = ch
		return result
	}) 
}

func (sup *Supervisor) Stop() bool {
	return sup.doForServices(func(s *ServiceSpec) bool {
		s.service.Stop()
		return true
	}) 
}

func (sup *Supervisor) doForServices(f func (s *ServiceSpec) bool) bool {
	result := true
	for _, s := range sup.serviceSpec {
		result = result && f(s)
	}
	return result
}
