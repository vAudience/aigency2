package models

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

var (
	ErrInvalidOrgCostBudget = errors.New("invalid organization cost budget data")
)

type OrgCostBudgetWriteDto struct {
	TotalBudget *float64 `json:"total_budget" validate:"omitempty,min=0" writexs:"system:struct,admin:struct,owner:struct" readxs:"system:struct,admin:struct,owner:struct"`
	UsedBudget  *float64 `json:"used_budget" validate:"omitempty,min=0" writexs:"system:struct,admin:struct" readxs:"system:struct,admin:struct,owner:struct"`
} //@name OrgCostBudgetWriteDto

type OrgCostBudget struct {
	OrgID           string  `json:"org_id" validate:"required,min=1,max=64" writexs:"system:struct" readxs:"system:struct,admin:struct,owner:struct,org-owner:struct"`
	TotalBudget     float64 `json:"total_budget" validate:"min=0" writexs:"system:struct,admin:struct,owner:struct" readxs:"system:struct,admin:struct,owner:struct,org-owner:struct"`
	UsedBudget      float64 `json:"used_budget" validate:"min=0" writexs:"system:struct,admin:struct" readxs:"system:struct,admin:struct,owner:struct,org-owner:struct"`
	RemainingBudget float64 `json:"remaining_budget" validate:"min=0" writexs:"system:struct" readxs:"system:struct,admin:struct,owner:struct,org-owner:struct"`
	UpdatedAt       int64   `json:"updated_at" writexs:"system:struct" readxs:"system:struct,admin:struct,owner:struct,org-owner:struct"`
	UpdatedBy       string  `json:"updated_by" writexs:"system:struct" readxs:"system:struct,admin:struct,owner:struct,org-owner:struct"`
} //@name OrgCostBudget

type OrgCostBudgetCheck struct {
	OrgID            string `json:"org_id" validate:"required"`
	SufficientBudget bool   `json:"sufficient_budget" validate:"required"`
} //@name OrgCostBudgetCheck

func NewOrgCostBudget(orgID string) *OrgCostBudget {
	now := nuts.TimeToJSTimestamp(time.Now())
	entity := OrgCostBudget{
		OrgID:     orgID,
		UpdatedAt: now,
	}
	return &entity
}

func ValidateOrgCostBudget(budget *OrgCostBudget) error {
	validate := validator.New()
	err := validate.Struct(budget)
	if err != nil {
		return ErrInvalidOrgCostBudget
	}
	return nil
}
