# goiface
Go tool for working with interface

Install
---
`go get github.com/vkd/goiface`

Tools
---
* implement interface - `goiface impl <receiver> <interface>`

Using
---
```
$ goiface impl github.com/vkd/goiface/testdata.MyType github.com/vkd/goiface/testdata.MyIface
func (m MyType) Iface() *MyType {
	panic("not implemented")
}

```
