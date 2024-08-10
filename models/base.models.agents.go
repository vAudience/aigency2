// base/base.models.agents.go
package base

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

var (
	ErrInvalidAgent = errors.New("invalid agent data")
	// ErrUnauthorizedAccess = errors.New("unauthorized access")
)

const (
	IDPREFIX_AGENT = "agent"
	IDLENGTH_AGENT = 16
)

type AgentWriteDto struct {
	Name                *string          `json:"name" validate:"omitempty,min=1,max=255" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	Description         *string          `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	ModelID             *string          `json:"model_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	AvatarURL           *string          `json:"avatar_url" validate:"omitempty,url" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	SystemMessages      *[]string        `json:"system_messages" validate:"omitempty,unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	InitialUserMessages *[]string        `json:"initial_user_messages" validate:"omitempty,unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	AttachedFileIDs     *[]string        `json:"attached_file_ids" validate:"omitempty,unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	AssignedTools       *[]string        `json:"assigned_tools" validate:"omitempty,unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	IsPublic            *bool            `json:"is_public" writexs:"system,admin,owner" readxs:"admin,owner"`
	ModelHostLocation   *HostingLocation `json:"model_host_location" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	MetaData            *map[string]any  `json:"meta_data" writexs:"system,admin,owner" readxs:"system,admin,owner"`
} //@name AgentWriteDto

type Agent struct {
	ID                  string          `json:"id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner,org"`
	Name                string          `json:"name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	Description         string          `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	ModelID             string          `json:"model_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	AvatarURL           string          `json:"avatar_url" validate:"omitempty,url" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	SystemMessages      []string        `json:"system_messages" validate:"unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	InitialUserMessages []string        `json:"initial_user_messages" validate:"unique,dive,min=1,max=1024" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	AttachedFileIDs     []string        `json:"attached_file_ids" validate:"unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	AssignedTools       []string        `json:"assigned_tools" validate:"unique,dive,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	ModelHostLocation   HostingLocation `json:"model_host_location" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	IsPublic            bool            `json:"is_public" writexs:"system,admin,owner" readxs:"admin,owner"`
	MetaData            map[string]any  `json:"meta_data" writexs:"system,admin,owner" readxs:"system,admin,owner,org"`
	OwnerId             string          `json:"owner_id" validate:"required,min=1,max=64" writexs:"system,admin" readxs:"system,admin,owner,org"`
	OwnerOrganizationId string          `json:"owner_organization_id" validate:"required,min=1,max=64" writexs:"system,admin" readxs:"system,admin,owner,org"`
	CreatedAt           int64           `json:"created_at" writexs:"system,admin" readxs:"system,admin,owner,org"`
	UpdatedBy           string          `json:"updated_by" validate:"omitempty,min=1,max=64" writexs:"system,admin" readxs:"system,admin,owner,org"`
	UpdatedAt           int64           `json:"updated_at" writexs:"system,admin" readxs:"system,admin,owner,org"`
} //@name Agent

func NewAgent() *Agent {
	ID := CreateAgentID()
	now := nuts.TimeToJSTimestamp(time.Now())
	entity := Agent{
		ID:        ID,
		MetaData:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}
	return &entity
}

func CreateAgentID() string {
	return nuts.NID(IDPREFIX_AGENT, IDLENGTH_AGENT)
}

func IsAgentID(id string) (isModelID bool) {
	isModelID = (len(id) == IDLENGTH_AGENT+len(IDPREFIX_AGENT)+1) && strings.HasPrefix(id, IDPREFIX_AGENT)
	return isModelID
}

func ValidateAgent(agent *Agent) error {
	validate := validator.New()

	err := validate.Struct(agent)
	if err != nil {
		return ErrInvalidAgent
	}

	return nil
}

func (agent *Agent) GetSystemMessagesAsString() string {
	return strings.Join(agent.SystemMessages, "\n")
}
