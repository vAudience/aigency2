// base/base.models.aigencyrun.go
package base

import (
	"strings"

	nuts "github.com/vaudience/go-nuts"
)

const (
	IDPREFIX_AGENCYRUN = "arun"
	IDLENGTH_AGENCYRUN = 16
)

type AIgencyRun struct {
	ID                  string         `json:"id" validate:"required,min=1,max=64" setters:"system" filter:"admin,owner"`
	OwnerOrganizationID string         `json:"owner_organization_id" validate:"required,min=1,max=64" setters:"system" filter:"admin,owner"`
	OwnerUserID         string         `json:"owner_user_id" validate:"required,min=1,max=64" setters:"system" filter:"admin,owner"`
	MissionID           string         `json:"mission_id" validate:"required,min=1,max=64" setters:"system" filter:"admin,owner"`
	ChannelID           string         `json:"channel_id" validate:"required,min=1,max=64" setters:"system" filter:"admin,owner"`
	AgentID             string         `json:"agent_id" validate:"required,min=1,max=64" setters:"system" filter:"admin,owner"`
	Parameters          map[string]any `json:"parameters" setters:"system" filter:"admin,owner"`
	ToolJobIDs          []string       `json:"tool_job_ids" validate:"omitempty,dive" setters:"system" filter:"admin,owner"`
	TriggerMessageID    string         `json:"trigger_message_id" validate:"omitempty,min=1,max=64" setters:"system" filter:"admin,owner"`
	ResultingMessageID  string         `json:"resulting_message_id" validate:"omitempty,min=1,max=64" setters:"system" filter:"admin,owner"`
}

func NewAIgencyRun() *AIgencyRun {
	entity := AIgencyRun{
		ID:         CreateAIgencyRunID(),
		Parameters: make(map[string]any),
	}
	return &entity
}

func CreateAIgencyRunID() string {
	return nuts.NID(IDPREFIX_AGENCYRUN, IDLENGTH_AGENCYRUN)
}

func IsAIgencyRunID(id string) bool {
	return len(id) == IDLENGTH_AGENCYRUN+len(IDPREFIX_AGENCYRUN)+1 && strings.HasPrefix(id, IDPREFIX_AGENCYRUN)
}
