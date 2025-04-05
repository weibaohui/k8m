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

package analysis

import (
	"context"
	"fmt"
	"sync"
	"time"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analyzer"
	"github.com/weibaohui/k8m/pkg/k8sgpt/common"
	"github.com/weibaohui/kom/kom"
)

type Analysis struct {
	Context        context.Context
	ClusterID      string                 // 集群ID
	Filters        []string               // 资源类型
	Namespace      string                 // 资源命名空间
	LabelSelector  string                 // k8s 获取资源的label selector
	MaxConcurrency int                    // 资源并发
	WithDoc        bool                   // 是否携带字段解释
	WithStats      bool                   // 是否携带统计信息
	Stats          []common.AnalysisStats // 统计信息
	Results        []common.Result        // 分析结果
	Errors         []string               // 错误
}

type (
	AnalysisStatus string
	AnalysisErrors []string
)

const (
	StateOK              AnalysisStatus = "OK"
	StateProblemDetected AnalysisStatus = "ProblemDetected"
)

type ResultWithStatus struct {
	Errors      AnalysisErrors         `json:"errors,omitempty"`      // 错误信息
	Status      AnalysisStatus         `json:"status,omitempty"`      // 统计状态信息
	Problems    int                    `json:"problems,omitempty"`    // 错误统计数量
	Results     []common.Result        `json:"results,omitempty"`     // 分析统计结果
	Stats       []common.AnalysisStats `json:"stats,omitempty"`       // 运行状态统计
	LastRunTime time.Time              `json:"lastRunTime,omitempty"` // 运行时间
}

// Run 运行入口
func Run(cfg *Analysis) (*ResultWithStatus, error) {
	if cfg == nil {
		return nil, fmt.Errorf("分析选项不能为空")
	}
	runner := cfg
	runner.RunAnalysis()

	result, err := runner.ResultWithStatus()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *Analysis) RunAnalysis() {

	_, analyzerMap := analyzer.GetAnalyzerMap()
	openapiSchema := &openapi_v2.Document{}
	if a.WithDoc {
		openapiSchema = kom.Cluster(a.ClusterID).Status().OpenAPISchema()
	}
	analyzerConfig := common.Analyzer{
		ClusterID:     a.ClusterID,
		Context:       a.Context,
		Namespace:     a.Namespace,
		LabelSelector: a.LabelSelector,
		OpenapiSchema: openapiSchema,
	}

	semaphore := make(chan struct{}, a.MaxConcurrency)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// if the filters flag is specified
	for _, filter := range a.Filters {
		if item, ok := analyzerMap[filter]; ok {
			semaphore <- struct{}{}
			wg.Add(1)
			a.executeAnalyzer(item, filter, analyzerConfig, semaphore, &wg, &mutex)
		} else {
			a.Errors = append(a.Errors, fmt.Sprintf("\"%s\" filter does not exist. Please run k8sgpt filters list.", filter))
		}
	}
	wg.Wait()

}

func (a *Analysis) executeAnalyzer(analyzer common.IAnalyzer, filter string, analyzerConfig common.Analyzer, semaphore chan struct{}, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer wg.Done()

	var startTime time.Time
	var elapsedTime time.Duration

	// Start the timer
	if a.WithStats {
		startTime = time.Now()
	}

	// Run the analyzer
	results, err := analyzer.Analyze(analyzerConfig)

	// Measure the time taken
	if a.WithStats {
		elapsedTime = time.Since(startTime)
	}
	stat := common.AnalysisStats{
		Analyzer:     filter,
		DurationTime: elapsedTime,
	}

	mutex.Lock()
	defer mutex.Unlock()

	if err != nil {
		if a.WithStats {
			a.Stats = append(a.Stats, stat)
		}
		a.Errors = append(a.Errors, fmt.Sprintf("[%s] %s", filter, err))
	} else {
		if a.WithStats {
			a.Stats = append(a.Stats, stat)
		}
		a.Results = append(a.Results, results...)
	}
	<-semaphore
}
