package supervisor

import "log"

type Supervisor struct {
	serviceSpec map [string] *ServiceSpec
	stopSign bool
	started bool
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

func (sup *Supervisor) Start() (chan bool, bool) { // A supervisor is a service
	result := sup.doForServices(serviceStarter)
	sup.started = true
	if !result {
		return nil, false
	}

	ping := make(chan bool)	

	//run supervisor loop
	go sup.Loop(ping)
	return ping, true
}

func (sup *Supervisor) Loop(ch chan bool) {
	for true {
		result := sup.doForServices(func(s *ServiceSpec) bool {
			// check for service aliveness
			restart := false
			ch := s.ping
			if ch == nil || closed(ch) {
				restart = true
			} else {
				ch<- true // send a ping
				ok := <-ch // listen for response
				if !ok {
					restart = true
				}
			}
			// if restart is needed follow restart policy
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
	sup.started = false
	close(ch)
}

func (sup *Supervisor) Stop() bool {
	sup.stopSign = true
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
