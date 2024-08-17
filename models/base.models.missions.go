package models

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	nuts "github.com/vaudience/go-nuts"
)

var (
	ErrInvalidMission               = errors.New("invalid mission data")
	ErrMissionStatusUpdateInvalid   = errors.New("invalid mission status update data")
	ErrMissionStatusUpdateAfterDone = errors.New("mission status update after mission is done")
)

const (
	IDPREFIX_MISSION             = "mis"
	IDLENGTH_MISSION             = 16
	IDPREFIX_MISSIONSTATUSUPDATE = "msu"
	IDLENGTH_MISSIONSTATUSUPDATE = 16
)

type MissionInstructionsDto struct {
	Text string `json:"text" example:"Tell me a fun fact about basketball"`
} //@name MissionInstructionsDto

type MissionIssueDto struct {
	Description  string                  `json:"description" example:"Mission to find a fun fact about basketball"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Instructions *MissionInstructionsDto `json:"instructions" example:"Tell me a fun fact about basketball"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
}

type MissionResultsDto struct {
	AgentTeamID  string           `json:"agent_team_id" example:"agteam_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	MissionID    string           `json:"mission_id" example:"mis_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	ExecutionID  string           `json:"execution_id" example:"exe_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	ChannelID    string           `json:"channel_id" example:"ch_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	InputText    string           `json:"input_text" example:"Tell me a fun fact about basketball"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	ResponseText string           `json:"response_text" example:"Basketball is the only major American sport with a clearly identifiable inventor. James Naismith wrote the sport’s original 13 rules as part of a December 1891 class assignment at a Young Men’s Christian Association (YMCA) training school in Springfield, Massachusetts."  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	ErrorMessage string           `json:"error_message" example:""  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	FinishReason string           `json:"finish_reason" example:"Mission completed successfully"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	InputTokens  int              `json:"input_tokens" writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	OutputTokens int              `json:"output_tokens" writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Timestamp    int64            `json:"timestamp" example:"1620000000000"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	TimeNeeded   int64            `json:"time_needed" example:"1620000000000"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	FeaturesUsed []AIModelFeature `json:"features_used" writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
} //@name MissionResultsDto

type Mission struct {
	ID                  string                   `json:"id" example:"mis_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	OwnerOrganizationID string                   `json:"owner_organization_id" example:"org_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	OwnerID             string                   `json:"owner_id" example:"usr_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	CreatorName         string                   `json:"creator_name" example:"John Doe"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	MissionTeamID       string                   `json:"mission_team_id" example:"agteam_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Content             *AIgencyMessageList      `json:"content"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Description         string                   `json:"description" example:"Mission to find a fun fact about basketball"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	StatusUpdates       *MissionStatusUpdateList `json:"status_updates" writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	CreatedAt           int64                    `json:"created_at" example:"1620000000000"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	CreatedBy           string                   `json:"created_by" example:"usr_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	UpdateAt            int64                    `json:"updated_at" example:"1620000000000"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	CompletedAt         int64                    `json:"completed_at" example:"1620000000000"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	CompletionReson     string                   `json:"completion_reason" example:"Mission completed successfully"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
} //@name Mission

func NewMission() *Mission {
	mission := &Mission{}
	mission.ID = CreateMissionID()
	mission.StatusUpdates = NewMissionStatusUpdateList()
	mission.Content = NewAIgencyMessageList()
	return mission
}

func CreateMissionID() string {
	return nuts.NID(IDPREFIX_MISSION, IDLENGTH_MISSION)
}

func IsMissionID(id string) bool {
	isMatchingID := (len(id) == IDLENGTH_MISSION+len(IDPREFIX_MISSION)+1) && strings.HasPrefix(id, IDPREFIX_MISSION)
	return isMatchingID
}

func ValidateMission(entity *Mission) error {
	if entity == nil {
		return ErrInvalidMission
	}
	validate := validator.New()

	err := validate.Struct(entity)
	if err != nil {
		return ErrInvalidMission
	}

	if !IsMissionID(entity.ID) {
		return ErrInvalidMission
	}
	return nil
}

// ========= MissionStatusUpdate =========

type MissionStatus string //@name MissionStatus

const (
	MissionStatusCreated   MissionStatus = "created"
	MissionStatusStarted   MissionStatus = "started"
	MissionStatusPaused    MissionStatus = "paused"
	MissionStatusResumed   MissionStatus = "resumed"
	MissionStatusCanceled  MissionStatus = "canceled"
	MissionStatusFailed    MissionStatus = "failed"
	MissionStatusCompleted MissionStatus = "completed"
)

type MissionStatusUpdate struct {
	ID          string        `json:"id" example:"msu_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	MissionID   string        `json:"mission_id" example:"mis_1234567890123456"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Status      MissionStatus `json:"status" writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Description string        `json:"description" example:"Mission started"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
	Timestamp   int64         `json:"timestamp" example:"1620000000000"  writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
} //@name MissionStatusUpdate

type MissionStatusUpdateList struct {
	safety        sync.Mutex
	StatusUpdates []*MissionStatusUpdate `json:"status_updates" writexs:"system,admin,org,user" readxs:"system,admin,org,user"`
} //@name MissionStatusUpdateList

func NewMissionStatusUpdateList() *MissionStatusUpdateList {
	msul := &MissionStatusUpdateList{}
	msul.StatusUpdates = make([]*MissionStatusUpdate, 0)
	return msul
}

func NewMissionStatusUpdate(mission *Mission, status MissionStatus, description string) *MissionStatusUpdate {
	update := &MissionStatusUpdate{}
	update.ID = CreateMissionStatusUpdateID()
	update.Status = status
	update.Description = description
	update.Timestamp = nuts.TimeToJSTimestamp(time.Now())
	update.MissionID = mission.ID
	return update
}

func CreateMissionStatusUpdateID() string {
	return nuts.NID(IDPREFIX_MISSIONSTATUSUPDATE, IDLENGTH_MISSIONSTATUSUPDATE)
}

func IsMissionStatusUpdateID(id string) bool {
	isMatchingID := (len(id) == len(IDPREFIX_MISSIONSTATUSUPDATE)+1+IDLENGTH_MISSIONSTATUSUPDATE) && strings.HasPrefix(id, IDPREFIX_MISSIONSTATUSUPDATE)
	return isMatchingID
}

func (msu *MissionStatusUpdate) IsDone() bool {
	return msu.Status == MissionStatusCompleted || msu.Status == MissionStatusFailed || msu.Status == MissionStatusCanceled
}

func (msu *MissionStatusUpdate) IsRunning() bool {
	return msu.Status == MissionStatusStarted || msu.Status == MissionStatusPaused || msu.Status == MissionStatusResumed
}

func (msulist *MissionStatusUpdateList) MarshalJSON() ([]byte, error) {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	return json.Marshal(msulist.StatusUpdates)
}

func (msulist *MissionStatusUpdateList) UnmarshalJSON(data []byte) error {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	return json.Unmarshal(data, &msulist.StatusUpdates)
}

func (msulist *MissionStatusUpdateList) AddStatus(msu *MissionStatusUpdate) error {
	if msu == nil {
		nuts.L.Debugf("[MissionStatusUpdateList.AddStatus] Invalid status update(%v)", msu)
		return ErrMissionStatusUpdateInvalid
	}
	if !IsMissionStatusUpdateID(msu.ID) {
		nuts.L.Debugf("[MissionStatusUpdateList.AddStatus] Invalid status update ID(%s)", msu.ID)
		return ErrMissionStatusUpdateInvalid
	}
	// check if latestStatusUpdate is already done - if so, return error
	if msulist.IsDone() {
		nuts.L.Debugf("[MissionStatusUpdateList.AddStatus] Mission(%s) is already done", msu.MissionID)
		return ErrMissionStatusUpdateAfterDone
	}
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	msulist.StatusUpdates = append(msulist.StatusUpdates, msu)
	nuts.L.Debugf("[MissionStatusUpdateList.AddStatus] Added status update(%s)(%s) to mission(%s)", msu.ID, msu.Description, msu.MissionID)
	return nil
}

func (msulist *MissionStatusUpdateList) GetLatestStatusUpdate() *MissionStatusUpdate {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	if len(msulist.StatusUpdates) == 0 {
		return nil
	}
	return msulist.StatusUpdates[len(msulist.StatusUpdates)-1]
}

func (msulist *MissionStatusUpdateList) GetStatusUpdates() []*MissionStatusUpdate {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	statusListCopy := make([]*MissionStatusUpdate, len(msulist.StatusUpdates))
	copy(statusListCopy, msulist.StatusUpdates)
	return statusListCopy
}

func (msulist *MissionStatusUpdateList) GetStatusUpdatesCount() int {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	return len(msulist.StatusUpdates)
}

func (msulist *MissionStatusUpdateList) IsDone() bool {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	if len(msulist.StatusUpdates) == 0 {
		return false
	}
	latestUpdate := msulist.StatusUpdates[len(msulist.StatusUpdates)-1]
	return latestUpdate.IsDone()
}

func (msulist *MissionStatusUpdateList) IsRunning() bool {
	msulist.safety.Lock()
	defer msulist.safety.Unlock()
	if len(msulist.StatusUpdates) == 0 {
		return false
	}
	latestUpdate := msulist.StatusUpdates[len(msulist.StatusUpdates)-1]
	return latestUpdate.IsRunning()
}
