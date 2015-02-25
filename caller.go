package socket

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type decoderFunc func([]byte, interface{}) error

type caller struct {
	Func    reflect.Value
	Args    []reflect.Type
	Decoder decoderFunc
}

func newCaller(f interface{}, dec decoderFunc) (*caller, error) {
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
		Func:    fv,
		Args:    args,
		Decoder: dec,
	}, nil
}

func (c *caller) GetArgs() []interface{} {
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

func (c *caller) decodeArgs(rargs []json.RawMessage, args []interface{}) {
	for i, _ := range args {
		if err := c.Decoder(rargs[i], &args[i]); err != nil {
			return
		}
	}
}

func (c *caller) Call(rargs []json.RawMessage) []reflect.Value {
	args := c.GetArgs()
	c.decodeArgs(rargs, args)
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

	return c.Func.Call(a)
}
