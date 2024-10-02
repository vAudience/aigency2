package models

import (
	"fmt"
	"strings"
)

type AdapterToolExecutionState string

const (
	AdapterToolExecutionState_Queued    AdapterToolExecutionState = "Queued"
	AdapterToolExecutionState_Running   AdapterToolExecutionState = "Running"
	AdapterToolExecutionState_Completed AdapterToolExecutionState = "Completed"
	AdapterToolExecutionState_Cancelled AdapterToolExecutionState = "Cancelled"
	AdapterToolExecutionState_Failed    AdapterToolExecutionState = "Failed"
	AdapterToolExecutionState_Unknown   AdapterToolExecutionState = "Unknown"
)

type AIgentAdapter interface {
	GetName() string
	GetDescription() string
	GetOpenAIFunctionParameters() (parameters OpenaiFunctionParameters)
	ValidateArguments(jobData AdapterExecutionData) (valid bool, err error)
	ExportOpenAIFunctionDefinition() (functionName string, openaiDefinition string)
	Execute(jobData AdapterExecutionData) (jobResults JobResults)
}

type AdapterExecutionData struct {
	AdapterName string         `json:"adapterName"`
	JobId       string         `json:"jobId"` // is callId for openai
	MissionId   string         `json:"missionId"`
	ThreadId    string         `json:"threadId"`
	RunId       string         `json:"runId"`
	Arguments   map[string]any `json:"arguments"`
	// AsyncCallback OnToolJobFinishedCallback `json:"-"`
}

func (aed *AdapterExecutionData) GetArgument(key string) (val any, ok bool) {
	val, ok = aed.Arguments[key]
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsString(key string) (val string, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return "", false
	}
	val, ok = content.(string)
	if !ok {
		return "", false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsInt(key string) (val int, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return 0, false
	}
	val, ok = content.(int)
	if !ok {
		return 0, false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsFloat(key string) (val float64, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return 0, false
	}
	val, ok = content.(float64)
	if !ok {
		return 0, false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsBool(key string) (val bool, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return false, false
	}
	val, ok = content.(bool)
	if !ok {
		return false, false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsArrayString(key string) (val []string, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return []string{}, false
	}
	val, ok = content.([]string)
	if !ok {
		return []string{}, false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsArrayInt(key string) (val []int, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return []int{}, false
	}
	val, ok = content.([]int)
	if !ok {
		return []int{}, false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsArrayFloat(key string) (val []float64, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return []float64{}, false
	}
	val, ok = content.([]float64)
	if !ok {
		return []float64{}, false
	}
	return val, ok
}

func (aed *AdapterExecutionData) GetArgumentValueAsArrayBool(key string) (val []bool, ok bool) {
	content, found := aed.Arguments[key]
	if !found {
		return []bool{}, false
	}
	val, ok = content.([]bool)
	if !ok {
		return []bool{}, false
	}
	return val, ok
}

func ValidateToolArguments(tool *NatsTool, jobData AdapterExecutionData) (updatedJobData AdapterExecutionData, valid bool, err error) {
	// check for existence of an Argument in jobData that has the key of param.Name or any of param.Aliases (check all case insensitive).
	// if not found, and required -> return false, error . if not found and not required -> set to default value
	// if found, check if it is of the correct type, if not -> return false, error
	// if found and enum not empty, check if it is in the enum, if not -> return false, error
	// if all checks pass, return true, nil
	filteredValidJobDataArguments := make(map[string]any)
	for _, param := range tool.Parameters {
		allowedLowercaseAliases := []string{param.Name}
		for _, alias := range param.Aliases {
			allowedLowercaseAliases = append(allowedLowercaseAliases, strings.ToLower(alias))
		}
		// copy aliases to the "real" key
		foundParam := false
		for _, alias := range allowedLowercaseAliases {
			if _, ok := jobData.Arguments[alias]; ok {
				jobData.Arguments[param.Name] = jobData.Arguments[alias]
				foundParam = true
				break
			}
		}
		if param.Required && !foundParam {
			return jobData, false, fmt.Errorf("required argument(%s) is missing", param.Name)
		}
		validVal, err := param.SetValue(jobData.Arguments[param.Name])
		if err != nil || validVal == nil {
			if param.Required {
				return jobData, false, fmt.Errorf("required argument(%s) is invalid", param.Name)
			}
			// we ignore invalid values for non-required params
			continue
		}
		filteredValidJobDataArguments[param.Name] = validVal
	}
	jobData.Arguments = filteredValidJobDataArguments
	return jobData, true, nil
}
