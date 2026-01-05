package service

import (
	"sync/atomic"
)

type leaderService struct {
}

var leaderFlag int32

// IsCurrentLeader 中文函数注释：查询当前实例是否为主（Leader）。
func (s *leaderService) IsCurrentLeader() bool {
	return atomic.LoadInt32(&leaderFlag) == 1
}

// setCurrentLeader 中文函数注释：设置当前实例的主备状态。
func (s *leaderService) SetCurrentLeader(isLeader bool) {
	if isLeader {
		atomic.StoreInt32(&leaderFlag, 1)
	} else {
		atomic.StoreInt32(&leaderFlag, 0)
	}
}
