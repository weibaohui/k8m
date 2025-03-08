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

	appsv1 "k8s.io/api/core/v1"
)

type PvcAnalyzer struct{}

func (PvcAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "PersistentVolumeClaim"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	// search all namespaces for pods that are not running
	var list []*appsv1.PersistentVolumeClaim
	err := kom.Cluster(a.ClusterID).WithContext(a.Context).Resource(&appsv1.PersistentVolumeClaim{}).WithLabelSelector(a.LabelSelector).Namespace(a.Namespace).List(&list).Error
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, pvc := range list {
		var failures []common.Failure

		// Check for empty rs
		if pvc.Status.Phase == appsv1.ClaimPending {

			// parse the event log and append details
			evt, err := util.FetchLatestEvent(a.Context, a.ClusterID, pvc.Namespace, pvc.Name)
			if err != nil || evt == nil {
				continue
			}
			if evt.Reason == "ProvisioningFailed" && evt.Note != "" {
				failures = append(failures, common.Failure{
					Text:      evt.Note,
					Sensitive: []common.Sensitive{},
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)] = common.PreAnalysis{
				PersistentVolumeClaim: *pvc,
				FailureDetails:        failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, pvc.Name, pvc.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := util.GetParent(a.Context, a.ClusterID, value.PersistentVolumeClaim.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
