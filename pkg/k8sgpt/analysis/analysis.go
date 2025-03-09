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
	"strings"
	"sync"
	"time"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/weibaohui/k8m/pkg/ai"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analyzer"
	"github.com/weibaohui/k8m/pkg/k8sgpt/common"
	"github.com/weibaohui/k8m/pkg/k8sgpt/util"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

type Analysis struct {
	Context        context.Context
	ClusterID      string                 // 集群ID
	AIClient       ai.IAI                 // AI
	Filters        []string               // 资源类型
	Namespace      string                 // 资源命名空间
	LabelSelector  string                 // k8s 获取资源的label selector
	Explain        bool                   // 是否使用AI进行解释
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

type JsonOutput struct {
	Errors   AnalysisErrors  `json:"errors"`   // 错误信息
	Status   AnalysisStatus  `json:"status"`   // 统计状态信息
	Problems int             `json:"problems"` // 错误统计数量
	Results  []common.Result `json:"results"`  // 分析统计结果
}

// Run 运行入口
func Run(cfg *Analysis) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("分析选项不能为空")
	}
	runner := cfg

	defer runner.Close()
	runner.RunAnalysis()
	if cfg.Explain {
		if err := runner.ExplainResultsByAI(true); err != nil {
			return nil, err
		}
	}

	output := "json"
	outputData, err := runner.PrintOutput(output)
	if err != nil {
		return nil, err
	}
	return outputData, nil
}

func (a *Analysis) RunAnalysis() {
	var activeFilters []string

	coreAnalyzerMap, analyzerMap := analyzer.GetAnalyzerMap()
	openapiSchema := &openapi_v2.Document{}
	if a.WithDoc {
		openapiSchema = kom.Cluster(a.ClusterID).Status().OpenAPISchema()
	}
	analyzerConfig := common.Analyzer{
		ClusterID:     a.ClusterID,
		Context:       a.Context,
		Namespace:     a.Namespace,
		LabelSelector: a.LabelSelector,
		AIClient:      a.AIClient,
		OpenapiSchema: openapiSchema,
	}

	semaphore := make(chan struct{}, a.MaxConcurrency)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	// if there are no filters selected and no active_filters then run coreAnalyzer
	if len(a.Filters) == 0 && len(activeFilters) == 0 {
		for name, item := range coreAnalyzerMap {
			wg.Add(1)
			semaphore <- struct{}{}
			go a.executeAnalyzer(item, name, analyzerConfig, semaphore, &wg, &mutex)

		}
		wg.Wait()
		return
	}
	// if the filters flag is specified
	if len(a.Filters) != 0 {
		for _, filter := range a.Filters {
			if item, ok := analyzerMap[filter]; ok {
				semaphore <- struct{}{}
				wg.Add(1)
				go a.executeAnalyzer(item, filter, analyzerConfig, semaphore, &wg, &mutex)
			} else {
				a.Errors = append(a.Errors, fmt.Sprintf("\"%s\" filter does not exist. Please run k8sgpt filters list.", filter))
			}
		}
		wg.Wait()
		return
	}

	// use active_filters
	for _, filter := range activeFilters {
		if item, ok := analyzerMap[filter]; ok {
			semaphore <- struct{}{}
			wg.Add(1)
			go a.executeAnalyzer(item, filter, analyzerConfig, semaphore, &wg, &mutex)
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

func (a *Analysis) ExplainResultsByAI(anonymize bool) error {
	if len(a.Results) == 0 {
		return nil
	}

	for index, analysis := range a.Results {
		var texts []string

		for _, failure := range analysis.Error {
			if anonymize {
				for _, s := range failure.Sensitive {
					failure.Text = util.ReplaceIfMatch(failure.Text, s.Unmasked, s.Masked)
				}
			}
			texts = append(texts, fmt.Sprintf("错误：%s\n,相关字段解释：%s", failure.Text, failure.KubernetesDoc))
		}

		promptTemplate := ai.PromptMap["default"]
		// If the resource `Kind` comes from an "integration plugin",
		// maybe a customized prompt template will be involved.
		if prompt, ok := ai.PromptMap[analysis.Kind]; ok {
			promptTemplate = prompt
		}
		result, err := a.getAIResultForSanitizedFailures(texts, promptTemplate)
		if err != nil {

			// Check for exhaustion.
			if strings.Contains(err.Error(), "status code: 429") {
				return fmt.Errorf("exhausted API quota for AI provider %s: %v", a.AIClient.GetName(), err)
			}
			return fmt.Errorf("failed while calling AI provider %s: %v", a.AIClient.GetName(), err)
		}

		if anonymize {
			for _, failure := range analysis.Error {
				for _, s := range failure.Sensitive {
					result = strings.ReplaceAll(result, s.Masked, s.Unmasked)
				}
			}
		}

		analysis.Details = result

		a.Results[index] = analysis
	}
	return nil
}

func (a *Analysis) getAIResultForSanitizedFailures(texts []string, promptTmpl string) (string, error) {
	inputKey := strings.Join(texts, " ")

	// Process template.
	prompt := fmt.Sprintf(strings.TrimSpace(promptTmpl), inputKey)
	klog.V(6).Infof("提示词: \n%s\n", prompt)
	response, err := a.AIClient.GetCompletion(a.Context, prompt)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (a *Analysis) Close() {
	if a.AIClient == nil {
		return
	}
	a.AIClient.Close()
}
