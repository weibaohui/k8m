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

	"github.com/weibaohui/k8m/pkg/ai"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analyzer"
	"github.com/weibaohui/k8m/pkg/k8sgpt/common"
	"github.com/weibaohui/k8m/pkg/k8sgpt/util"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
)

type Analysis struct {
	Context            context.Context
	Filters            []string
	Language           string
	AIClient           ai.IAI
	Results            []common.Result
	Errors             []string
	Namespace          string
	LabelSelector      string
	Explain            bool
	MaxConcurrency     int
	AnalysisAIProvider string // The name of the AI Provider used for this analysis
	WithDoc            bool
	WithStats          bool
	Stats              []common.AnalysisStats
	ClusterID          string
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
	Provider string          `json:"provider"`
	Errors   AnalysisErrors  `json:"errors"`
	Status   AnalysisStatus  `json:"status"`
	Problems int             `json:"problems"`
	Results  []common.Result `json:"results"`
}

func NewAnalysis(ctx context.Context, clusterID string, filters []string, namespace string, labelSelector string, explain bool, maxConcurrency int, withDoc bool, withStats bool) (*Analysis, error) {

	// Load remote cache if it is configured.

	a := &Analysis{
		Context:        ctx,
		ClusterID:      clusterID,
		Filters:        filters,
		Namespace:      namespace,
		LabelSelector:  labelSelector,
		Explain:        explain,
		MaxConcurrency: maxConcurrency,
		WithDoc:        withDoc,
		WithStats:      withStats,
	}
	if !explain {
		// Return early if AI use was not requested.
		return a, nil
	}

	client, err := service.AIService().DefaultClient()
	if err != nil {
		return nil, err
	}
	a.AIClient = client
	a.AnalysisAIProvider = "k8m ai"
	return a, nil
}

func (a *Analysis) RunAnalysis() {
	var activeFilters []string

	coreAnalyzerMap, analyzerMap := analyzer.GetAnalyzerMap()

	// we get the openapi schema from the server only if required by the flag "with-doc"
	if a.WithDoc {
		// TODO 找具体的定义文档
		// var openApiErr error
		kom.DefaultCluster().Status().Docs()
		// openapiSchema, openApiErr = a.Client.Client.Discovery().OpenAPISchema()
		// if openApiErr != nil {
		// 	a.Errors = append(a.Errors, fmt.Sprintf("[KubernetesDoc] %s", openApiErr))
		// }
	}

	analyzerConfig := common.Analyzer{
		ClusterID:     a.ClusterID,
		Context:       a.Context,
		Namespace:     a.Namespace,
		LabelSelector: a.LabelSelector,
		AIClient:      a.AIClient,
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

func (a *Analysis) GetAIResults(output string, anonymize bool) error {
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
			texts = append(texts, failure.Text)
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
	prompt := fmt.Sprintf(strings.TrimSpace(promptTmpl), a.Language, inputKey)
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
