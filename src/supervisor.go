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

type Policy int

const (
	ALWAYS = iota
	NEVER
)

type ServiceSpec struct {
	service Service
	restartPolicy Policy
}

func (sup Supervisor) RegisterService(name string, s Service) {
	sup.serviceSpec[name] = s
}

func (sup Supervisor) UnregisterService(name string) bool {
	// check if the key exists
	if _, exists := sup.serviceSpec[name]; !exists {
		return false // return false if it didn't
	}
	sup.serviceSpec[name] = nil, false // delete the key
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
