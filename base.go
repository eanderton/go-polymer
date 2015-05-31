/*
The MIT License (MIT)

Copyright (c) 2015 Eric Anderton
*/
package polymer

import (
  "github.com/gopherjs/gopherjs/js"
  "fmt"
)

type PolymerBase struct {
  this *js.Object
}

func(p *PolymerBase) InitComponent(this *js.Object) {
  p.this = this
}

// Polymer utility functions
// https://www.polymer-project.org/1.0/docs/devguide/utility-functions.html

type QNode struct {
  *js.Object
}

func(p *PolymerBase) Q(name... string) *QNode {
  q := QNode{ p.this.Get("$") }
  return q.Q(name...)
}

func(q *QNode) Q(name... string) *QNode {
  ctx := q.Object
  for _, n := range name {
    if ctx == nil {
      break
    }
    ctx = ctx.Get(n)
  }
  return &QNode { ctx }
}

// TODO: use a DOM library to wrap *js.Object here?
type DomNode struct {
  *js.Object
}

func(p *PolymerBase) Query(selector string) DomNode {
  return DomNode{ p.this.Call("$$", selector) }
}

func(p *PolymerBase) ToggleClass(name string, state bool, node DomNode) {
  p.this.Call("toggleClass", name, state, node)
}

func(p *PolymerBase) AttributeFollows(name string, newNode, oldNode DomNode) {
  p.this.Call("attributeFollows", name, newNode, oldNode)
}

type EventOptions struct {
  Node *js.Object
  NoBubble bool
  Cancelable bool
}

func(p *PolymerBase) Fire(evtType string, detail interface{}, opts EventOptions) {
  p.this.Call("fire", evtType, detail, js.M {
    "node": opts.Node,
    "bubble": !opts.NoBubble,
    "cancelable": opts.Cancelable,
  })
}

func(p *PolymerBase) FireBasic(evtType string) {
  p.Fire(evtType, nil, EventOptions{})
}

type AsyncHandle struct {
  *js.Object
}

func(p *PolymerBase) Async(fn interface{}, waitMilliseconds uint) AsyncHandle {
  return AsyncHandle{ p.this.Call("async", fn, waitMilliseconds) }
}

func(p *PolymerBase) CancelAsync(handle AsyncHandle) {
  p.this.Call("cancelAsync", handle)
}

func(p *PolymerBase) Debounce(jobName string, fn interface{}, waitMilliseconds uint) {
  p.this.Call("debounce", jobName, fn, waitMilliseconds)
}

func(p *PolymerBase) CancelDebouncer(jobName string) {
  p.this.Call("cancelDebouncer", jobName)
}

func(p *PolymerBase) FlushDebouncer(jobName string) {
  p.this.Call("flushDebouncer", jobName)
}

func(p *PolymerBase) IsDebouncerActive(jobName string) bool {
  return p.this.Call("isDebouncerActive", jobName).Bool()
}

func(p *PolymerBase) Transform(expr string, node DomNode) {
  p.this.Call("transform", expr, node)
}

func(p *PolymerBase) Translate3d(x, y, z string, node DomNode) {
  p.this.Call("transform", x, y, z, node)
}

type ImportResult struct {
  Doc DomNode
  Err error
}

type ImportChan chan ImportResult

func(p *PolymerBase) ImportHref(url string) ImportChan {
  result := make(ImportChan)
  go func() {
    p.this.Call("importHref", func(e *js.Object) {
      result <- ImportResult { Doc: DomNode { e.Get("target").Get("import") }}
    }, func(e *js.Object) {
      result <- ImportResult { Err: fmt.Errorf("%s", e) }
    })
  }()
  return result
}

func(p *PolymerBase) OnMutation(node *DomNode, fn func()) {
  p.this.Call("onMutation", node, fn)
}
