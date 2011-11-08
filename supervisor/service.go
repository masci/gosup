package supervisor

// The type of the fun that gets run in the service loop.
// Returns false to signal the loop it should end.
type LoopFun func() bool

// The type of a generic service capable of running in a supervision tree.
type GenericService struct {
	run LoopFun
	*startStopLock
}

// Constructs a new service using the LoopFun
func NewService(f LoopFun) *GenericService {
	return &GenericService{run:f, startStopLock:new(startStopLock)}
}

// Start a GenericService
func (s *GenericService) Start() (chan bool, bool) {
	s.setStarted(true)
	s.setStopSign(false)
	ch := make(chan bool, 1)
	go s.loop(ch, s.run)
	return ch, true
}

// Stop a GenericService.
func (s *GenericService) Stop() {
	s.setStopSign(true)
}

func (s *GenericService) loop(ch chan bool, f LoopFun) {
	defer close(ch)
	defer s.setStarted(false)
	for true {
		select {
		// handle ping requests if there is one
		case ping := <-ch:
			if ping {
				ch <- true
			}
		default:
			// noop
		}
		// obey the stop signs
		if s.stopSign {
			return
		}
		if !f() { // stop if the LoopFun returns false
			s.Stop()
		}
	}
}
