package base

import "strings"

// define AIgencyMessageType as enum
type AIgencyMessageType string //@name AIgencyMessageType

const (
	AIgencyMessageTypeMessage          AIgencyMessageType = "message"
	AIgencyMessageTypeStateUpdate      AIgencyMessageType = "stateUpdate"
	AIgencyMessageTypeDelta            AIgencyMessageType = "delta"
	AIgencyMessageTypeToolResponse     AIgencyMessageType = "toolResponse"
	AIgencyMessageTypeAnthropicToolUse AIgencyMessageType = "tool_use"
)

func (mtype AIgencyMessageType) String() string {
	return string(mtype)
}

func (mtype AIgencyMessageType) Set(value string) {
	switch strings.ToLower(value) {
	case "message":
		mtype = AIgencyMessageTypeMessage
	case "state-update":
		mtype = AIgencyMessageTypeStateUpdate
	default:
		mtype = AIgencyMessageTypeMessage
	}
}

func (mtype AIgencyMessageType) Equals(compareMTypeString string) bool {
	return strings.EqualFold(mtype.String(), compareMTypeString)
}

func (mtype AIgencyMessageType) EqualsRole(compareMType AIgencyMessageType) bool {
	return mtype == compareMType
}
