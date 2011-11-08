/*
 Copyright 2010 Jeremy Wall (jeremy@marzhillstudios.com)
 Use of this source code is governed by the Artistic License 2.0.
 That License is included in the LICENSE file.

 The supervisor package implements an elementary supervision tree
 for goroutines ala erlangs supervision trees.

 Supervisors monitor their immediate children routines.

 A child routine is defined by a ServiceSpec and is registered with a
 supervisor using RegisterService(name, spec).

 The supervisor monitors the children and restarts them according to a
 restart policy.
 */
package supervisor

import (
	"log"
	. "sync"
)

/*
 A supervisor in a supervision tree. Supervisors should be constructed using
 NewSupervisor() so that default values are filled in correctly.

 Supervisors implement the Service interface so thay can be children of other
 supervisors forming a tree.
 */
type Supervisor struct {
	serviceSpec map [string] *ServiceSpec
	stopSignLock Mutex
	stopSign bool
	startedLock Mutex
	started bool
}

/*
 Constructor for a Supervisor.

 Fills in the default values for a supervisor it.
 */
func NewSupervisor() *Supervisor {
	sup := new(Supervisor)
	sup.serviceSpec = make(map[string] *ServiceSpec)
	return sup
}

/*
 The interface all child services must implement to be part of a supervision
 tree.

 For an example look at the supervisor code. Supervisors implement the Service
 interface so they can be children of a supervision tree themselves
 */
type Service interface {
	/*
	 Returns true when successfull and a ping channel to check for health
	 status from the process.
	
	 The channel should respond to any ping request with a true
	 if the service is healthy or false if it is not. If it responds
	 with false the supervisor will follow the restart policy for that
	 Service's ServiceSpec

	 The Service is responsible for closing this channel which is one
	 of the signals to the supervisor that the service has shutdown.
	 */
	Start() (chan bool, bool)
	/*
	 Stops the supervisor service.

	 Stop should be a noop when called on a stopped Service.
	 */
	Stop()
}

type Policy int

const (
	ALWAYS = iota // 0 value and therefore the default for restart policies
	NEVER
	DIEALSO
)

// The description of a child service in a Supervision tree.
type ServiceSpec struct {
	// The service to run.
	service Service
	// The restart policy for restartPolicy
	//
	// The Policy should be one of the CONSTANTS:
	//	ALWAYS: always restart the service
	//	NEVER: Never restart the service
	//	DIEALSO: Supervisor should die as well
	restartPolicy Policy
	// The channel to use for querying the service
	// for health
	ping chan bool
	// TODO(jwall): service spec should have a count field as well?
}

/*
 Register a Service with the supervisor.

 Panics if the supervisor has already started.
 */
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

/*
 UnRegister a Service with the supervisor.

 Panics if the supervisor has already started.
 */
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
	s.service.Stop()
	ch, result := s.service.Start()
	s.ping = ch
	return result
}

func (sup *Supervisor) SetStarted(b bool) {
	sup.startedLock.Lock()
	sup.started = b
	sup.startedLock.Unlock()
}

/*
 Start a supervisor.

 Returns a chan bool and bool.

 	chan bool: the ping channel for querying supervisor for health
 	bool: true for when successful false otherwise
 */
func (sup *Supervisor) Start() (chan bool, bool) { // A supervisor is a service
	result := sup.doForServices(serviceStarter)
	sup.SetStarted(true)
	sup.SetStopSign(false)
	if !result {
		return nil, false
	}

	ping := make(chan bool, 1)

	//run supervisor loop
	go sup.Loop(ping)
	return ping, true
}

func (sup *Supervisor) Loop(ch chan bool) {
	defer close(ch)
	defer sup.SetStarted(false)

	for true {
		select {
		case ping := <-ch:
			if ping {
				ch <- true
			}
		default:
			// noop
		}

		result := sup.doForServices(func(s *ServiceSpec) bool {
			// check for service aliveness
			restart := false
			ch := s.ping
			log.Printf("the ping channel is: %s", s.ping)
			if ch == nil {
				restart = true
			} else {
				// first check to see if the channel is open
				select {
				case _, open := <-ch:
					if !open {
						restart = true
					}
				default:
					ch <- true // send a ping
					healthy, ok := <-ch // listen for response
					if !ok || (ok && !healthy) {
						restart = true
					}
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

// Stop a supervisor
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
