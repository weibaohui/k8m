// Package watcher 实现规则匹配器
package worker

import (
	"github.com/weibaohui/k8m/pkg/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/models"
)

// RuleMatcher 规则匹配器（支持按集群匹配）
type RuleMatcher struct {
	rules map[string]config.RuleConfig
}

// NewRuleMatcher 创建规则匹配器（按集群规则）
func NewRuleMatcher(rules map[string]config.RuleConfig) *RuleMatcher {
	return &RuleMatcher{rules: rules}
}

// getRule 返回当前事件所属集群的规则；若不存在则返回空规则（表示不过滤）
func (r *RuleMatcher) getRule(cluster string) config.RuleConfig {
	if r == nil || r.rules == nil {
		return config.RuleConfig{}
	}
	if rc, ok := r.rules[cluster]; ok {
		return rc
	}
	return config.RuleConfig{}
}

// Match 判断事件是否匹配对应集群规则
func (r *RuleMatcher) Match(event *models.K8sEvent) bool {
	rule := r.getRule(event.Cluster)
	if rule.IsEmpty() {
		return true
	}

	matched := false

	// 命名空间匹配
	if len(rule.Namespaces) > 0 {
		for _, ns := range rule.Namespaces {
			if event.Namespace == ns {
				matched = true
				break
			}
		}
		if !matched && !rule.Reverse {
			return false
		}
	}

	// 原因匹配 包括message
	if len(rule.Reasons) > 0 {
		reasonMatched := false
		for _, reason := range rule.Reasons {
			if event.Reason == reason {
				reasonMatched = true
				break
			}
		}
		if !reasonMatched && !rule.Reverse {
			return false
		}
		matched = matched || reasonMatched
	}

	// 类型匹配
	if len(rule.Types) > 0 {
		typeMatched := false
		for _, eventType := range rule.Types {
			if event.Type == eventType {
				typeMatched = true
				break
			}
		}
		if !typeMatched && !rule.Reverse {
			return false
		}
		matched = matched || typeMatched
	}

	// 未设置匹配条件时默认匹配
	if !matched && len(rule.Namespaces) == 0 && len(rule.Reasons) == 0 && len(rule.Types) == 0 {
		matched = true
	}

	if rule.Reverse {
		return !matched
	}
	return matched
}

// UpdateRules 更新整套规则
func (r *RuleMatcher) UpdateRules(rules map[string]config.RuleConfig) { r.rules = rules }
