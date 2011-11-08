package supervisor

import "testing"
import "time"

type FakeService struct {}

func (f FakeService) Start() (chan bool, bool) {
	ch := make(chan bool)
	return ch, true
}

func (f FakeService) Stop() {
}

func TestServiceInterface(t *testing.T) {
	spec := ServiceSpec{service: FakeService{},
	                    restartPolicy: NEVER}
	_ = spec.service.(Service)
}	

func helperRegisterServiceSpecTests(name string, sup *Supervisor, t *testing.T) *ServiceSpec {
	return helperRegisterServiceTests(name, sup, FakeService{}, t)
}

func helperRegisterServiceTests(name string, sup *Supervisor, s Service, t *testing.T) *ServiceSpec {
	preListSize := len(sup.serviceSpec)
	spec := ServiceSpec{service: s}
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
	sup := NewSupervisor()
	helperRegisterServiceSpecTests("foo", sup, t)
}

func TestUnregisterServiceSpec(t *testing.T) {
	sup := NewSupervisor()
	specName := "foo"
	helperRegisterServiceSpecTests(specName, sup, t)
	sup.UnregisterService(specName)
	list := sup.serviceSpec
	if _, ok := list[specName]; ok {
		t.Error("Failed to unregister spec -- spec still there")
	}
}

func TestRegisterServiceSpecOnStartedSupervisor(t *testing.T) {
	sup := NewSupervisor()
	specName := "foo"
	sup.Start()
	defer func() {
		if x := recover(); x == nil {
			t.Error("Failed to panic when registering after started")
		}
	}()
	helperRegisterServiceSpecTests(specName, sup, t)
	sup.Stop()
}

func TestPingChannelnilOrClosed(t *testing.T) {
	sup := NewSupervisor()
	spec := ServiceSpec{service: FakeService{}}
	spec.restartPolicy = DIEALSO
	sup.RegisterService("foo", &spec)
	sup.Start()
	defer sup.Stop()

	if closed(spec.ping) {
		t.Error("service was not started")
	}
	spec.ping = nil
	time.Sleep(1e9)
	if !sup.stopSign || sup.started {
		t.Error("Supervisor did not die when channel was nil")
	}
	sup.Start()
	if closed(spec.ping) {
		t.Error("service was not started")
	}
	close(spec.ping)
	time.Sleep(1e9)
	if !sup.stopSign || sup.started {
		t.Error("Supervisor did not die when channel was closed")
	}
}

func TestPingChannelReportsUnhealthy(t *testing.T) {
	sup := NewSupervisor()
	spec := ServiceSpec{service: FakeService{}}
	spec.restartPolicy = DIEALSO
	sup.RegisterService("foo", &spec)
	sup.Start()
	defer sup.Stop()
	
	go func() {
		_ = <-spec.ping // listen for ping request
		spec.ping <- false // send an unhealthy response
	}()
	if closed(spec.ping) {
		t.Error("service was not started")
	}
	time.Sleep(1e9)
	if !sup.stopSign || sup.started {
		t.Error("Supervisor did not die when channel reported unhealthy")
	}
}

func TestSupervisorPings(t *testing.T) {
	sup := NewSupervisor()
	helperRegisterServiceSpecTests("foo", sup, t)
	ch, _ := sup.Start()
	defer sup.Stop()
	ch <- true
	ok := <-ch
	if !ok {
		t.Error("Supervisor did not respond to ping")
	}
}
