package supervisor

import "testing"

type FakeService struct {}

func (f FakeService) Start() (chan bool, bool) {
	ch := make(chan bool)
	go f.Loop(ch)
	return ch, true
}

func (f FakeService) Loop(chan bool) {
	for true {
	}
}

func (f FakeService) Stop() {
}

func TestServiceInterface(t *testing.T) {
	spec := ServiceSpec{service: FakeService{},
	                    restartPolicy: NEVER}
	_ = spec.service.(Service)
}	

func helperRegisterServiceSpecTests(name string, sup *Supervisor, t *testing.T) *ServiceSpec {
	spec := ServiceSpec{service: FakeService{}}
	preListSize := len(sup.serviceSpec)
	sup.RegisterService("foo", &spec)
	list := sup.serviceSpec
	if len(list) <= preListSize {
		t.Error("Failed to register spec -- list too short")
	}
	if spec2, ok := list["foo"]; !(ok && spec2 == &spec){
		t.Error("Failed to register spec -- spec not the same")
	}
	return &spec
}

func TestRegisterServiceSpec(t *testing.T) {
	sup := newSupervisor()
	helperRegisterServiceSpecTests("foo", sup, t)
}

func TestUnregisterServiceSpec(t *testing.T) {
	sup := newSupervisor()
	specName := "foo"
	helperRegisterServiceSpecTests(specName, sup, t)
	sup.UnregisterService(specName)
	list := sup.serviceSpec
	if _, ok := list[specName]; ok {
		t.Error("Failed to unregister spec -- spec still there")
	}
}

func TestRegisterServiceSpecOnStartedSupervisor(t *testing.T) {
	sup := newSupervisor()
	specName := "foo"
	sup.started = true
	defer func() {
		if x := recover(); x == nil {
			t.Error("Failed to panic when registering after started")
		}
	}()
	helperRegisterServiceSpecTests(specName, sup, t)
}
