package socket

import (
	"fmt"
	"reflect"
)

type caller struct {
	Func reflect.Value
	Args []reflect.Type
}

func newPacketHandler(f interface{}) (*caller, error) {
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		return nil, fmt.Errorf("f is not func")
	}
	ft := fv.Type()
	if ft.NumIn() == 0 {
		return &caller{
			Func: fv,
		}, nil
	}
	args := make([]reflect.Type, ft.NumIn())
	for i, n := 0, ft.NumIn(); i < n; i++ {
		args[i] = ft.In(i)
	}
	return &caller{
		Func: fv,
		Args: args,
	}, nil
}

func (c *caller) getArgs() []interface{} {
	ret := make([]interface{}, len(c.Args))
	for i, argT := range c.Args {
		if argT.Kind() == reflect.Ptr {
			argT = argT.Elem()
		}
		v := reflect.New(argT)
		ret[i] = v.Interface()
	}
	return ret
}

func (c *caller) call(args ...interface{}) {
	a := make([]reflect.Value, len(args))
	for i, arg := range args {
		v := reflect.ValueOf(arg)
		if c.Args[i].Kind() != reflect.Ptr {
			if v.IsValid() {
				v = v.Elem()
			} else {
				v = reflect.Zero(c.Args[i])
			}
		}
		a[i] = v
	}

	c.Func.Call(a)
}

func (c *caller) OnPacket(p Packet) {
	args := c.getArgs()
	p.DecodeArgs(args...)
	c.call(args...)
}
