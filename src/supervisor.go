package supervisor

import (
	"log"
	. "sync"
)

type Supervisor struct {
	serviceSpec map [string] *ServiceSpec
	stopSignLock Mutex
	stopSign bool
	startedLock Mutex
	started bool
}

func newSupervisor() *Supervisor {
	sup := new(Supervisor)
	sup.serviceSpec = make(map[string] *ServiceSpec)
	return sup
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
	DIEALSO
)

type ServiceSpec struct {
	service Service
	restartPolicy Policy
	ping chan bool
	// TODO(jwall): service spec should have a count field as well?
}

func (sup *Supervisor) RegisterService(name string, s *ServiceSpec) {
	if sup.started {
		log.Panic("Attempt to register service while supervisor started")
	}
	if !(s.service == nil) {
		sup.serviceSpec[name] = s
	} else {
		log.Panicf("No Service in service spec: %s", s)
	}
	
}

func (sup *Supervisor) UnregisterService(name string) bool {
	if sup.started {
		log.Panic("Attempt to unregister service while supervisor started")
	}
	// check if the key exists
	if _, exists := sup.serviceSpec[name]; !exists {
		return false // return false if it didn't
	}
	sup.serviceSpec[name] = &ServiceSpec{}, false // delete the key
	return true
}

func serviceStarter(s *ServiceSpec) bool {
	ch, result := s.service.Start()
	s.ping = ch
	return result
}

func (sup *Supervisor) SetStarted(b bool) {
	sup.startedLock.Lock()
	sup.started = b
	sup.startedLock.Unlock()
}

func (sup *Supervisor) Start() (chan bool, bool) { // A supervisor is a service
	result := sup.doForServices(serviceStarter)
	sup.SetStarted(true)
	sup.SetStopSign(false)
	if !result {
		return nil, false
	}

	ping := make(chan bool)	

	//run supervisor loop
	go sup.Loop(ping)
	return ping, true
}

func (sup *Supervisor) Loop(ch chan bool) {
	defer close(ch)
	defer sup.SetStarted(false)

	for true {
		result := sup.doForServices(func(s *ServiceSpec) bool {
			// check for service aliveness
			restart := false
			ch := s.ping
			log.Printf("the ping channel is: %s", s.ping)
			if ch == nil || closed(ch) {
				restart = true
			} else {
				ch<- true // send a ping
				healthy, ok := <-ch // listen for response
				if ok && !healthy {
					restart = true
				} else if !ok {
					restart = true
			        } else {
					restart = false
				}
			}
			// if restart is needed follow restart policy
			log.Printf("the restart policy is: %s", s.restartPolicy)
			if restart {
				switch s.restartPolicy {
					case ALWAYS:
						serviceStarter(s)
					case DIEALSO:
					 	return false
				}
			}
			// return true if supervisor tree is still valid
			return true
		})
		if !result {
			sup.Stop()
			break
		}
		if sup.stopSign {
			break // time to stop
		}
	}
}

func (sup *Supervisor) SetStopSign(b bool) {
	sup.stopSignLock.Lock()
	sup.stopSign = true
	sup.stopSignLock.Unlock()
}

func (sup *Supervisor) Stop() bool {
	sup.SetStopSign(true)
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
