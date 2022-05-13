package dispatch

import (
	"errors"
	"fmt"
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/catch"
	"time"

	"github.com/vela-security/vela-public/safecall"
	"github.com/vela-security/vela-minion/model"
	"github.com/vela-security/vela-minion/tunnel"
)

type dataReq struct {
	Data any `json:"data"`
}

type substances []*substance

type substance struct {
	Name  string `json:"name"`
	Chunk []byte `json:"chunk"`
	Hash  string `json:"hash"`
}

func (s substance) startup() bool {
	return s.Name == "startup"
}

type taskResult struct {
	Removes []string     `json:"removes"` // 需要删除的配置名字
	Updates []*substance `json:"updates"` // 需要执行的配置
}

// empty 没有差异
func (t taskResult) empty() bool {
	return len(t.Removes) == 0 && len(t.Updates) == 0
}

// velaTask 配置管理器
type velaTask struct {
	xEnv assert.Environment
}

func (vt velaTask) reload(cli *tunnel.Client, req *substance) error {
	if err := vt.Execute(req); err != nil {
		return err
	}
	vt.sync(cli)
	return nil
}

// sync 同步配置
func (vt velaTask) sync(cli *tunnel.Client) {
	for {
		ret, err := vt.postTasks(cli)
		if err != nil {
			vt.xEnv.Warnf("上报 tasks 错误: %v", err)
		}
		if ret.empty() {
			break
		}

		vt.remove(ret.Removes)
		_ = vt.safeExecutes(ret.Updates)

		time.Sleep(time.Second)
	}
}

func (vt velaTask) safeExecutes(ss substances) (err error) {
	fn := func() error { return vt.executes(ss) }
	onTimeout := func() { err = errors.New("执行超时") }
	onPanic := func(cause any) { err = fmt.Errorf("执行发生了 panic: %v", cause) }
	onError := func(ex error) { err = fmt.Errorf("执行发生了错误: %s", ex) }

	safecall.New(!vt.xEnv.IsDebug()).Timeout(time.Minute).
		OnTimeout(onTimeout).
		OnPanic(onPanic).
		OnError(onError).
		Exec(fn)

	return
}

// executes 执行多个配置脚本
func (vt velaTask) executes(ss substances) error {
	size := len(ss)
	switch size {
	case 0:
		return nil
	case 1:
		return vt.Execute(ss[0])

	default:
		cch := catch.New()
		for i := 0; i < size; i++ {
			sub := ss[i]
			if sub.startup() {
				cch.Try(sub.Name, vt.Execute(sub))
				continue
			}
			vt.xEnv.Infof("register 配置: %s", sub.Name)
			cch.Try(sub.Name, vt.register(sub))

		}

		if e := cch.Wrap(); e != nil {
			return e
		}
	}

	return vt.wakeup()
}

// Execute 执行单个配置
func (vt velaTask) Execute(s *substance) error {
	if s == nil || s.Name == "" || len(s.Chunk) == 0 {
		return nil
	}
	vt.xEnv.Infof("执行配置: %s", s.Name)
	return vt.xEnv.DoTask(s.Name, s.Chunk, assert.TRANSPORT)
}

// register 注册配置
func (vt velaTask) register(s *substance) error {
	return vt.xEnv.RegisterTask(s.Name, s.Chunk, assert.TRANSPORT)
}

// wakeup 唤醒启动
func (vt velaTask) wakeup() error {
	return vt.xEnv.WakeupTask(assert.TRANSPORT)
}

// remove 删除配置
func (vt velaTask) remove(names []string) {
	for _, name := range names {
		vt.xEnv.Infof("执行配置: %s", name)
		_ = vt.xEnv.RemoveTask(name, assert.TRANSPORT)
	}
}

// postTasks 向中心端上报 tasks
func (vt velaTask) postTasks(cli *tunnel.Client) (taskResult, error) {
	var ret taskResult
	data := vt.tasks()
	req := &dataReq{Data: data}
	err := cli.PostJSON("/v1/task/sync", req, &ret)
	return ret, err
}

// tasks 获取所有任务配置
func (vt velaTask) tasks() model.VelaTasks {
	return vt.xEnv.TaskList()
}
