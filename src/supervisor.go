package supervisor

type Supervisor struct {
	serviceSpec map [string] Service
}

type Service interface {
	// returns true when successfull
	// and a ping channel to check for life
	Start() (chan bool, bool)
	Stop() // return true when successfull
}

func (sup Supervisor) RegisterService(name string, s Service) {
	sup.serviceSpec[name] = s
}

func (sup Supervisor) UnregisterService(name string) bool {
	if sup.serviceSpec[name] == nil {
		return false
	}
	sup.serviceSpec[name] = nil
	return true
}

func (sup Supervisor) Start() bool {
	return sup.doForServices(func(s Service) bool {
	        _, result := s.Start() // TODO(jwall): store the channel
		return result
	}) 
}

func (sup Supervisor) Stop() bool {
	return sup.doForServices(func(s Service) bool {
		s.Stop()
		return true
	}) 
}

func (sup Supervisor) doForServices(f func (s Service) bool) bool {
	result := true
	for _, s := range sup.serviceSpec {
		result = result && f(s)
	}
	return result
}
