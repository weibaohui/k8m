package worker

import (
	"strings"

	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
)

// RuleMatcher 中文函数注释：事件规则匹配器。
type RuleMatcher struct {
	clusterRules map[string]config.RuleConfig
}

// NewRuleMatcher 中文函数注释：创建匹配器。
func NewRuleMatcher(clusterRules map[string]config.RuleConfig) *RuleMatcher {
	return &RuleMatcher{clusterRules: clusterRules}
}

// Match 中文函数注释：判断事件是否命中指定集群规则。
func (m *RuleMatcher) Match(event *models.K8sEvent) bool {
	rule, ok := m.clusterRules[event.Cluster]
	if !ok || rule.IsEmpty() {
		return true
	}
	matched := false
	nsOK := len(rule.Namespaces) == 0 || containsExact(rule.Namespaces, event.Namespace)
	nameOK := len(rule.Names) == 0 || containsPartial(rule.Names, event.Name)
	reasonOK := len(rule.Reasons) == 0 || containsPartial(rule.Reasons, event.Reason) || containsPartial(rule.Reasons, event.Message)
	matched = nsOK && nameOK && reasonOK
	if rule.Reverse {
		return !matched
	}
	return matched
}

func containsExact(list []string, v string) bool {
	for _, s := range list {
		if strings.TrimSpace(s) == v {
			return true
		}
	}
	return false
}

func containsPartial(list []string, v string) bool {
	for _, s := range list {
		ss := strings.TrimSpace(s)
		if ss != "" && strings.Contains(v, ss) {
			return true
		}
	}
	return false
}
