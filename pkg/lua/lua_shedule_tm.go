package lua

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
)

// CancellableJob 可取消的任务包装
type CancellableJob struct {
	ctx    context.Context
	cancel context.CancelFunc
	f      func(ctx context.Context)
}

func NewCancellableJob(f func(ctx context.Context)) *CancellableJob {
	ctx, cancel := context.WithCancel(context.Background())
	return &CancellableJob{ctx: ctx, cancel: cancel, f: f}
}

func (cj *CancellableJob) Run() {
	// 任务内部需周期性检查 ctx.Done() 或在长耗时操作时退出
	cj.f(cj.ctx)
}

func (cj *CancellableJob) Cancel() {
	cj.cancel()
}

// TaskManager 任务管理
type TaskManager struct {
	c     *cron.Cron
	mu    sync.Mutex
	tasks map[string]cron.EntryID
	jobs  map[string]*CancellableJob
}

func NewTaskManager() *TaskManager {
	c := cron.New(
		cron.WithChain(
			cron.Recover(cron.DefaultLogger),
			cron.SkipIfStillRunning(cron.DefaultLogger),
		),
	)
	return &TaskManager{
		c:     c,
		tasks: make(map[string]cron.EntryID),
		jobs:  make(map[string]*CancellableJob),
	}
}

func (tm *TaskManager) Start() {
	tm.c.Start()
}

func (tm *TaskManager) Stop() context.Context {
	return tm.c.Stop() // 返回一个 context，等待已在运行的 job 结束
}

// Add 新增任务，name 用于后续管理
func (tm *TaskManager) Add(name, spec string, jobFunc func(ctx context.Context)) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 如果已存在同名任务，先删除
	if id, ok := tm.tasks[name]; ok {
		tm.c.Remove(id)
		if j, ex := tm.jobs[name]; ex {
			j.Cancel()
			delete(tm.jobs, name)
		}
		delete(tm.tasks, name)
	}

	cjob := NewCancellableJob(jobFunc)
	id, err := tm.c.AddJob(spec, cjob)
	if err != nil {
		return err
	}
	tm.tasks[name] = id
	tm.jobs[name] = cjob
	return nil
}

// Remove 删除任务（不会中断正在运行的 job，但会 cancel 长任务）
func (tm *TaskManager) Remove(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if id, ok := tm.tasks[name]; ok {
		tm.c.Remove(id)
		delete(tm.tasks, name)
	}
	if j, ok := tm.jobs[name]; ok {
		j.Cancel() // 触发 job 内部通过 ctx 退出
		delete(tm.jobs, name)
	}
}

// Update = Remove + Add（立即生效），如果需要立刻触发一次可以在 Add 后主动 go jobFunc(ctx)
func (tm *TaskManager) Update(name, spec string, jobFunc func(ctx context.Context)) error {
	// 简单实现：Add 会自动移除旧的同名任务
	return tm.Add(name, spec, jobFunc)
}
