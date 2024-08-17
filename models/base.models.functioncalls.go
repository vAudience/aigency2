package models

import "encoding/json"

/*
	https://platform.openai.com/docs/guides/function-calling
	example JSON function definition:
	{
		"type": "function",
		"function": {
			"name": "get_current_weather",
			"description": "Get the current weather in a given location",
			"parameters": {
				"type": "object",
				"properties": {
					"location": {
						"type": "string",
						"description": "The city and state, e.g. San Francisco, CA",
					},
					"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]},
				},
				"required": ["location"],
			},
		}
	}
*/

type OpenaiFunction struct {
	Type     string                   `json:"type"`
	Function OpenaiFunctionDefinition `json:"function"`
}

func (of OpenaiFunction) ExportJsonString() string {
	jsonBytes, _ := json.Marshal(of)
	return string(jsonBytes)
}

type OpenaiFunctionDefinition struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Parameters  OpenaiFunctionParameters `json:"parameters"`
}

type OpenaiFunctionParameters struct {
	Type       string                             `json:"type"`
	Properties map[string]OpenaiFunctionParameter `json:"properties"`
	Required   []string                           `json:"required"`
}

type OpenaiFunctionParameter struct {
	Type        string                         `json:"type"`
	Description string                         `json:"description"`
	Items       *OpenaiParameterTypeArrayItems `json:"items,omitempty"`
	Enum        []string                       `json:"enum,omitempty"`
}

type OpenaiParameterTypeArrayItems struct {
	Type string `json:"type,omitempty"`
}

func NewOpenAIFunctionDefinition(functionName string, functionDescription string, parameters map[string]OpenaiFunctionParameter, requireds []string) (openaiFunc OpenaiFunction) {
	// ensure that for every parameter enum is not null, but an empty array
	for _, parameter := range parameters {
		if parameter.Enum == nil {
			parameter.Enum = []string{}
		}
	}
	openaiFunction := OpenaiFunction{
		Type: "function",
		Function: OpenaiFunctionDefinition{
			Name:        functionName,
			Description: functionDescription,
			Parameters: OpenaiFunctionParameters{
				Type:       "object",
				Properties: parameters,
				Required:   requireds,
			},
		},
	}
	return openaiFunction
}
