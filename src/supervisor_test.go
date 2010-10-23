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
	spec := ServiceSpec{service: FakeService{}}
	_ = spec.service.(Service)
}	

func TestRegisterServiceSpec(t *testing.T) {
	spec := ServiceSpec{service: FakeService{}}
	sup := newSupervisor()
	sup.RegisterService("foo", &spec)
	list := sup.serviceSpec
	if len(list) != 1 {
		t.Error("Failed to register spec -- list too short")
	}
	if spec2, ok := list["foo"]; !(ok && spec2 == &spec){
		t.Error("Failed to register spec -- spec not the same")
	}
}
