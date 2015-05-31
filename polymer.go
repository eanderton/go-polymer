/*
The MIT License (MIT)

Copyright (c) 2015 Eric Anderton
*/
package polymer

import (
  "github.com/gopherjs/gopherjs/js"
  "reflect"
  "strings"
)

// TODO: event api: https://www.polymer-project.org/1.0/docs/devguide/events.html

func makeCallback(fnName string, args... interface{}) *js.Object{
  return js.MakeFunc(
      func(this *js.Object, arguments []*js.Object) interface{} {
    finalArgs := []interface{} {}
    finalArgs = append(finalArgs, args...)
    for _, arg := range arguments {
      finalArgs = append(finalArgs, arg)
    }
    return this.Call(fnName, finalArgs...)
  })
}

func makeCallbackZero(fnName string) *js.Object{
  return js.MakeFunc(
      func(this *js.Object, arguments []*js.Object) interface{} {
    return this.Call(fnName)
  })
}

func Polymer(tagName string, defaultObj interface{}) {
  // build a prototype that bridges to a constructor and runtime-generated methods
  proto := js.M{}

  fields := map[string]*tagInfo {}
  handlers := []string {}
  exports := []string {}

  // get type metadata
  typ := reflect.TypeOf(defaultObj)
  oType := typ
  v := reflect.ValueOf(defaultObj)
  if typ.Kind() == reflect.Ptr{
    typ = typ.Elem()
    v = v.Elem()
  }

  // get field metadata
  for ii := 0; ii < typ.NumField(); ii++ {
    // get value and tag info
    field := typ.Field(ii)
    value := v.FieldByName(field.Name).Interface()
    taginfo, ok := newTagInfo(field)
    if !ok {
      continue
    }

    // set field lookups and default value
    fields[field.Name] = taginfo
    proto[taginfo.Alias] = value

    // set up property handlers
    if _, ok := oType.MethodByName(taginfo.Handler); ok {
      handlers = append(handlers, taginfo.Handler)
      proto[taginfo.Alias + "Changed"] = makeCallback("__" + taginfo.Handler, taginfo.Name)
    } else {
      proto[taginfo.Alias + "Changed"] = makeCallback("__propertyChanged", taginfo.Name)
    }
  }

  // set up function hooks
  for ii := 0; ii < oType.NumMethod(); ii++ {
    method := oType.Method(ii)
    if strings.HasPrefix(method.Name, "On") && method.Type.NumOut() <= 1 {
      exports = append(exports, method.Name)
      proto[method.Name] = makeCallback("__" + method.Name)
    }
  }

  // set up constructor
  proto["created"] = js.MakeFunc(
      func(this *js.Object, arguments []*js.Object) interface{} {
    // create a new instance
    obj := reflect.New(typ).Interface()

    // set defaults
    oType := reflect.ValueOf(obj)
    elem := oType.Elem()
    for _, taginfo := range fields {
      val := v.FieldByName(taginfo.Name)
      elem.FieldByName(taginfo.Name).Set(val)
    }

    // set specific property change handlers
    for _, methodName := range handlers {
      method := oType.MethodByName(methodName)
      func(methodName string) {
        this.Set("__" + methodName, func(fieldName string, oldValue, newValue interface{}) {
          newVal := reflect.ValueOf(newValue)
          oldVal := reflect.ValueOf(oldValue)
          elem.FieldByName(fieldName).Set(newVal)
          method.Call([]reflect.Value { oldVal, newVal })
        })
      }(methodName)
    }

    // set exported functions
    for _, fnName := range exports {
      func(fnName string) {
        argNum := oType.MethodByName(fnName).Type().NumIn()
        this.Set("__" + fnName, func(args... *js.Object) interface{} {
          values := []reflect.Value {}
          // copy all arguments up to the number supported
          for ii := 0; ii < argNum; ii++ {
            values = append(values, reflect.ValueOf(args[ii]))
          }
          return oType.MethodByName(fnName).Call(values)
        })
      }(fnName)
    }

    if pcl, ok := obj.(PropertyListener); ok {
      // set property change handler
      this.Set("__propertyChanged", func(fieldName string, oldValue, newValue interface{}) {
        elem.FieldByName(fieldName).Set(reflect.ValueOf(newValue))
        pcl.PropertyChanged(fieldName, oldValue, newValue)
      })
    } else {
      // set generic property change handler
      this.Set("__propertyChanged", func(fieldName string, oldValue, newValue interface{}) {
        elem.FieldByName(fieldName).Set(reflect.ValueOf(newValue))
      })
    }

    // AttributeListener hooks
    if acl, ok := obj.(AttributeListener); ok {
      this.Set("__attributeChanged", acl.AttributeChanged)
    }

    // LifecycleListener hooks
    if lcl, ok := obj.(LifecycleListener); ok {
      this.Set("__created", lcl.Created)
      this.Set("__ready", lcl.Ready)
      this.Set("__attached", lcl.Attached)
      this.Set("__domReady", lcl.DomReady)
      this.Set("__detached", lcl.Detached)
    }

    // Updateable hooks and init
    if uc, ok := obj.(Updateable); ok {
      // set update channel and goroutines
      update := make(UpdateChan)

      // listen for updates and move values to the component
      go func() {
        for {
          _, ok := <- update
          if !ok {
            break // channel is closed: end loop
          }
          // bounce values from Go obj (wc) to JS obj (this)
          for fieldName, taginfo := range fields {
            this.Set(taginfo.Alias, elem.FieldByName(fieldName).Interface())
          }
        }
      }()

      // 'create' the component with the update channel and the context
      uc.RegisterComponent(update)
    }

    // generic component hooks and init
    if c, ok := obj.(Component); ok {
      c.InitComponent(this)
    }

    return nil
  })

  // prototype callbacks for Lifecycle
  if _, ok := defaultObj.(LifecycleListener); ok {
    proto["ready"] = makeCallbackZero("__ready")
    proto["attached"] = makeCallbackZero("__attached")
    proto["domReady"] = makeCallbackZero("__domReady")
    proto["detached"] = makeCallbackZero("__detached")
  }

  // prototype callbacks for Attributes
  if _, ok := defaultObj.(AttributeListener); ok {
    proto["attributeChanged"] = makeCallback("__attributeChanged")
  }

  // the actual call to Polymer
  js.Global.Call("Polymer", tagName, proto)
}

