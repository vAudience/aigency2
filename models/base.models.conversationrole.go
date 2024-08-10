package base

import "strings"

type ConversationRole string //@name ConversationRole

const (
	ConversationRoleUnknown   ConversationRole = "unknown"
	ConversationRoleUser      ConversationRole = "user"
	ConversationRoleSystem    ConversationRole = "system"
	ConversationRoleAssistant ConversationRole = "assistant"
	ConversationRoleAIgent    ConversationRole = "assistant"
)

func (role ConversationRole) String() string {
	return string(role)
}

func (role ConversationRole) Set(value string) {
	switch strings.ToLower(value) {
	case "user":
		role = ConversationRoleUser
	case "aigent":
		role = ConversationRoleAIgent
	case "system":
		role = ConversationRoleAssistant
	case "assistant":
		role = ConversationRoleAssistant
	default:
		role = ConversationRoleUnknown
	}
}

func (role ConversationRole) Equals(compareRoleString string) bool {
	return strings.EqualFold(role.String(), compareRoleString)
}

func (role ConversationRole) EqualsRole(compareRole ConversationRole) bool {
	return role == compareRole
}
