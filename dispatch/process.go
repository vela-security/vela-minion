package dispatch

import (
	"errors"
	"fmt"
	"github.com/vela-security/vela-public/assert"
	"reflect"

	"github.com/vela-security/vela-minion/tunnel"
)

type process struct {
	fn  reflect.Value // 执行的方法
	arg bool          // 是否带有参数
	ptr bool          // 参数是否是指针类型
	in  reflect.Value // 入参反射类型
}

func (p process) execute(cli *tunnel.Client, msg *tunnel.Receive) error {
	rc := reflect.ValueOf(cli)
	var args []reflect.Value
	if !p.arg { // 没有其它参数的情况
		args = []reflect.Value{rc}
	} else {
		pin := p.in.Interface()
		if err := msg.Bind(pin); err != nil {
			return err
		}
		rin := reflect.ValueOf(pin)
		if !p.ptr {
			rin = rin.Elem()
		}
		args = []reflect.Value{rc, rin}
	}

	defer func() { _ = recover() }()

	ret := p.fn.Call(args)[0]
	if ret.IsNil() {
		return nil
	}

	return ret.Interface().(error)
}

var cliType = reflect.TypeOf(new(tunnel.Client))
var errorType = reflect.TypeOf((*error)(nil)).Elem()

func (d *dispatch) register(opcode assert.Opcode, fn any) error {
	if fn == nil {
		return errors.New("方法不能为空")
	}
	rvf := reflect.ValueOf(fn)
	if rvf.Kind() != reflect.Func {
		return fmt.Errorf("必须是%s类型", reflect.Func)
	}
	rtf := reflect.TypeOf(fn)
	nin, nou := rtf.NumIn(), rtf.NumOut()
	if nin != 1 && nin != 2 {
		return errors.New("方法参数数量不对")
	}
	if nou != 1 {
		return errors.New("返回值参数数量不对")
	}
	if rtf.In(0) != cliType {
		return fmt.Errorf("第一个参数必须是%s类型", cliType)
	}
	if rtf.Out(0) != errorType {
		return fmt.Errorf("方法返回值必须是%s类型", errorType)
	}
	proc := &process{fn: rvf}
	if nin == 2 {
		in, ptr := rtf.In(1), false
		if ptr = in.Kind() == reflect.Ptr; ptr {
			in = in.Elem()
		}
		proc.in, proc.arg, proc.ptr = reflect.New(in), true, ptr
	}

	d.pmu.Lock()
	defer d.pmu.Unlock()

	d.processes[opcode] = proc

	return nil
}
