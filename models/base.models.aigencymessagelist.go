package models

import (
	"encoding/json"
	"sort"
	"sync"
)

type AIgencyMessageList struct {
	messages []*AIgencyMessage
	mu       sync.Mutex
} //@name AIgencyMessageList

func NewAIgencyMessageList() *AIgencyMessageList {
	return &AIgencyMessageList{
		messages: []*AIgencyMessage{},
	}
}

func (ml *AIgencyMessageList) AddMessage(message *AIgencyMessage) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.messages = append(ml.messages, message)
	ml.sortMessagesByCreatedAt()
}

func (ml *AIgencyMessageList) CopyMessages(other *AIgencyMessageList) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.messages = append(ml.messages, other.messages...)
	ml.sortMessagesByCreatedAt()
}

func (ml *AIgencyMessageList) AddMessages(messages []*AIgencyMessage) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.messages = append(ml.messages, messages...)
	ml.sortMessagesByCreatedAt()
}

func (ml *AIgencyMessageList) GetMessages() []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	return ml.messages
}

func (ml *AIgencyMessageList) GetLatestMessage() *AIgencyMessage {
	if len(ml.messages) > 0 {
		return ml.messages[len(ml.messages)-1]
	}
	return nil
}

func (ml *AIgencyMessageList) GetMessageByIndex(index int) *AIgencyMessage {
	if index >= 0 && index < len(ml.messages) {
		return ml.messages[index]
	}
	return nil
}

func (ml *AIgencyMessageList) ClearMessages() {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.messages = []*AIgencyMessage{}
}

func (ml *AIgencyMessageList) RemoveMessageByIndex(index int) {
	if index >= 0 && index < len(ml.messages) {
		ml.mu.Lock()
		defer ml.mu.Unlock()

		ml.messages = append(ml.messages[:index], ml.messages[index+1:]...)
		ml.sortMessagesByCreatedAt()
	}
}

func (ml *AIgencyMessageList) sortMessagesByCreatedAt() {
	sort.Slice(ml.messages, func(i, j int) bool {
		return ml.messages[i].CreatedAt < ml.messages[j].CreatedAt
	})
}

// GetMessagesByReferenceID retrieves messages by their reference ID
func (ml *AIgencyMessageList) GetMessagesByReferenceID(referenceID string) []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	var messages []*AIgencyMessage
	for _, msg := range ml.messages {
		if msg.ReferenceID == referenceID {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GetMessagesByResponseToID retrieves messages by their response-to ID
func (ml *AIgencyMessageList) GetMessagesByResponseToID(responseToID string) []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	var messages []*AIgencyMessage
	for _, msg := range ml.messages {
		if msg.ResponseToID == responseToID {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GetMessagesBySenderID retrieves messages by sender ID
func (ml *AIgencyMessageList) GetMessagesBySenderID(senderID string) []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	var messages []*AIgencyMessage
	for _, msg := range ml.messages {
		if msg.SenderID == senderID {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GetMessagesByRole retrieves messages by the sender's role
func (ml *AIgencyMessageList) GetMessagesByRole(role ConversationRole) []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	var messages []*AIgencyMessage
	for _, msg := range ml.messages {
		if msg.SenderConversationRole == role {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GetMessagesByType retrieves messages by type
func (ml *AIgencyMessageList) GetMessagesByType(messageType AIgencyMessageType) []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	var messages []*AIgencyMessage
	for _, msg := range ml.messages {
		if msg.Type == messageType {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GetLatestMessageByType retrieves the latest message by type
func (ml *AIgencyMessageList) GetLatestMessageByType(messageType AIgencyMessageType) *AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	for i := len(ml.messages) - 1; i >= 0; i-- {
		if ml.messages[i].Type == messageType {
			return ml.messages[i]
		}
	}
	return nil
}

// GetFirstMessageByType retrieves the first message of a given type
func (ml *AIgencyMessageList) GetFirstMessageByType(messageType AIgencyMessageType) []*AIgencyMessage {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	var messages []*AIgencyMessage
	for _, msg := range ml.messages {
		if msg.Type == messageType {
			messages = append(messages, msg)
			break
		}
	}
	return messages
}

// GetMessagesCount returns the count of messages in the list
func (ml *AIgencyMessageList) GetMessagesCount() int {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	return len(ml.messages)
}

// MarshalJSON customizes the JSON output of AIgencyMessageList
func (ml *AIgencyMessageList) MarshalJSON() ([]byte, error) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return json.Marshal(ml.messages)
}

// UnmarshalJSON customizes the JSON input for AIgencyMessageList
func (ml *AIgencyMessageList) UnmarshalJSON(data []byte) error {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return json.Unmarshal(data, &ml.messages)
}

// ensure that the sequence of messages always alternates between role Assistant and role User
// our method offers different violation resolve approaches:
// "ignore" the violation and keep the messages as they are
// "delete" + "keepNewest || keepOldest"  this deletes all but 1 message of the same role in sequence and keeps the newest or oldest message (oldest is lower index in the list)
// "merge" + "keepNewest || keepOldest" this merges all messages of the same role in sequence into one message by concatenating the content and keeps the newest or oldest message (oldest is lower index in the list)
// "insertSeparators" + "separatorContent" this inserts messages of the other role in between messages of the same role in sequence with the given separator content

type RoleSequenceViolationResolveStrategy string

const (
	RoleSequenceViolationResolveIgnore           RoleSequenceViolationResolveStrategy = "ignore"
	RoleSequenceViolationResolveDelete           RoleSequenceViolationResolveStrategy = "delete"
	RoleSequenceViolationResolveMerge            RoleSequenceViolationResolveStrategy = "merge"
	RoleSequenceViolationResolveInsertSeparators RoleSequenceViolationResolveStrategy = "insertSeparators"
)

type RoleSequenceViolationResolveStrategyKeepParameter string

const (
	RoleSequenceViolationResolveKeepNewest RoleSequenceViolationResolveStrategyKeepParameter = "keepNewest"
	RoleSequenceViolationResolveKeepOldest RoleSequenceViolationResolveStrategyKeepParameter = "keepOldest"
)

type RoleSequenceViolationResolveStrategyOptions struct {
	ResolveStrategy  RoleSequenceViolationResolveStrategy
	KeepParameter    RoleSequenceViolationResolveStrategyKeepParameter
	SeparatorContent string
}

func (ml *AIgencyMessageList) ResolveRoleSequenceViolations(resolveStrategy RoleSequenceViolationResolveStrategyOptions) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if len(ml.messages) < 2 {
		return
	}

	var resolvedMessages []*AIgencyMessage
	currentRole := ml.messages[0].SenderConversationRole
	sameRoleMessages := []*AIgencyMessage{ml.messages[0]}

	for i := 1; i < len(ml.messages); i++ {
		if ml.messages[i].SenderConversationRole == currentRole {
			sameRoleMessages = append(sameRoleMessages, ml.messages[i])
		} else {
			resolvedMessages = append(resolvedMessages, ml.resolveViolation(sameRoleMessages, resolveStrategy)...)
			currentRole = ml.messages[i].SenderConversationRole
			sameRoleMessages = []*AIgencyMessage{ml.messages[i]}
		}
	}

	// Resolve the last group of messages
	resolvedMessages = append(resolvedMessages, ml.resolveViolation(sameRoleMessages, resolveStrategy)...)

	ml.messages = resolvedMessages
}

func (ml *AIgencyMessageList) resolveViolation(messages []*AIgencyMessage, resolveStrategy RoleSequenceViolationResolveStrategyOptions) []*AIgencyMessage {
	if len(messages) == 1 {
		return messages
	}

	switch resolveStrategy.ResolveStrategy {
	case RoleSequenceViolationResolveIgnore:
		return messages

	case RoleSequenceViolationResolveDelete:
		if resolveStrategy.KeepParameter == RoleSequenceViolationResolveKeepNewest {
			return []*AIgencyMessage{messages[len(messages)-1]}
		}
		return []*AIgencyMessage{messages[0]}

	case RoleSequenceViolationResolveMerge:
		mergedMessage := &AIgencyMessage{
			Type:                   messages[0].Type,
			ReferenceID:            messages[0].ReferenceID,
			ResponseToID:           messages[0].ResponseToID,
			MissionID:              messages[0].MissionID,
			ChannelID:              messages[0].ChannelID,
			ChannelName:            messages[0].ChannelName,
			AIgentThreadID:         messages[0].AIgentThreadID,
			SenderID:               messages[0].SenderID,
			SenderName:             messages[0].SenderName,
			SenderConversationRole: messages[0].SenderConversationRole,
			OwnerOrganizationId:    messages[0].OwnerOrganizationId,
			MetaData:               make(map[string]any),
			CompletionParameters:   make(map[string]any),
			ChatCompletionConfig:   make(map[string]any),
			Content:                NewAIgencyMessageContentList(),
			Attachments:            NewAIgencyMessageFileList(),
			AIServiceID:            messages[0].AIServiceID,
			AIModelID:              messages[0].AIModelID,
			TokenDirection:         messages[0].TokenDirection,
			CreatedForFeature:      messages[0].CreatedForFeature,
		}

		for _, msg := range messages {
			// Merge Content
			if msg.Content != nil {
				for _, content := range msg.Content.GetData() {
					mergedMessage.Content.AddContent(content)
				}
			}

			// Merge Attachments
			if msg.Attachments != nil {
				for _, file := range msg.Attachments.GetFiles() {
					mergedMessage.Attachments.AddFile(file)
				}
			}

			// Merge other fields
			mergedMessage.TokenCount += msg.TokenCount
			for k, v := range msg.MetaData {
				mergedMessage.MetaData[k] = v
			}
			for k, v := range msg.CompletionParameters {
				mergedMessage.CompletionParameters[k] = v
			}
			for k, v := range msg.ChatCompletionConfig {
				mergedMessage.ChatCompletionConfig[k] = v
			}
		}

		if resolveStrategy.KeepParameter == RoleSequenceViolationResolveKeepNewest {
			mergedMessage.ID = messages[len(messages)-1].ID
			mergedMessage.CreatedAt = messages[len(messages)-1].CreatedAt
			mergedMessage.UpdatedAt = messages[len(messages)-1].UpdatedAt
		} else {
			mergedMessage.ID = messages[0].ID
			mergedMessage.CreatedAt = messages[0].CreatedAt
			mergedMessage.UpdatedAt = messages[0].UpdatedAt
		}

		return []*AIgencyMessage{mergedMessage}

	case RoleSequenceViolationResolveInsertSeparators:
		resolvedMessages := make([]*AIgencyMessage, 0, len(messages)*2-1)
		for i, msg := range messages {
			if i > 0 {
				separator := &AIgencyMessage{
					ID:                     CreateAIgencyMessageID(),
					Type:                   AIgencyMessageTypeMessage,
					MissionID:              msg.MissionID,
					ChannelID:              msg.ChannelID,
					ChannelName:            msg.ChannelName,
					AIgentThreadID:         msg.AIgentThreadID,
					SenderID:               "system",
					SenderName:             "System",
					SenderConversationRole: ConversationRoleSystem,
					OwnerOrganizationId:    msg.OwnerOrganizationId,
					Content:                NewAIgencyMessageContentList(),
					AIServiceID:            msg.AIServiceID,
					AIModelID:              msg.AIModelID,
					CreatedAt:              msg.CreatedAt - 1,
					UpdatedAt:              msg.CreatedAt - 1,
				}
				separatorContent := NewAIgencyMessageContent(AIgencyMessageContentTypeText, &resolveStrategy.SeparatorContent, nil)
				separator.Content.AddContent(separatorContent)
				resolvedMessages = append(resolvedMessages, separator)
			}
			resolvedMessages = append(resolvedMessages, msg)
		}
		return resolvedMessages

	default:
		return messages
	}
}
