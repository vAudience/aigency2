package models

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

const (
	IDPREFIX_AIMODELSERVICE = "ams"
	IDLENGTH_AIMODELSERVICE = 16
)

var (
	ErrInvalidAIModelService = errors.New("invalid AI model service data")
)

type HostingLocation string //@name HostingLocation

const (
	HostingLocationUSA     HostingLocation = "usa"
	HostingLocationEU      HostingLocation = "europe"
	HostingLocationGERMANY HostingLocation = "germany"
	HostingLocationUK      HostingLocation = "uk"
	HostingLocationSWISS   HostingLocation = "swiss"
	HostingLocationANY     HostingLocation = "any"
)

type AIModelServiceWithModels struct {
	Service *AIModelServiceObject `json:"service"`
	Models  []*AIModel            `json:"models"`
} //@name AIModelServiceWithModels

type AIModelServiceWriteDto struct {
	Name             *string                     `json:"name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"*"`
	Description      *string                     `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"*"`
	CostMultiplier   *float64                    `json:"cost_multiplier" validate:"omitempty" writexs:"system,admin,owner" readxs:"system,admin,owner"`   // 1.0 is default, we use this to adjust our margin
	ServiceImpl      *string                     `json:"service_impl" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"` // this is used for internal identification!
	HostingLocations *map[string]HostingLocation `json:"hosting_locations" validate:"omitempty" writexs:"system,admin,owner" readxs:"admin,owner"`
	IsPublic         *bool                       `json:"is_public" writexs:"system,admin,owner" readxs:"admin,owner"`
} //@name AIModelServiceWriteDto

type ExecutionResult struct {
	ExecutionID  string           `json:"execution_id"` //maps to CompletionID for Text2Text
	ModelID      string           `json:"model_id"`
	ServiceID    string           `json:"service_id"`
	Timestamp    int64            `json:"timestamp"`   // in Milliseconds unix epoch
	TimeNeeded   int64            `json:"time_needed"` // in Milliseconds
	ErrorMessage string           `json:"error_message"`
	FinishReason string           `json:"finish_reason"`
	FeaturesUsed []AIModelFeature `json:"features_used"`
} //@name ExecutionResult

type ExecutionResultText2Text struct {
	ExecutionResult
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

type AIModelServiceObject struct {
	ID                  string                     `json:"id" validate:"required,min=1,max=64" writexs:"system" readxs:"*"`
	Name                string                     `json:"name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"*"`
	Description         string                     `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"*"`
	CostMultiplier      float64                    `json:"cost_multiplier" validate:"omitempty" writexs:"system,admin,owner" readxs:"system,admin,owner"`   // 1.0 is default, we use this to adjust our margin
	ServiceImpl         string                     `json:"service_impl" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"` // this is used for internal identification!
	HostingLocations    map[string]HostingLocation `json:"hosting_locations" validate:"omitempty" writexs:"system,admin,owner" readxs:"admin,owner"`
	IsPublic            bool                       `json:"is_public" writexs:"system,admin,owner" readxs:"admin,owner"`
	OwnerId             string                     `json:"owner_id" validate:"required,min=1,max=64" writexs:"system" readxs:"admin,owner"`
	OwnerOrganizationId string                     `json:"owner_organization_id" validate:"required,min=1,max=64" writexs:"system" readxs:"admin,owner"`
	CreatedAt           int64                      `json:"created_at" writexs:"system" readxs:"admin,owner"`
	UpdatedBy           string                     `json:"updated_by" validate:"omitempty,min=1,max=64" writexs:"system" readxs:"admin,owner"`
	UpdatedAt           int64                      `json:"updated_at" writexs:"system" readxs:"admin,owner"`
} //@name AIModelServiceObject

func NewAIModelService() *AIModelServiceObject {
	ID := CreateAIModelServiceID()
	now := nuts.TimeToJSTimestamp(time.Now())
	entity := AIModelServiceObject{
		ID:        ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return &entity
}

func CreateAIModelServiceID() string {
	return nuts.NID(IDPREFIX_AIMODELSERVICE, IDLENGTH_AIMODELSERVICE)
}

func IsAIModelServiceID(id string) (isModelID bool) {
	isModelID = (len(id) == IDLENGTH_AIMODELSERVICE+len(IDPREFIX_AIMODELSERVICE)+1) && strings.HasPrefix(id, IDPREFIX_AIMODELSERVICE)
	return isModelID
}

func ValidateAIModelService(service *AIModelServiceObject) error {
	validate := validator.New()

	err := validate.Struct(service)
	if err != nil {
		nuts.L.Debugf("[ValidateAIModelService] Validation error: %v\n%s", err, nuts.GetPrettyJson(service))
		return ErrInvalidAIModelService
	}

	return nil
}
