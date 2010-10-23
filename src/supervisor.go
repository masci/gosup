package supervisor

type Supervisor struct {
	serviceSpec map [string] Service
}

type Service interface {
	Start() bool // returns true when successfull
	Stop()  bool // return true when successfull
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
	        return s.Start() }) 
}

func (sup Supervisor) Stop() bool {
	return sup.doForServices(func(s Service) bool {
		return s.Stop() }) 
}

func (sup Supervisor) doForServices(f func (s Service) bool) bool {
	result := true
	for _, s := range sup.serviceSpec {
		result = result && f(s)
	}
	return result
}
