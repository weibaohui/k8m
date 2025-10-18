package lua

import (
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"k8s.io/klog/v2"
)

// TaskExecutionControl 任务执行控制结构体
// 用于管理定时任务的执行状态，防止已删除的任务继续执行
type TaskExecutionControl struct {
	RandomToken string       `json:"random_token"` // 随机字符串，用于在AddFunc中标识任务
	EntryID     cron.EntryID `json:"entry_id"`     // cron任务的EntryID
	IsDeleted   bool         `json:"is_deleted"`   // 是否已删除
	ScheduleID  uint         `json:"schedule_id"`  // 关联的巡检计划ID
	CreatedAt   time.Time    `json:"created_at"`   // 创建时间
}

// TaskControlManager 任务控制管理器
// 负责管理所有任务的执行控制状态
type TaskControlManager struct {
	mu                sync.RWMutex                           // 读写锁
	tokenToControl    map[string]*TaskExecutionControl       // 随机字符串 -> 控制结构
	entryToControl    map[cron.EntryID]*TaskExecutionControl // EntryID -> 控制结构
	scheduleToControl map[uint]*TaskExecutionControl         // ScheduleID -> 控制结构
}

// 全局任务控制管理器实例
var taskControlManager = &TaskControlManager{
	tokenToControl:    make(map[string]*TaskExecutionControl),
	entryToControl:    make(map[cron.EntryID]*TaskExecutionControl),
	scheduleToControl: make(map[uint]*TaskExecutionControl),
}
 
// RegisterTask 注册任务控制信息
// 在AddFunc执行前调用，建立随机字符串与EntryID的映射关系
func (tcm *TaskControlManager) RegisterTask(scheduleID uint, entryID cron.EntryID, token string) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	control := &TaskExecutionControl{
		RandomToken: token,
		EntryID:     entryID,
		IsDeleted:   false,
		ScheduleID:  scheduleID,
		CreatedAt:   time.Now(),
	}

	tcm.tokenToControl[token] = control
	tcm.entryToControl[entryID] = control
	tcm.scheduleToControl[scheduleID] = control

	klog.V(6).Infof("注册任务控制信息: scheduleID=%d, entryID=%d, token=%s", scheduleID, entryID, token)
}

// MarkTaskDeleted 标记任务为已删除状态
// 当任务被删除时调用，防止任务继续执行
func (tcm *TaskControlManager) MarkTaskDeleted(scheduleID uint) bool {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	if control, exists := tcm.scheduleToControl[scheduleID]; exists {
		control.IsDeleted = true
		klog.V(6).Infof("标记任务为已删除: scheduleID=%d, token=%s", scheduleID, control.RandomToken)
		return true
	}
	return false
}

// IsTaskDeleted 检查任务是否已被删除
// 在任务执行前调用，如果任务已被删除则不执行
func (tcm *TaskControlManager) IsTaskDeleted(token string) bool {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	if control, exists := tcm.tokenToControl[token]; exists {
		return control.IsDeleted
	}
	// 如果找不到控制信息，认为任务已被删除
	return true
}

// RemoveTask 完全移除任务控制信息
// 在任务被彻底清理时调用
func (tcm *TaskControlManager) RemoveTask(scheduleID uint) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	if control, exists := tcm.scheduleToControl[scheduleID]; exists {
		delete(tcm.tokenToControl, control.RandomToken)
		delete(tcm.entryToControl, control.EntryID)
		delete(tcm.scheduleToControl, scheduleID)
		klog.V(6).Infof("移除任务控制信息: scheduleID=%d, token=%s", scheduleID, control.RandomToken)
	}
}

// GetTaskByScheduleID 根据ScheduleID获取任务控制信息
func (tcm *TaskControlManager) GetTaskByScheduleID(scheduleID uint) (*TaskExecutionControl, bool) {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	control, exists := tcm.scheduleToControl[scheduleID]
	return control, exists
}

// GetTaskByEntryID 根据EntryID获取任务控制信息
func (tcm *TaskControlManager) GetTaskByEntryID(entryID cron.EntryID) (*TaskExecutionControl, bool) {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	control, exists := tcm.entryToControl[entryID]
	return control, exists
}

// ListActiveTasks 列出所有活跃的任务
func (tcm *TaskControlManager) ListActiveTasks() []*TaskExecutionControl {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	var activeTasks []*TaskExecutionControl
	for _, control := range tcm.scheduleToControl {
		if !control.IsDeleted {
			activeTasks = append(activeTasks, control)
		}
	}
	return activeTasks
}
