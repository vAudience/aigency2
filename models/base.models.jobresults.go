package models

import "fmt"

type JobResults struct {
	JobId       string                    `json:"jobId"`
	AdapterName string                    `json:"adapterName"`
	ResultTexts []string                  `json:"resultTexts"`
	ResultFiles []AdapterFileInfo         `json:"resultFiles"`
	FinalState  AdapterToolExecutionState `json:"finalState"`
	Err         error                     `json:"err"`
}

func NewJobResults(jobId string, adapterName string) *JobResults {
	entity := JobResults{}
	entity.JobId = jobId
	entity.AdapterName = adapterName
	entity.ResultTexts = make([]string, 0)
	entity.ResultFiles = make([]AdapterFileInfo, 0)
	entity.FinalState = AdapterToolExecutionState_Unknown
	return &entity
}

func (jr *JobResults) GetResultText(joinBy string) string {
	joined := ""
	for _, result := range jr.ResultTexts {
		joined += result + joinBy
	}
	return joined
}

func (jr *JobResults) GetResultFiles() []AdapterFileInfo {
	return jr.ResultFiles
}

func (jr *JobResults) AddResultText(text string) {
	jr.ResultTexts = append(jr.ResultTexts, text)
}

func (jr *JobResults) AddResultFile(file AdapterFileInfo) {
	jr.ResultFiles = append(jr.ResultFiles, file)
}

func (jr *JobResults) GetResultFilesText() string {
	joined := ""
	for n, result := range jr.ResultFiles {
		joined += fmt.Sprintf("%d. File Name: '%s'\n", n+1, result.FileName)
		joined += fmt.Sprintf("  - Mime Type: '%s'\n", result.MimeType)
		joined += fmt.Sprintf("  - Local File Path: '%s'\n", result.LocalPath)
		joined += fmt.Sprintf("  - Public Url: '%s'\n", result.PublicUrl)
	}
	return joined
}
