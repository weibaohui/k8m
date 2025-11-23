// Package watcher 实现规则匹配器
package watcher

import (
	"github.com/weibaohui/k8m/pkg/eventhandler/model"
)

// RuleMatcher 规则匹配器
type RuleMatcher struct {
	config *model.RuleConfig
}

// NewRuleMatcher 创建规则匹配器
func NewRuleMatcher(config *model.RuleConfig) *RuleMatcher {
	return &RuleMatcher{
		config: config,
	}
}

// Match 判断事件是否匹配规则
func (r *RuleMatcher) Match(event *model.Event) bool {
	// 如果规则配置为空，则匹配所有事件
	if r.config.IsEmpty() {
		return true
	}

	matched := false

	// 检查命名空间匹配
	if len(r.config.Namespaces) > 0 {
		for _, ns := range r.config.Namespaces {
			if event.Namespace == ns {
				matched = true
				break
			}
		}
		if !matched && !r.config.Reverse {
			return false
		}
	}

	// 检查原因匹配
	if len(r.config.Reasons) > 0 {
		reasonMatched := false
		for _, reason := range r.config.Reasons {
			if event.Reason == reason {
				reasonMatched = true
				break
			}
		}
		if !reasonMatched && !r.config.Reverse {
			return false
		}
		matched = matched || reasonMatched
	}

	// 检查类型匹配
	if len(r.config.Types) > 0 {
		typeMatched := false
		for _, eventType := range r.config.Types {
			if event.Type == eventType {
				typeMatched = true
				break
			}
		}
		if !typeMatched && !r.config.Reverse {
			return false
		}
		matched = matched || typeMatched
	}

	// 如果没有设置任何匹配条件，则根据反向开关决定
	if !matched && len(r.config.Namespaces) == 0 && len(r.config.Reasons) == 0 && len(r.config.Types) == 0 {
		matched = true
	}

	// 应用反向选择
	if r.config.Reverse {
		return !matched
	}

	return matched
}

// UpdateConfig 更新规则配置
func (r *RuleMatcher) UpdateConfig(config *model.RuleConfig) {
	r.config = config
}
