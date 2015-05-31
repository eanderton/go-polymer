/*
The MIT License (MIT)

Copyright (c) 2015 Eric Anderton
*/
package polymer

import (
  "reflect"
  "strings"
)

/*
polymer:"<alias>"
polymer:"-"
polymer:",alias:<alias>"
polymer:",ignore"
polymer:",onchange:<fn>"
*/

type tagInfo struct {
  Name string
  Alias string
  Handler string
}

var (
  // Registry for all types to be ignored when processing Polymer tags.
  // Any type in this array will not participate in Polymer field handling,
  // will be regarded as though `polymer:"-"` was used on the field.
  IgnoredTagTypes = []reflect.Type {
    reflect.TypeOf(BasicComponent {}),
    reflect.TypeOf(PolymerBase {}),
    reflect.TypeOf(UpdateableAdapter {}),
    reflect.TypeOf(make(UpdateChan)),
  }
)

func isIgnoredType(field reflect.StructField) bool {
  for _, typ := range IgnoredTagTypes {
   if typ == field.Type {
      return true
   }
  }
  return false
}

// provide some way to turn a field 'off'
func newTagInfo(field reflect.StructField) (*tagInfo, bool) {
  if isIgnoredType(field) {
    return nil, false
  }

  tag := field.Tag.Get("polymer")
  if tag == "-" {
    return nil, false  // explicitly not mapped
  }

  // tag defaults
  ti := &tagInfo {
    Name: field.Name,
    Alias: field.Name,
    Handler: field.Name + "Changed",
  }

  // split tag parts out
  parts := strings.Split(strings.TrimSpace(tag), ",")
  if len(parts) > 0 {
    // first property is the alias name iff it doesn't look like a pair
    startIndex := 0
    alias := parts[0]
    if alias != "" && strings.Index(alias, ":") == -1 {
      ti.Alias = alias
      startIndex = 1
    }

    // iterate over remaining parts
    for ii := startIndex; ii < len(parts); ii++ {
      var name, val string

      // break things down into <name> or <name>:<val> pairs
      args := strings.Split(parts[ii], ":")
      switch len(args) {
        case 0:
          continue  // nothing to do, keep going
        case 1:
          name = strings.TrimSpace(args[0])
        default:
          name = strings.TrimSpace(args[0])
          val = strings.TrimSpace(args[1])
      }

      // handle supported attributes
      switch name {
        case "ignore":
          return nil, false  // explicitly not mapped
        case "alias":
          ti.Alias = val
        case "onchange":
          ti.Handler = val
      }
    }
  }

  // return completed tag info
  return ti, true
}


