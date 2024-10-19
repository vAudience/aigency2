package models

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

func IsAgentPromptID(id string) bool {
	if !strings.HasPrefix(id, IDPREFIX_AGENT_PROMPT) {
		return false
	}
	return len(id) == IDLENGTH_AGENT_PROMPT+len(IDPREFIX_AGENT_PROMPT)
}

const (
	IDPREFIX_AGENT_PROMPT = "prt"
	IDLENGTH_AGENT_PROMPT = 12
	AIC_VAR_SEPARATOR     = ";"
	AIC_VERSION_SEPARATOR = "@"
	MAX_PREVIOUS_VERSIONS = 3
)

type AgentPromptVisibilityStates string

const (
	AgentPromptVisibilityPrivate      AgentPromptVisibilityStates = "private"
	AgentPromptVisibilityOrganization AgentPromptVisibilityStates = "org"
	AgentPromptVisibilityPublic       AgentPromptVisibilityStates = "public"
	AgentPromptVisibilityCurated      AgentPromptVisibilityStates = "curated"
)

type PromptInjectionMode string

const (
	PromptInjectionModeAppend  PromptInjectionMode = "append"
	PromptInjectionModePrepend PromptInjectionMode = "prepend"
	PromptInjectionModeReplace PromptInjectionMode = "replace"
)

type PromptVersion struct {
	Version   int    `json:"version"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

type AgentPrompt struct {
	ID                  string                      `json:"id" validate:"required"`
	Title               string                      `json:"title" validate:"required"`
	Description         string                      `json:"description"`
	ThumbnailUrl        string                      `json:"thumbnail_url"`
	Tags                []string                    `json:"tags"`
	Space               string                      `json:"space"`
	CurrentVersion      int                         `json:"current_version"`
	Versions            []*PromptVersion            `json:"versions"`
	OwnerId             string                      `json:"owner_id" validate:"required"`
	OwnerOrganizationId string                      `json:"owner_organization_id" validate:"required"`
	CreatedAt           int64                       `json:"created_at"`
	UpdatedAt           int64                       `json:"updated_at"`
	Visibility          AgentPromptVisibilityStates `json:"visibility"`
}

type AgentPromptRenderDto struct {
	PromptID        string            `json:"prompt_id"`
	Content         string            `json:"content"`
	VarReplacements map[string]string `json:"var_replacements"`
}

func NewAgentPrompt() *AgentPrompt {
	return &AgentPrompt{
		ID:             CreateAgentPromptID(),
		CreatedAt:      time.Now().UnixMilli(),
		UpdatedAt:      time.Now().UnixMilli(),
		CurrentVersion: 1,
		Versions:       []*PromptVersion{},
		Visibility:     AgentPromptVisibilityPrivate,
	}
}

func CreateAgentPromptID() string {
	return nuts.NID(IDPREFIX_AGENT_PROMPT, IDLENGTH_AGENT_PROMPT)
}

func ValidateAgentPrompt(prompt *AgentPrompt) error {
	validate := validator.New()
	return validate.Struct(prompt)
}

func (ap *AgentPrompt) AddVersion(content string, createdBy string) {
	newVersion := &PromptVersion{
		Version:   ap.CurrentVersion,
		Content:   content,
		CreatedAt: time.Now().UnixMilli(),
		CreatedBy: createdBy,
	}

	ap.Versions = append(ap.Versions, newVersion)
	ap.CurrentVersion++
	ap.UpdatedAt = time.Now().UnixMilli()

	// Keep only the first version and the last MAX_PREVIOUS_VERSIONS
	if len(ap.Versions) > MAX_PREVIOUS_VERSIONS+1 {
		ap.Versions = append([]*PromptVersion{ap.Versions[0]}, ap.Versions[len(ap.Versions)-MAX_PREVIOUS_VERSIONS:]...)
	}
}

func (ap *AgentPrompt) GetVersion(version int) *PromptVersion {
	if version == 0 {
		return ap.Versions[len(ap.Versions)-1] // Latest version
	}
	for _, v := range ap.Versions {
		if v.Version == version {
			return v
		}
	}
	return nil
}
