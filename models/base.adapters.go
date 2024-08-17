package models

import (
	"fmt"
	"strings"

	nuts "github.com/vaudience/go-nuts"
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
	for _, param := range tool.Parameters {
		allowedLowercaseAliases := []string{param.Name}
		for _, alias := range param.Aliases {
			allowedLowercaseAliases = append(allowedLowercaseAliases, strings.ToLower(alias))
		}
		foundParam := false
		for key, value := range jobData.Arguments {
			if foundParam {
				break
			}
			if !nuts.StringSliceContains(allowedLowercaseAliases, strings.ToLower(key)) {
				continue
			}
			foundParam = true
			// if found, check if it is of the correct type, if not -> return false, error
			// string, number, boolean, array:string, array:number, array:boolean are the only types allowed for now
			switch param.VarType {
			case "string":
				if _, ok := value.(string); !ok {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not a valid string", key)
					}
				}
			case "number":
				if _, ok := value.(float64); !ok {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not a valid number", key)
					}
				}
			case "boolean":
				if _, ok := value.(bool); !ok {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not a valid boolean", key)
					}
				}
			case "array:string":
				if _, ok := value.([]string); !ok {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not a valid array:string", key)
					}
				}
			case "array:number":
				if _, ok := value.([]float64); !ok {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not a valid array:number", key)
					}
				}
			case "array:boolean":
				if _, ok := value.([]bool); !ok {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not a valid array:bool", key)
					}
				}
			}
			// if found and enum not empty, check if it is in the enum, if not -> return false, error
			if len(param.Enum) > 0 && param.VarType == "string" {
				if !nuts.StringSliceContains(param.Enum, value.(string)) {
					if param.DefaultValue != nil {
						jobData.Arguments[key] = param.DefaultValue(jobData)
					} else {
						return jobData, false, fmt.Errorf("argument(%s) is not in the enum(%v)", key, param.Enum)
					}
				}
			}
			if !foundParam && param.Required {
				return jobData, false, fmt.Errorf("argument(%s) is required", key)
			} else if !foundParam && !param.Required && param.DefaultValue != nil {
				jobData.Arguments[key] = param.DefaultValue(jobData)
			}
			// ensure the original key is used/set/updated
			jobData.Arguments[param.Name] = jobData.Arguments[key]
		}
	}
	return jobData, true, nil
}
