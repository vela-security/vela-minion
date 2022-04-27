package dispatch

import (
	"errors"
	"fmt"
	"github.com/vela-security/public/assert"
	"time"

	"github.com/vela-security/vela-minion/internal/safecall"
	"github.com/vela-security/vela-minion/model"
	"github.com/vela-security/vela-minion/rockcli"
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

// rockTask 配置管理器
type rockTask struct {
	xEnv assert.Environment
}

func (rt rockTask) reload(cli *rockcli.Client, req *substance) error {
	if err := rt.Execute(req); err != nil {
		return err
	}
	rt.sync(cli)
	return nil
}

// sync 同步配置
func (rt rockTask) sync(cli *rockcli.Client) {
	for {
		ret, err := rt.postTasks(cli)
		if err != nil {
			rt.xEnv.Warnf("上报 tasks 错误: %v", err)
		}
		if ret.empty() {
			break
		}

		rt.remove(ret.Removes)
		_ = rt.safeExecutes(ret.Updates)

		time.Sleep(time.Second)
	}
}

func (rt rockTask) safeExecutes(ss substances) (err error) {
	fn := func() error { return rt.executes(ss) }
	onTimeout := func() { err = errors.New("执行超时") }
	onPanic := func(cause any) { err = fmt.Errorf("执行发生了 panic: %v", cause) }
	onError := func(ex error) { err = fmt.Errorf("执行发生了错误: %s", ex) }

	safecall.New().Timeout(time.Minute).
		OnTimeout(onTimeout).
		OnPanic(onPanic).
		OnError(onError).
		Exec(fn)

	return
}

// executes 执行多个配置脚本
func (rt rockTask) executes(ss substances) error {
	size := len(ss)
	if size == 0 {
		return nil
	}
	if size == 1 {
		return rt.Execute(ss[0])
	}

	for _, sub := range ss {
		if sub.startup() {
			_ = rt.Execute(sub)
			continue
		}
		rt.xEnv.Infof("register 配置: %s", sub.Name)
		_ = rt.register(sub)
	}

	return rt.wakeup()
}

// Execute 执行单个配置
func (rt rockTask) Execute(s *substance) error {
	if s == nil || s.Name == "" || len(s.Chunk) == 0 {
		return nil
	}
	rt.xEnv.Infof("执行配置: %s", s.Name)
	return rt.xEnv.DoTask(s.Name, s.Chunk, assert.TRANSPORT)
}

// register 注册配置
func (rt rockTask) register(s *substance) error {
	return rt.xEnv.RegisterTask(s.Name, s.Chunk, assert.TRANSPORT)
}

// wakeup 唤醒启动
func (rt rockTask) wakeup() error {
	return rt.xEnv.WakeupTask(assert.TRANSPORT)
}

// remove 删除配置
func (rt rockTask) remove(names []string) {
	for _, name := range names {
		rt.xEnv.Infof("执行配置: %s", name)
		_ = rt.xEnv.RemoveTask(name, assert.TRANSPORT)
	}
}

// postTasks 向中心端上报 tasks
func (rt rockTask) postTasks(cli *rockcli.Client) (taskResult, error) {
	var ret taskResult
	data := rt.tasks()
	req := &dataReq{Data: data}
	err := cli.PostJSON("/v1/task/sync", req, &ret)
	return ret, err
}

// tasks 获取所有任务配置
func (rt rockTask) tasks() model.RockTasks {
	return rt.xEnv.TaskList()
}
