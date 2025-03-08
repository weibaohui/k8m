/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package analyzer

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/k8sgpt/common"
	"github.com/weibaohui/k8m/pkg/k8sgpt/util"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type NodeAnalyzer struct{}

func (NodeAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Node"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})
	var list []*v1.Node
	err := kom.Cluster(a.ClusterID).WithContext(a.Context).Resource(&v1.Node{}).WithLabelSelector(a.LabelSelector).List(&list).Error
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, node := range list {
		var failures []common.Failure
		for _, nodeCondition := range node.Status.Conditions {
			// https://kubernetes.io/docs/concepts/architecture/nodes/#condition
			switch nodeCondition.Type {
			case v1.NodeReady:
				if nodeCondition.Status == v1.ConditionTrue {
					break
				}
				failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
			// k3s `EtcdIsVoter`` should not be reported as an error
			case v1.NodeConditionType("EtcdIsVoter"):
				break
			default:
				if nodeCondition.Status != v1.ConditionFalse {
					failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[node.Name] = common.PreAnalysis{
				Node:           *node,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, node.Name, "").Set(float64(len(failures)))

		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := util.GetParent(a.Context, a.ClusterID, value.Node.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, err
}

func addNodeConditionFailure(failures []common.Failure, nodeName string, nodeCondition v1.NodeCondition) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("%s has condition of type %s, reason %s: %s", nodeName, nodeCondition.Type, nodeCondition.Reason, nodeCondition.Message),
		Sensitive: []common.Sensitive{
			{
				Unmasked: nodeName,
				Masked:   util.MaskString(nodeName),
			},
		},
	})
	return failures
}
