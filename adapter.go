/*
The MIT License (MIT)

Copyright (c) 2015 Eric Anderton
*/
package polymer

type LifecycleAdapter struct {}
func(c *LifecycleAdapter) Created() {}
func(c *LifecycleAdapter) Ready() {}
func(c *LifecycleAdapter) Attached() {}
func(c *LifecycleAdapter) DomReady() {}
func(c *LifecycleAdapter) Detached() {}

type UpdateableAdapter struct {
  update UpdateChan
}

func(a *UpdateableAdapter) RegisterComponent(update UpdateChan) {
  a.update = update
}

func(a *UpdateableAdapter) UpdateComponent() {
  a.update <- struct{}{}
}

type BasicComponent struct {
  PolymerBase
  UpdateableAdapter
}
