/*
The MIT License (MIT)

Copyright (c) 2015 Eric Anderton
*/
package polymer

import (
  "github.com/gopherjs/gopherjs/js"
)

type Component interface {
  InitComponent(this *js.Object)
}

type UpdateChan chan struct{}

type Updateable interface {
  RegisterComponent(update UpdateChan)
  UpdateComponent()
}

type LifecycleListener interface {
  Created()
  Ready()
  Attached()
  DomReady()
  Detached()
}

type PropertyListener interface {
  PropertyChanged(fieldName string, oldVal, newVal interface{})
}

type AttributeListener interface {
  AttributeChanged(attrName string, oldVal, newVal interface{})
}
