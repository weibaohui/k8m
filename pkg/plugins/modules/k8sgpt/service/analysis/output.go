package analysis

import (
	"encoding/json"
	"fmt"
	"time"
)

func (a *Analysis) jsonOutput() ([]byte, error) {
	var problems int
	var status AnalysisStatus
	for _, result := range a.Results {
		problems += len(result.Error)
	}
	if problems > 0 {
		status = StateProblemDetected
	} else {
		status = StateOK
	}

	result := ResultWithStatus{
		Problems: problems,
		Results:  a.Results,
		Errors:   a.Errors,
		Status:   status,
	}
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %v", err)
	}
	return output, nil
}
func (a *Analysis) ResultWithStatus() (ResultWithStatus, error) {
	var problems int
	var status AnalysisStatus
	for _, result := range a.Results {
		problems += len(result.Error)
	}
	if problems > 0 {
		status = StateProblemDetected
	} else {
		status = StateOK
	}

	result := ResultWithStatus{
		Problems:    problems,
		Results:     a.Results,
		Errors:      a.Errors,
		Status:      status,
		Stats:       a.Stats,
		LastRunTime: time.Now(),
	}

	return result, nil
}
