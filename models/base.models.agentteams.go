package models

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

var (
	ErrInvalidAgentTeam = errors.New("invalid agent team data")
	// ErrUnauthorizedAccess = errors.New("unauthorized access")
)

const (
	IDPREFIX_AGENTTEAM = "agteam"
	IDLENGTH_AGENTTEAM = 16
)

type FullAgentTeam struct {
	AgentTeam *AgentTeam `json:"agent_team" readxs:"system,admin,owner"`
	Agents    []*Agent   `json:"agents" readxs:"system,admin,owner"`
}

type AgentTeamMessagesInjectionMode string //@name AgentTeamMessagesInjectionMode

const (
	AgentTeamMessagesInjectionModeAppend  AgentTeamMessagesInjectionMode = "append" // is default
	AgentTeamMessagesInjectionModePrepend AgentTeamMessagesInjectionMode = "prepend"
	AgentTeamMessagesInjectionModeReplace AgentTeamMessagesInjectionMode = "replace"
)

// this Dto uses pointers to check for null values assuming we get json in
type AgentTeamWriteDto struct {
	Name                    *string                         `json:"name" validate:"omitempty,min=1,max=255" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Description             *string                         `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Tags                    *[]string                       `json:"tags" validate:"omitempty,unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	AgentIDs                *[]string                       `json:"agent_ids" validate:"omitempty,unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	CoordinatingAgentID     *string                         `json:"coordinating_agent_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	SystemMessages          *[]string                       `json:"system_messages" validate:"omitempty,unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	SystemMessagesMode      *AgentTeamMessagesInjectionMode `json:"system_messages_mode" validate:"omitempty,oneof=append prepend replace" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	InitialUserMessages     *[]string                       `json:"initial_user_messages" validate:"unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	InitialUserMessagesMode AgentTeamMessagesInjectionMode  `json:"initial_user_messages_mode" validate:"oneof=append prepend replace" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	MetaData                *map[string]any                 `json:"meta_data" writexs:"system,admin,owner" readxs:"system,admin,owner"`
} //@name AgentTeamWriteDto

type AgentTeam struct {
	ID                      string                         `json:"id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	Name                    string                         `json:"name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Description             string                         `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Tags                    []string                       `json:"tags" validate:"unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	AgentIDs                []string                       `json:"agent_ids" validate:"unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	CoordinatingAgentID     string                         `json:"coordinating_agent_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	SystemMessages          []string                       `json:"system_messages" validate:"unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	SystemMessagesMode      AgentTeamMessagesInjectionMode `json:"system_messages_mode" validate:"oneof=append prepend replace" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	InitialUserMessages     []string                       `json:"initial_user_messages" validate:"unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	InitialUserMessagesMode AgentTeamMessagesInjectionMode `json:"initial_user_messages_mode" validate:"oneof=append prepend replace" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	MetaData                map[string]any                 `json:"meta_data" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	OwnerId                 string                         `json:"owner_id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	OwnerOrganizationId     string                         `json:"owner_organization_id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	CreatedAt               int64                          `json:"created_at" writexs:"system" readxs:"system,admin,owner"`
	UpdatedBy               string                         `json:"updated_by" validate:"omitempty,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	UpdatedAt               int64                          `json:"updated_at" writexs:"system" readxs:"system,admin,owner"`
} //@name AgentTeam

func NewAgentTeam() *AgentTeam {
	ID := CreateAgentTeamID()
	now := nuts.TimeToJSTimestamp(time.Now())
	entity := AgentTeam{
		ID:                      ID,
		AgentIDs:                make([]string, 0),
		Tags:                    make([]string, 0),
		SystemMessagesMode:      AgentTeamMessagesInjectionModeAppend,
		InitialUserMessagesMode: AgentTeamMessagesInjectionModeAppend,
		SystemMessages:          make([]string, 0),
		InitialUserMessages:     make([]string, 0),
		MetaData:                make(map[string]any),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	return &entity
}

func CreateAgentTeamID() string {
	return nuts.NID(IDPREFIX_AGENTTEAM, IDLENGTH_AGENTTEAM)
}

func IsAgentTeamID(id string) (isModelID bool) {
	isModelID = (len(id) == IDLENGTH_AGENTTEAM+len(IDPREFIX_AGENTTEAM)+1) && strings.HasPrefix(id, IDPREFIX_AGENTTEAM)
	// nuts.L.Debugf("(%s)IsAgentTeamID: %v --> length expected(%d) is(%d) prefix(%s)", id, isModelID, IDLENGTH_AGENTTEAM+len(IDPREFIX_AGENTTEAM)+1, len(id), IDPREFIX_AGENTTEAM)
	return isModelID
}

func ValidateAgentTeam(agentTeam *AgentTeam) error {
	validate := validator.New()

	err := validate.Struct(agentTeam)
	if err != nil {
		return ErrInvalidAgentTeam
	}

	return nil
}

// // TODO: this is a PLACEHOLDER implementation
// func (entity *AgentTeam) HasAccess(clientRole StructFieldWriteXSRole, userId string, orgId string, xsMode AccessMode) (xsError error) {
// 	if entity == nil {
// 		return nil
// 	}
// 	if entity.OwnerId == userId {
// 		return nil
// 	}
// 	if entity.OwnerOrganizationId == orgId && xsMode == AccessModeRead {
// 		return nil
// 	}
// 	if clientRole == WriteXSRoleSystem || clientRole == WriteXSRoleAdmin {
// 		return nil
// 	}
// 	return ErrUnauthorizedAccess
// }
