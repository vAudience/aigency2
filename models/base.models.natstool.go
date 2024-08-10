package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	goopenai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	nuts "github.com/vaudience/go-nuts"
)

type OnToolJobFinishedCallback func(results JobResults)

type ToolExecutor func(tool *NatsTool, jobData AdapterExecutionData) (jobResults JobResults)

var ErrJobNotFound = errors.New("job not found")
var ErrNoExecutorForTool = errors.New("no executor set for tool")

var NATS_MANAGER_SERVER_URL string = "nats://localhost:4222"
var NATS_MANAGER_USERNAME string = "nats"
var NATS_MANAGER_PASSWORD string = "pw"
var AdapterBaseWorkdir string = "/aigency.aigent.studio/"
var AdapterBaseWebUrl string = "https://aigency.aigent.studio/"

var (
	NATS_TOPIC_TOOLS_ANNOUNCEMENTS string = "aigency.tools.announce"
	NATS_TOPIC_TOOLS_JOBS_NEW      string = "aigency.tools.jobs.new.{{tool.name}}"
	NATS_TOPIC_TOOLS_JOBS_STOP     string = "aigency.tools.jobs.stop.{{tool.name}}"
	NATS_TOPIC_TOOLS_JOBS_UPDATES  string = "aigency.tools.jobs.update"
)

func NewNatsToolJob() (emptyJob *NatsToolJob) {
	emptyJob = &NatsToolJob{
		JobID:          "",
		ToolName:       "",
		Parameters:     make(map[string]any),
		UpdatesChannel: make(chan *NatsToolJobUpdates),
		MissionId:      "",
		ThreadId:       "",
		RunId:          "",
		SubmittedAt:    time.Now(),
	}
	return emptyJob
}

func CreateToolJobFromExecutionData(executionData AdapterExecutionData) (job *NatsToolJob) {
	job = NewNatsToolJob()
	job.JobID = executionData.JobId
	job.ToolName = executionData.AdapterName
	job.Parameters = executionData.Arguments
	job.MissionId = executionData.MissionId
	job.MissionBaseUrl = path.Join(AdapterBaseWebUrl, executionData.MissionId)
	job.ThreadId = executionData.ThreadId
	job.RunId = executionData.RunId
	job.SubmittedAt = time.Now()
	return job
}

func NewToolManager() *NatsToolManager {
	// Example on how to start the tool manager and listen for announcements
	NATS_MANAGER_SERVER_URL = viper.GetString("NATS_SERVER_URL")
	NATS_MANAGER_USERNAME = viper.GetString("NATS_MANAGER_USERNAME")
	NATS_MANAGER_PASSWORD = viper.GetString("NATS_MANAGER_PASSWORD")
	toolManager, err := NewNatsToolManager(NATS_MANAGER_SERVER_URL, NATS_MANAGER_USERNAME, NATS_MANAGER_PASSWORD)
	if err != nil {
		nuts.L.Fatalf("[NewToolManager] Failed to create tool manager: %v", err)
	}
	toolManager.ListenForToolAnnouncements()
	toolManager.ListenForToolJobUpdates()
	// Add more logic here for toolCall distribution, updates/results handling, etc.
	return toolManager
}

type NatsTool struct {
	AIgentAdapter
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Type           string              `json:"type"`
	Parameters     []NatsToolParameter `json:"parameters"`
	ResponseFormat []NatsToolParameter `json:"response_format"`
	Version        string              `json:"version"`
	LastAnnounce   time.Time           `json:"-"`
	jobTopic       string              `json:"-"`
	executor       ToolExecutor        `json:"-"`
	natsClient     *nats.Conn          `json:"-"`
}

func (tool *NatsTool) GetName() string {
	return tool.Name
}

func (tool *NatsTool) GetDescription() string {
	return tool.Description
}

func (tool *NatsTool) GetOpenAIFunctionParameters() (parameters OpenaiFunctionParameters) {
	parameters = OpenaiFunctionParameters{
		Type:       "object",
		Properties: make(map[string]OpenaiFunctionParameter),
		Required:   []string{},
	}
	for _, param := range tool.Parameters {
		parameters.Properties[param.Name] = OpenaiFunctionParameter{
			Type:        param.VarType,
			Description: param.Description,
			Enum:        param.Enum,
		}
		if param.Required {
			parameters.Required = append(parameters.Required, param.Name)
		}
	}
	return parameters
}

func (tool *NatsTool) ExportOpenAIFunctionDefinition() (functionName string, openaiFunctionDefinition string) {
	openaiDef := OpenaiFunctionDefinition{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters:  tool.GetOpenAIFunctionParameters(),
	}
	openaiDefinitionBytes, err := json.Marshal(openaiDef)
	if err != nil {
		nuts.L.Errorf("failed to marshal openai function definition: %s", err)
		return "", ""
	}
	openaiFunctionDefinition = string(openaiDefinitionBytes)
	return tool.Name, openaiFunctionDefinition
}

func (tool *NatsTool) SetExecutor(newExecutor ToolExecutor) {
	tool.executor = newExecutor
}

func (tool *NatsTool) HasExecutor() bool {
	return tool.executor != nil
}

func (tool *NatsTool) Execute(data AdapterExecutionData) (jobResults JobResults) {
	var logName string = "[NatsTool.Execute] "
	if tool.executor == nil {
		jobResults.Err = ErrNoExecutorForTool
		return jobResults
	}
	jobResults = tool.executor(tool, data)
	// publish the results as a JobUpdate
	msg := fmt.Sprintf("Job(%s) for tool(%s) ended with status(%s)", data.JobId, tool.Name, jobResults.FinalState)
	if jobResults.Err != nil {
		msg += " and error: " + jobResults.Err.Error()
	}
	jobUpdate := NatsToolJobUpdates{
		JobID:          data.JobId,
		ToolName:       tool.Name,
		ToolVersion:    tool.Version,
		Status:         jobResults.FinalState,
		UpdateMsg:      msg,
		SubmittedAt:    time.Now(),
		UpdatedAt:      time.Now(),
		NewResultData:  jobResults.ResultTexts,
		NewResultFiles: jobResults.ResultFiles,
	}
	err := tool.natsClient.Publish(NATS_TOPIC_TOOLS_JOBS_UPDATES, tool.MarshalJobUpdate(jobUpdate))
	if err != nil {
		nuts.L.Errorf("%sfailed to publish job update: (%v)", logName, err)
	}
	return jobResults
}

func (tool *NatsTool) Announce() error {
	toolJsonBytes, err := json.Marshal(tool)
	if err != nil {
		return err
	}
	return tool.natsClient.Publish(NATS_TOPIC_TOOLS_ANNOUNCEMENTS, toolJsonBytes)
}

func (tool *NatsTool) ConnectToNATS(serverAddress string, username string, password string) error {
	nc, err := nats.Connect(serverAddress, nats.UserInfo(username, password))
	if err != nil {
		return err
	}
	tool.natsClient = nc
	tool.jobTopic = strings.ReplaceAll(NATS_TOPIC_TOOLS_JOBS_NEW, "{{tool.name}}", tool.Name)
	tool.ListenForNewJobs()
	nuts.Interval(func() bool {
		err := tool.Announce()
		if err != nil {
			nuts.L.Errorf("failed to announce tool: %v", err)
		}
		return true
	}, 60*time.Second, true)
	return nil
}

func (tool *NatsTool) CloseNATS() {
	if tool.natsClient == nil {
		return
	}
	tool.natsClient.Close()
}

func (tool *NatsTool) ListenForNewJobs() {
	var logName string = "[NatsTool.ListenForNewJobs] "
	if tool.natsClient == nil {
		nuts.L.Errorf("NATS client not set for tool: %s", tool.Name)
		return
	}
	tool.natsClient.Subscribe(tool.jobTopic, tool.NewJobHandler)
	nuts.L.Debugf("%sListening for new jobs on topic(%s)", logName, tool.jobTopic)
}

func (tool *NatsTool) NewJobHandler(jobMmsg *nats.Msg) {
	var logName string = "[NatsTool.NewJobHandler] "
	var job NatsToolJob
	var jobData AdapterExecutionData
	jobResults := *NewJobResults(job.JobID, tool.Name)
	err := json.Unmarshal(jobMmsg.Data, &job)
	if err != nil {
		nuts.L.Errorf("Error unmarshaling tool job: %v", err)
		jobResults.Err = fmt.Errorf("error unmarshaling tool job: %v", err)
		jobUpdate := NatsToolJobUpdates{
			JobID:          job.JobID,
			ToolName:       tool.Name,
			ToolVersion:    tool.Version,
			Status:         AdapterToolExecutionState_Failed,
			UpdateMsg:      fmt.Sprintf("error unmarshaling tool job: %v", err),
			SubmittedAt:    time.Now(),
			UpdatedAt:      time.Now(),
			NewResultData:  []string{},
			NewResultFiles: []AdapterFileInfo{},
		}
		// publish results via nats
		err = tool.natsClient.Publish(NATS_TOPIC_TOOLS_JOBS_UPDATES, tool.MarshalJobUpdate(jobUpdate))
		if err != nil {
			nuts.L.Errorf("failed to publish job results: %v", err)
		}
	} else if job.ToolName != tool.Name {
		jobResults.Err = fmt.Errorf("tool name mismatch: expected %s, got %s", tool.Name, job.ToolName)
		nuts.L.Debugf("%s?????????????? Tool name mismatch: expected %s, got %s", logName, tool.Name, job.ToolName)
		return
	} else {
		jobData = AdapterExecutionData{
			AdapterName: tool.Name,
			JobId:       job.JobID,
			MissionId:   job.MissionId,
			ThreadId:    job.ThreadId,
			RunId:       job.RunId,
			Arguments:   make(map[string]any),
		}
		for paramName, paramValue := range job.Parameters {
			jobData.Arguments[paramName] = paramValue
		}
		nuts.L.Debugf("%s :) ;) :-* Executing job(%s) with tool(%s)", logName, job.JobID, tool.Name)
		jobResults = tool.Execute(jobData)
	}
}

func (tool *NatsTool) MarshalResults(jobResults JobResults) (jobResultsJsonBytes []byte) {
	var err error
	jobResultsJsonBytes, err = json.Marshal(jobResults)
	if err != nil {
		nuts.L.Errorf("failed to marshal job results: %v", err)
		emptyMarhsallingErrorResults := JobResults{
			JobId:       jobResults.JobId,
			AdapterName: tool.Name,
			Err:         fmt.Errorf("failed to marshal job results: %v", err),
			FinalState:  AdapterToolExecutionState_Failed,
			ResultTexts: []string{},
			ResultFiles: []AdapterFileInfo{},
		}
		jobResultsJsonBytes, _ = json.Marshal(emptyMarhsallingErrorResults)
	}
	return jobResultsJsonBytes
}

func (tool *NatsTool) MarshalJobUpdate(jobUpdate NatsToolJobUpdates) (jobUpdateJsonBytes []byte) {
	var err error
	jobUpdateJsonBytes, err = json.Marshal(jobUpdate)
	if err != nil {
		nuts.L.Errorf("failed to marshal job update: %v", err)
		emptyMarhsallingErrorUpdate := NatsToolJobUpdates{
			JobID:          jobUpdate.JobID,
			ToolName:       tool.Name,
			ToolVersion:    tool.Version,
			Status:         AdapterToolExecutionState_Failed,
			UpdateMsg:      fmt.Sprintf("failed to marshal job update: %v", err),
			SubmittedAt:    time.Now(),
			UpdatedAt:      time.Now(),
			NewResultData:  []string{},
			NewResultFiles: []AdapterFileInfo{},
		}
		jobUpdateJsonBytes, _ = json.Marshal(emptyMarhsallingErrorUpdate)
	}
	return jobUpdateJsonBytes
}

// TODO: implement this
// func (tool *NatsTool) StopJobHandler(jobIDmsg *nats.Msg) {
// 	jobID := string(jobIDmsg.Data)
// 	job := tool.GetToolJob(jobID)
// 	if job == nil {
// 		nuts.L.Errorf("Job not found: %s", jobID)
// 		return
// 	}
// 	job.UpdateStatus(AdapterToolExecutionState_Stopped, "Job stopped by user", []string{}, []AdapterFileInfo{})
// }

// func (tool *NatsTool) ListenForStopJobs() {
// 	topic := strings.ReplaceAll(NATS_TOPIC_TOOLS_JOBS_STOP, "{{tool.name}}", tool.Name)
// 	tool.natsClient.Subscribe(topic, tool.StopJobHandler)
// }

// NatsToolParameter represents a parameter of a tool
type NatsToolParameter struct {
	Name         string                                  `json:"name"`
	Aliases      []string                                `json:"aliases"` // this is a list of aliases for the parameter-name since llms suck at being consistent
	Description  string                                  `json:"description"`
	VarType      string                                  `json:"var_type"`
	Required     bool                                    `json:"required"`
	Enum         []string                                `json:"enum"`
	DefaultValue func(data AdapterExecutionData) any     `json:"-"`
	Validate     func(value any) (valid bool, err error) `json:"-"` // this is a function to validate the parameter value
}

// NatsToolJob represents a job for a tool
type NatsToolJob struct {
	JobID          string                    `json:"job_id"` // for openai this is the CallId
	Status         AdapterToolExecutionState `json:"status"`
	StatusMessage  string                    `json:"status_message"`
	Updates        []NatsToolJobUpdates      `json:"updates"`
	ToolName       string                    `json:"tool_name"`
	ToolVersion    string                    `json:"tool_version"`
	Safety         sync.Mutex                `json:"-"`
	Parameters     map[string]any            `json:"parameters"`
	MissionId      string                    `json:"mission_id"`
	MissionBaseUrl string                    `json:"mission_base_url"`
	ThreadId       string                    `json:"thread_id"`
	RunId          string                    `json:"run_id"`
	SubmittedAt    time.Time                 `json:"submitted_at"`
	LatestUpdateAt time.Time                 `json:"latest_update_at"`
	ResultFiles    []AdapterFileInfo         `json:"created_files"`
	ResultData     []string                  `json:"result_data"`
	EndedAt        time.Time                 `json:"ended_at"`
	UpdatesChannel chan *NatsToolJobUpdates  `json:"-"`
}

func (job *NatsToolJob) AddResultFile(file AdapterFileInfo) {
	job.Safety.Lock()
	defer job.Safety.Unlock()
	job.ResultFiles = append(job.ResultFiles, file)
}

func (job *NatsToolJob) AddResultData(data string) {
	job.Safety.Lock()
	defer job.Safety.Unlock()
	job.ResultData = append(job.ResultData, data)
}

func (job *NatsToolJob) IsEnded() bool {
	job.Safety.Lock()
	defer job.Safety.Unlock()
	return job.Status == AdapterToolExecutionState_Completed || job.Status == AdapterToolExecutionState_Cancelled || job.Status == AdapterToolExecutionState_Failed
}

func (job *NatsToolJob) UpdateStatus(status AdapterToolExecutionState, msg string, newResultData []string, newResultFiles []AdapterFileInfo) {
	job.Safety.Lock()
	up := NatsToolJobUpdates{
		JobID:          job.JobID,
		ToolName:       job.ToolName,
		ToolVersion:    job.ToolVersion,
		Status:         status,
		UpdateMsg:      msg,
		SubmittedAt:    job.SubmittedAt,
		UpdatedAt:      time.Now(),
		NewResultData:  newResultData,
		NewResultFiles: newResultFiles,
	}
	job.Updates = append(job.Updates, up)
	job.Status = status
	job.LatestUpdateAt = up.UpdatedAt
	job.StatusMessage = msg
	if len(newResultData) > 0 {
		job.ResultData = append(job.ResultData, newResultData...)
	}
	if len(newResultFiles) > 0 {
		job.ResultFiles = append(job.ResultFiles, newResultFiles...)
	}
	if status == AdapterToolExecutionState_Completed || status == AdapterToolExecutionState_Cancelled || status == AdapterToolExecutionState_Failed {
		job.EndedAt = time.Now()
	}
	job.Safety.Unlock()
	job.UpdatesChannel <- &up
}

func (job *NatsToolJob) GetResults() (results JobResults) {
	results.JobId = job.JobID
	results.AdapterName = job.ToolName
	results.ResultTexts = job.ResultData
	results.ResultFiles = job.ResultFiles
	results.FinalState = job.Status
	return results
}

type NatsToolJobUpdates struct {
	JobID          string                    `json:"job_id"`
	ToolName       string                    `json:"tool_name"`
	ToolVersion    string                    `json:"tool_version"`
	Status         AdapterToolExecutionState `json:"status"`
	UpdateMsg      string                    `json:"update_msg"`
	SubmittedAt    time.Time                 `json:"submitted_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
	NewResultData  []string                  `json:"new_result_data"`
	NewResultFiles []AdapterFileInfo         `json:"new_result_files"`
}

// NatsToolManager manages tools and toolCalls
type NatsToolManager struct {
	safety            sync.Mutex
	tools             map[string]*NatsTool    // map of tools by tool name
	toolJobs          map[string]*NatsToolJob // map of toolCalls by job id
	natsClient        *nats.Conn
	toolPruneInterval nuts.GoInterval
	jobPruneInterval  nuts.GoInterval
}

func NewNatsToolManager(natsURL string, username string, password string) (*NatsToolManager, error) {
	nc, err := nats.Connect(natsURL, nats.UserInfo(username, password))
	if err != nil {
		return nil, err
	}
	newTM := NatsToolManager{
		tools:      make(map[string]*NatsTool),    // map of tools by tool name
		toolJobs:   make(map[string]*NatsToolJob), // map of toolCalls by job id
		natsClient: nc,
	}
	newTM.toolPruneInterval = *nuts.Interval(newTM.PruneExpiredTools, 30*time.Second, false)
	newTM.jobPruneInterval = *nuts.Interval(newTM.PruneExpiredJobs, 60*time.Second, false)
	return &newTM, nil
}

func (tm *NatsToolManager) PruneExpiredTools() bool {
	var logName string = "[NatsToolManager.PruneExpiredTools] "
	tm.safety.Lock()
	defer tm.safety.Unlock()
	for name, tool := range tm.tools {
		if time.Since(tool.LastAnnounce) > 61*time.Second {
			nuts.L.Infof("%sPruning tool(%s)", logName, name)
			delete(tm.tools, name)
		}
	}
	return true
}

func (tm *NatsToolManager) PruneExpiredJobs() bool {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	for jobID, job := range tm.toolJobs {
		if job.IsEnded() && time.Since(job.EndedAt) > 300*time.Second {
			delete(tm.toolJobs, jobID)
		}
	}
	return true
}

// ListenForToolAnnouncements listens for tool announcements on NATS and updates tool availability
func (tm *NatsToolManager) ListenForToolAnnouncements() {
	var logName string = "[NatsToolManager.ListenForToolAnnouncements] "
	tm.natsClient.Subscribe(NATS_TOPIC_TOOLS_ANNOUNCEMENTS, func(m *nats.Msg) {
		var tool NatsTool
		err := json.Unmarshal(m.Data, &tool)
		if err != nil {
			nuts.L.Debugf("%sError unmarshaling tool announcement: %v\n%s", logName, err, string(m.Data))
			return
		}
		tm.safety.Lock()
		tool.LastAnnounce = time.Now()
		oldTool, ok := tm.tools[tool.Name]
		if !ok || oldTool.Version != tool.Version {
			nuts.L.Debugf("%sNEW TOOL(%s) announced", logName, tool.Name)
		}
		tm.tools[tool.Name] = &tool
		defer tm.safety.Unlock()
	})
}

func (tm *NatsToolManager) GetTool(name string) *NatsTool {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	tool, ok := tm.tools[name]
	if !ok {
		return nil
	}
	return tool
}

func (tm *NatsToolManager) GetToolNames() []string {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	names := make([]string, 0, len(tm.tools))
	for name := range tm.tools {
		names = append(names, name)
	}
	return names
}

func (tm *NatsToolManager) HasTool(name string) bool {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	_, ok := tm.tools[name]
	return ok
}

func (tm *NatsToolManager) GetOpenAIFunctionDefinitions() (definitions []goopenai.FunctionDefinition) {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	for _, tool := range tm.tools {
		funcDef := goopenai.FunctionDefinition{}
		_, openaiFunctionDefinition := tool.ExportOpenAIFunctionDefinition()
		err := json.Unmarshal([]byte(openaiFunctionDefinition), &funcDef)
		if err != nil {
			nuts.L.Errorf("failed to unmarshal openai function definition: %s", err)
			continue
		}
		definitions = append(definitions, funcDef)
	}
	return definitions
}

func (tm *NatsToolManager) GetOpenAIToolsForLib(toolNames []string) (tools []goopenai.AssistantTool) {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	tools = []goopenai.AssistantTool{}
	if len(toolNames) > 0 {
		for _, tool := range tm.tools {
			// if we have toolNames, only include those
			if !nuts.StringSliceContains(toolNames, tool.GetName()) {
				continue
			}
			_, openaiToolDefinition := tool.ExportOpenAIFunctionDefinition()
			funcDef := goopenai.FunctionDefinition{}
			err := json.Unmarshal([]byte(openaiToolDefinition), &funcDef)
			if err != nil {
				nuts.L.Errorf("failed to unmarshal openai function definition: %s", err)
				continue
			}
			funcTool := goopenai.AssistantTool{
				Type:     goopenai.AssistantToolTypeFunction,
				Function: &funcDef,
			}
			tools = append(tools, funcTool)
		}
		if nuts.StringSliceContains(toolNames, string(goopenai.AssistantToolTypeRetrieval)) {
			tools = append(tools, goopenai.AssistantTool{
				Type: goopenai.AssistantToolTypeRetrieval,
			})
		}
		if nuts.StringSliceContains(toolNames, string(goopenai.AssistantToolTypeCodeInterpreter)) {
			tools = append(tools, goopenai.AssistantTool{
				Type: goopenai.AssistantToolTypeCodeInterpreter,
			})
		}
	}
	return tools
}

func (tm *NatsToolManager) WriteOpenAIFunctionDefinitionsToDisc() {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	jsonString := "["
	for _, adapter := range tm.tools {
		_, openaiDefinition := adapter.ExportOpenAIFunctionDefinition()
		jsonString += openaiDefinition + ","
	}
	jsonString = jsonString[:len(jsonString)-1] + "]"
	roodDir := viper.GetString("ROOT_DIR")
	filePath := path.Join(roodDir, "adapters_tools_openai_function_definitions.json")
	err := os.WriteFile(filePath, []byte(jsonString), 0644)
	if err != nil {
		nuts.L.Errorf("failed to write openai_function_definitions.json: %s", err)
	}
}

func (tm *NatsToolManager) GetToolJob(jobID string) *NatsToolJob {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	job, ok := tm.toolJobs[jobID]
	if !ok {
		return nil
	}
	return job
}

func (tm *NatsToolManager) ListenForToolJobUpdates() {
	var logName string = "[NatsToolManager.ListenForToolJobUpdates] "
	tm.natsClient.Subscribe(NATS_TOPIC_TOOLS_JOBS_UPDATES, func(m *nats.Msg) {
		var jobUpdate NatsToolJobUpdates
		err := json.Unmarshal(m.Data, &jobUpdate)
		if err != nil {
			nuts.L.Debugf("%sError unmarshaling tool job update: %v -_>\n%s", logName, err, string(m.Data))
			return
		}
		tm.safety.Lock()
		defer tm.safety.Unlock()
		job, ok := tm.toolJobs[jobUpdate.JobID]
		if !ok {
			nuts.L.Debugf("%s!?!?!??!?!?!? Job not found(%s) in update:\n%s", logName, jobUpdate.JobID, nuts.GetPrettyJson(jobUpdate))
			return
		}
		nuts.L.Debugf("%sJob(%s) updated with status(%s) and msg(%s)", logName, jobUpdate.JobID, jobUpdate.Status, jobUpdate.UpdateMsg)
		job.UpdateStatus(jobUpdate.Status, jobUpdate.UpdateMsg, jobUpdate.NewResultData, jobUpdate.NewResultFiles)
	})
}

// AddToolCall adds a new NatsToolJob to the manager
// Implement the logic to add NatsToolJob based on incoming requests
func (tm *NatsToolManager) AddToolJob(job *NatsToolJob) (err error) {
	tm.safety.Lock()
	tm.toolJobs[job.JobID] = job
	tm.safety.Unlock()
	//publish via nats
	jobJsonBytes, err := json.Marshal(job)
	if err != nil {
		nuts.L.Errorf("failed to marshal job: %v", err)
		return
	}
	topic := strings.ReplaceAll(NATS_TOPIC_TOOLS_JOBS_NEW, "{{tool.name}}", job.ToolName)
	err = tm.natsClient.Publish(topic, jobJsonBytes)
	if err != nil {
		nuts.L.Errorf("failed to publish job: %v", err)
	}
	return err
}

// CancelToolJob stops an ongoing NatsToolJob
// Implement logic to stop a NatsToolJob, likely by sending a message to the appropriate tool
func (tm *NatsToolManager) StopToolJob(jobID string) (err error) {
	// publish via nats
	topic := strings.ReplaceAll(NATS_TOPIC_TOOLS_JOBS_STOP, "{{tool.name}}", tm.toolJobs[jobID].ToolName)
	err = tm.natsClient.Publish(topic, []byte(jobID))
	if err != nil {
		nuts.L.Errorf("failed to publish job stop: %v", err)
	}
	return err
}

// GetToolJobStatus returns the latest status of a NatsToolJob
func (tm *NatsToolManager) GetToolJobStatus(jobID string) (AdapterToolExecutionState, error) {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	job, ok := tm.toolJobs[jobID]
	if !ok {
		return AdapterToolExecutionState_Unknown, ErrJobNotFound
	}
	return job.Status, nil
}

// ListAvailableTools lists all currently available tools and their metadata
func (tm *NatsToolManager) ListAvailableTools() []*NatsTool {
	tm.safety.Lock()
	defer tm.safety.Unlock()
	tools := make([]*NatsTool, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (tm *NatsToolManager) ExecuteJob(executionData AdapterExecutionData) (job *NatsToolJob, err error) {
	var logName string = "[NatsToolManager.ExecuteJob] "
	// tool := tm.GetTool(executionData.AdapterName)
	// if tool == nil {
	// 	results.Err = errors.New(fmt.Sprintf("Tool not found: %s", executionData.AdapterName))
	// 	return results
	// }
	// results = tool.Execute(executionData)
	job = CreateToolJobFromExecutionData(executionData)
	nuts.L.Infof("%s--==>> Job submitted with jobId(%s) runId(%s) thread(%s) mission(%s) for tool(%s)", logName, job.JobID, job.RunId, job.ThreadId, job.MissionId, job.ToolName)
	err = tm.AddToolJob(job)
	return job, err
}
