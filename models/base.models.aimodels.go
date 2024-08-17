package models

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

var (
	ErrInvalidAIModel                = errors.New("invalid AI model data")
	ErrFeatureHasDifferentCapability = errors.New("feature has different capability")
)

const (
	IDPREFIX_AIMODEL = "aimodel"
	IDLENGTH_AIMODEL = 12
)

// AIModelCapability represents the capabilities of an AI model to execute certain actions

type AIModelCapability string //@name AIModelCapability

const (
	AIModelCapabilityFunctionCalling          AIModelCapability = "function-calling"
	AIModelCapabilityFunctionCallingStreaming AIModelCapability = "function-calling_streaming"
	AIModelCapabilityAcceptsDocumentFiles     AIModelCapability = "accepts-document-files"
	AIModelCapabilityTextToText               AIModelCapability = "text-to-text"
	AIModelCapabilityTextToTextStreaming      AIModelCapability = "text-to-text_streaming"
	AIModelCapabilityTextToImage              AIModelCapability = "text-to-image"
	AIModelCapabilityTextToSpeech             AIModelCapability = "text-to-speech"
	AIModelCapabilityTextToSpeechStreaming    AIModelCapability = "text-to-speech_streaming"
	AIModelCapabilityTextToMusic              AIModelCapability = "text-to-music"
	AIModelCapabilityTextToMusicStreaming     AIModelCapability = "text-to-music_streaming"
	AIModelCapabilityTextToVideo              AIModelCapability = "text-to-video"
	AIModelCapabilityTextToVideoStreaming     AIModelCapability = "text-to-video_streaming"
	AIModelCapabilitySpeechToText             AIModelCapability = "speech-to-text"
	AIModelCapabilitySpeechToTextStreaming    AIModelCapability = "speech-to-text_streaming"
	AIModelCapabilityImageToText              AIModelCapability = "image-to-text"
	AIModelCapabilityVideoToText              AIModelCapability = "video-to-text"
	AIModelCapabilityVideoToTextStreaming     AIModelCapability = "video-to-text_streaming"
)

// cost units for model Features
type AIModelCostUnit string //@name AIModelCostUnit

const (
	AIModelCostUnitInputPerMillionTokens     AIModelCostUnit = "input-tokens-per-million"
	AIModelCostUnitOutputPerMillionTokens    AIModelCostUnit = "output-tokens-per-million"
	AIModelCostUnitInputPerMillionCharacters AIModelCostUnit = "input-characters-per-million"
	AIModelCostUnitImageInputPerFile         AIModelCostUnit = "image-input-file"
	AIModelCostUnitAudioInputPerSecond       AIModelCostUnit = "audio-input-per-second"
	AIModelCostUnitVideoInputPerSecond       AIModelCostUnit = "video-input-per-second"
	AIModelCostUnitImageGenerationPerImage   AIModelCostUnit = "image-generation-per-image"
	AIModelCostUnitImageGenerationPerPixel   AIModelCostUnit = "image-generation-per-pixel"
	AIModelCostUnitAudioGenerationPerSecond  AIModelCostUnit = "audio-generation-per-second"
	AIModelCostUnitVideoGenerationPerSecond  AIModelCostUnit = "video-generation-per-second"
	AIModelCostUnitPerFunctionCall           AIModelCostUnit = "per-function-call"
)

type AIModelConstraintDirection string //@name AIModelConstraintDirection

const (
	AIModelConstraintDirectionMin AIModelConstraintDirection = "input"
	AIModelConstraintDirectionMax AIModelConstraintDirection = "output"
)

type AIModelMinMaxUnit string //@name AIModelMinMaxUnit

const (
	AIModelMinMaxUnitTokens            AIModelMinMaxUnit = "tokens"
	AIModelMinMaxUnitCharacters        AIModelMinMaxUnit = "characters"
	AIModelMinMaxUnitFiles             AIModelMinMaxUnit = "files"
	AIModelMinMaxUnitSeconds           AIModelMinMaxUnit = "seconds"
	AIModelMinMaxUnitImages            AIModelMinMaxUnit = "images"
	AIModelMinMaxUnitPixels            AIModelMinMaxUnit = "pixels"
	AIModelMinMaxUnitFilesizeMegabytes AIModelMinMaxUnit = "megabytes"
)

type AIModelConstraint struct {
	Direction AIModelConstraintDirection `json:"direction" writexs:"system" readxs:"*"`
	Min       float64                    `json:"min" writexs:"system" readxs:"*"`
	Max       float64                    `json:"max" writexs:"system" readxs:"*"`
	Unit      AIModelMinMaxUnit          `json:"unit" writexs:"system" readxs:"*"`
} //@name AIModelConstraint

/*  COST CALCULATION SYSTEM

--> a "Execution" should first identify the AIModel and all its Features.
--> Then, for each feature, it should identify the corresponding ExecutionCostTemplate and calculate the cost based on the usedUnits and the CostPerUnitInEuro
--> Finally, it should sum up all costs for all features and return the total cost

*/

// a AIModelFeature describes ONE capability of a AIModel and the associated cost caluclations templates, since there might be multiple parameters like functioncalling, etc.
type AIModelFeature struct {
	Capability        AIModelCapability       `json:"capability"`
	Constraints       []AIModelConstraint     `json:"constraints"`
	CostItemTemplates []ExecutionCostTemplate `json:"cost_item_templates"`
	CostItems         []ExecutionUsageCost    `json:"cost_items"`
} //@name AIModelFeature

func (feat *AIModelFeature) GetCostItemByCostUnit(costUnit AIModelCostUnit) *ExecutionCostTemplate {
	// nuts.L.Debugf("[GetCostItemByCostUnit] Checking cost items for costUnit(%v) via feat\n%s", costUnit, nuts.GetPrettyJson(feat))
	for _, costItem := range feat.CostItemTemplates {
		// nuts.L.Debugf("[GetCostItemByCostUnit] Checking cost item for costUnit(%v): %v", costUnit, costItem.CostUnit)
		if costItem.CostUnit == costUnit {
			// nuts.L.Debugf("[GetCostItemByCostUnit] Found cost item for costUnit(%v): %v", costUnit, costItem)
			return &costItem
		}
	}
	// nuts.L.Debugf("[GetCostItemByCostUnit] No cost item found for costUnit(%v)", costUnit)
	return nil
}

func (feat *AIModelFeature) CreateUsedFeatures(capability AIModelCapability, costUnit AIModelCostUnit, usedUnits float64, multiplier float64) (usedFeatures []AIModelFeature, err error) {
	usedFeatures = make([]AIModelFeature, 0)
	if feat.Capability != capability {
		return usedFeatures, ErrFeatureHasDifferentCapability
	}
	for _, costTemplate := range feat.CostItemTemplates {
		if costTemplate.CostUnit != costUnit {
			continue
		}
		usageCost := NewExecutionUsageCost(&costTemplate, usedUnits, multiplier)
		usedFeatures = append(usedFeatures, AIModelFeature{
			Capability: feat.Capability,
			CostItems:  []ExecutionUsageCost{*usageCost},
			// Constraints: feat.Constraints,
			// TODO: this needs to be filtered later - for alpha use, this is removed for now to keep multiplier in the cost calculation secret
			// CostItemTemplates: feat.CostItemTemplates,
		})
	}
	return usedFeatures, nil
}

// a ExecutionCostTemplate describes the cost calculation for a single, specific AIModelFeature
type ExecutionCostTemplate struct {
	Description       string          `json:"description" writexs:"system" readxs:"system,admin,orgadmin"`
	CostUnit          AIModelCostUnit `json:"cost_unit" writexs:"system" readxs:"system,admin,orgadmin"`
	CostPerUnitInEuro float64         `json:"cost_per_unit_in_euro" writexs:"system" readxs:"system,admin,orgadmin"`
} //@name ExecutionCostTemplate

// for each aimodel capability, there is a specific ExecutionCostTemplate and we use this and add usedUnits along with the resulting cost in euro
type ExecutionUsageCost struct {
	ExecutionCostTemplate
	UsedUnits                 float64 `json:"used_units" writexs:"system" readxs:"system,admin,orgadmin"`
	ResultingCostInEuro       float64 `json:"resulting_cost_in_euro" writexs:"system" readxs:"system,admin,orgadmin"`
	ResultingSourceCostInEuro float64 `json:"-" writexs:"system" readxs:"system,superadmin"`
} //@name ExecutionUsageCost

func NewExecutionUsageCost(template *ExecutionCostTemplate, usedUnits float64, multiplier float64) *ExecutionUsageCost {
	unitRelativeUsedUnits := usedUnits
	if template.CostUnit == AIModelCostUnitOutputPerMillionTokens {
		unitRelativeUsedUnits = usedUnits / 1_000_000
	} else if template.CostUnit == AIModelCostUnitInputPerMillionCharacters {
		unitRelativeUsedUnits = usedUnits / 1_000_000
	} else if template.CostUnit == AIModelCostUnitInputPerMillionTokens {
		unitRelativeUsedUnits = usedUnits / 1_000_000
	}
	templateCopy := *template
	// TODO: this needs to be filtered later - for alpha use, this is removed for now to keep multiplier in the cost calculation secret
	templateCopy.CostPerUnitInEuro = template.CostPerUnitInEuro * multiplier
	return &ExecutionUsageCost{
		ExecutionCostTemplate:     templateCopy,
		UsedUnits:                 unitRelativeUsedUnits,
		ResultingCostInEuro:       template.CostPerUnitInEuro * unitRelativeUsedUnits * multiplier,
		ResultingSourceCostInEuro: template.CostPerUnitInEuro * unitRelativeUsedUnits,
	}
}

// -- DTOs and DB Models --

type AIModelWriteDto struct {
	Name                  *string              `json:"name" validate:"omitempty,min=1,max=255" writexs:"system,admin,owner" readxs:"*"`
	Description           *string              `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"*"`
	DocumentationUrl      *string              `json:"documentation_url" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	ServiceID             *string              `json:"service_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	ModelID               *string              `json:"model_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"*"`
	MaxInputTokens        *int                 `json:"max_input_tokens" validate:"gte=0" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	MaxOutputTokens       *int                 `json:"max_output_tokens" validate:"gte=0" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Constraints           *[]AIModelConstraint `json:"constraints" writexs:"system,admin,owner" readxs:"*"`
	Features              *[]AIModelFeature    `json:"features" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	ServiceHostLocations  *[]HostingLocation   `json:"service_host_locations" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	AcceptedFileMimetypes *[]string            `json:"accepted_file_mimetypes" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Parameters            *map[string]any      `json:"parameters" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	IsPublic              *bool                `json:"is_public" writexs:"system,admin,owner" readxs:"system,admin,owner"`
} //@name AIModelWriteDto

type AIModel struct {
	ID                    string              `json:"id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	Name                  string              `json:"name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"*"`
	Description           string              `json:"description" validate:"omitempty,max=1024" writexs:"system,admin,owner" readxs:"*"`
	DocumentationUrl      string              `json:"documentation_url" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	ServiceID             string              `json:"service_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	ModelID               string              `json:"model_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	MaxInputTokens        int                 `json:"max_input_tokens" validate:"gte=0" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	MaxOutputTokens       int                 `json:"max_output_tokens" validate:"gte=0" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Constraints           []AIModelConstraint `json:"constraints" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Features              []AIModelFeature    `json:"features" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	ServiceHostLocations  []HostingLocation   `json:"service_host_locations" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	AcceptedFileMimetypes []string            `json:"accepted_file_mimetypes" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	Parameters            map[string]any      `json:"parameters" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	IsPublic              bool                `json:"is_public" writexs:"system,admin,owner" readxs:"system,admin,owner"`
	OwnerId               string              `json:"owner_id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	OwnerOrganizationId   string              `json:"owner_organization_id" validate:"required,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	CreatedAt             int64               `json:"created_at" writexs:"system" readxs:"system,admin,owner"`
	UpdatedBy             string              `json:"updated_by" validate:"omitempty,min=1,max=64" writexs:"system" readxs:"system,admin,owner"`
	UpdatedAt             int64               `json:"updated_at" writexs:"system" readxs:"system,admin,owner"`
} //@name AIModel

func NewAIModel() *AIModel {
	ID := CreateAIModelID()
	now := nuts.TimeToJSTimestamp(time.Now())
	entity := AIModel{
		ID:         ID,
		Parameters: make(map[string]any),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return &entity
}

func CreateAIModelID() string {
	return nuts.NID(IDPREFIX_AIMODEL, IDLENGTH_AIMODEL)
}

func IsAIModelID(id string) (isModelID bool) {
	isModelID = (len(id) == IDLENGTH_AIMODEL+len(IDPREFIX_AIMODEL)+1) && strings.HasPrefix(id, IDPREFIX_AIMODEL)
	return isModelID
}

func ValidateAIModel(aiModel *AIModel) error {
	validate := validator.New()

	err := validate.Struct(aiModel)
	if err != nil {
		return ErrInvalidAIModel
	}

	return nil
}

func (aiModel *AIModel) Validate() error {
	return ValidateAIModel(aiModel)
}

func (aiModel *AIModel) GetFeaturesForCapability(capability AIModelCapability) []*AIModelFeature {
	features := make([]*AIModelFeature, 0)
	for _, feature := range aiModel.Features {
		if feature.Capability == capability {
			features = append(features, &feature)
		}
	}
	return features
}

func (aiModel *AIModel) CalculateUsageCostsForFeature(capability AIModelCapability, costUnit AIModelCostUnit, usedUnits float64, multiplier float64) (totalUsedFeatures []AIModelFeature, err error) {
	totalUsedFeatures = make([]AIModelFeature, 0)
	for _, feature := range aiModel.Features {
		// nuts.L.Debugf("[CalculateUsageCostsForFeature] Checking Model(%s) feature: (%v) and costUnit(%v)", aiModel.Name, feature.Capability, costUnit)
		if feature.Capability != capability {
			// nuts.L.Debugf("Feature not found or has different capability feat(%v) capa(%v)", feature.Capability, capability)
			continue
		}
		if feature.GetCostItemByCostUnit(costUnit) == nil {
			// nuts.L.Debugf("Cost item not found for costUnit(%v)", costUnit)
			continue
		}
		usedFeatures, err := feature.CreateUsedFeatures(capability, costUnit, usedUnits, multiplier)
		if err != nil {
			return totalUsedFeatures, err
		}
		totalUsedFeatures = append(totalUsedFeatures, usedFeatures...)
	}
	return totalUsedFeatures, nil
}

func CalculateCostForText2Text(aiModel *AIModel, toolCallsUsed int, inputTokenCount int, outputTokenCount int, costMultiplier float64) (featuresUsed []AIModelFeature, err error) {
	nuts.L.Debugf("Calculating costs for AIModel(%s) with toolCallsUsed(%d), inputTokenCount(%d), outputTokenCount(%d)", aiModel.ID, toolCallsUsed, inputTokenCount, outputTokenCount)
	if outputTokenCount > 0 {
		usedFeats, err := aiModel.CalculateUsageCostsForFeature(AIModelCapabilityTextToText, AIModelCostUnitOutputPerMillionTokens, float64(outputTokenCount), costMultiplier)
		if err != nil {
			nuts.L.Errorf("failed to calculate feature costs: %w", err)
		}
		// nuts.L.Debugf("Used features for outputTokenCount: %v", usedFeats)
		featuresUsed = append(featuresUsed, usedFeats...)
	}
	if inputTokenCount > 0 {
		usedFeats, err := aiModel.CalculateUsageCostsForFeature(AIModelCapabilityTextToText, AIModelCostUnitOutputPerMillionTokens, float64(inputTokenCount), costMultiplier)
		if err != nil {
			nuts.L.Errorf("failed to calculate feature costs: %w", err)
		}
		// nuts.L.Debugf("Used features for inputTokenCount: %v", usedFeats)
		featuresUsed = append(featuresUsed, usedFeats...)
	}
	if toolCallsUsed > 0 {
		usedFeats, err := aiModel.CalculateUsageCostsForFeature(AIModelCapabilityFunctionCalling, AIModelCostUnitPerFunctionCall, float64(toolCallsUsed), costMultiplier)
		if err != nil {
			nuts.L.Errorf("failed to calculate feature costs: %w", err)
		}
		// nuts.L.Debugf("Used features for toolCallsUsed: %v", usedFeats)
		featuresUsed = append(featuresUsed, usedFeats...)
	}
	return featuresUsed, nil
}
