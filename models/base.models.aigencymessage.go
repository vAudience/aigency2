package models

import (
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkoukk/tiktoken-go"
	"github.com/tiktoken-go/tokenizer"
	nuts "github.com/vaudience/go-nuts"
)

const (
	IDPREFIX_AIGENCYMESSAGE = "msg"
	IDLENGTH_AIGENCYMESSAGE = 16
)

var (
	DefaultTextTokenEncoderModel = string(tokenizer.GPT4)
	DefaultTextTokenEncoder      *tiktoken.Tiktoken
)

type AIgencyMessageWriteDto struct {
	Type                   *AIgencyMessageType        `json:"type" validate:"required,oneof=message stateUpdate delta" writexs:"system,admin,owner" readxs:"admin,owner"`
	ReferenceID            *string                    `json:"reference_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	ResponseToID           *string                    `json:"response_to_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	MissionID              *string                    `json:"mission_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	ChannelID              *string                    `json:"channel_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	ChannelName            *string                    `json:"channel_name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"admin,owner"`
	AIgentThreadID         *string                    `json:"aigent_thread_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	SenderName             *string                    `json:"sender_name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"admin,owner"`
	SenderConversationRole *ConversationRole          `json:"sender_conversation_role" validate:"required,oneof=user assistant system" writexs:"system,admin,owner" readxs:"admin,owner"`
	MetaData               *map[string]any            `json:"meta_data" writexs:"system,admin,owner" readxs:"admin,owner"`
	CompletionParameters   *map[string]any            `json:"parameters" writexs:"system,admin,owner" readxs:"admin,owner"`
	ChatCompletionConfig   *map[string]any            `json:"chat_completion_config" writexs:"system,admin,owner" readxs:"admin,owner"`
	Content                *AIgencyMessageContentList `json:"content" validate:"" writexs:"system,admin,owner" readxs:"admin,owner"`
	Attachments            *AIgencyMessageFileList    `json:"attachments" validate:"omitempty,dive" writexs:"system,admin,owner" readxs:"admin,owner"`
	AIServiceID            *string                    `json:"ai_service_id" validate:"required,min=1,max=64" writexs:"system,admin" readxs:"admin"`
	AIModelID              *string                    `json:"ai_model_id" validate:"required,min=1,max=64" writexs:"system,admin" readxs:"admin"`
} //@name AIgencyMessageWriteDto

type AIgencyMessage struct {
	ID                     string                     `json:"id" validate:"required,min=1,max=64" writexs:"system" readxs:"admin,owner"`
	Type                   AIgencyMessageType         `json:"type" validate:"required,oneof=message stateUpdate delta" writexs:"system,admin,owner" readxs:"admin,owner"`
	ReferenceID            string                     `json:"reference_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	ResponseToID           string                     `json:"response_to_id" validate:"omitempty,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	MissionID              string                     `json:"mission_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	ChannelID              string                     `json:"channel_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	ChannelName            string                     `json:"channel_name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"admin,owner"`
	AIgentThreadID         string                     `json:"aigent_thread_id" validate:"required,min=1,max=64" writexs:"system,admin,owner" readxs:"admin,owner"`
	SenderID               string                     `json:"sender_id" validate:"required,min=1,max=64" writexs:"system" readxs:"admin,owner"`
	SenderName             string                     `json:"sender_name" validate:"required,min=1,max=255" writexs:"system,admin,owner" readxs:"admin,owner"`
	SenderConversationRole ConversationRole           `json:"sender_conversation_role" validate:"required,oneof=user assistant system" writexs:"system,admin,owner" readxs:"admin,owner"`
	OwnerOrganizationId    string                     `json:"owner_organization_id" validate:"required,min=1,max=64" writexs:"system" readxs:"admin,owner"`
	MetaData               map[string]any             `json:"meta_data" writexs:"system,admin,owner" readxs:"admin,owner"`
	CompletionParameters   map[string]any             `json:"parameters" writexs:"system,admin,owner" readxs:"admin,owner"`
	ChatCompletionConfig   map[string]any             `json:"chat_completion_config" writexs:"system,admin,owner" readxs:"admin,owner"`
	Content                *AIgencyMessageContentList `json:"content" validate:"" writexs:"system,admin,owner" readxs:"admin,owner"`
	Attachments            *AIgencyMessageFileList    `json:"attachments" validate:"omitempty,dive" writexs:"system,admin,owner" readxs:"admin,owner"`
	TokenCount             int                        `json:"token_count" writexs:"system" readxs:"admin,owner"`
	TokenDirection         TokenDirection             `json:"token_direction" validate:"required,oneof=inbound outbound" writexs:"system" readxs:"admin,owner"`
	AIServiceID            string                     `json:"ai_service_id" validate:"required,min=1,max=64" writexs:"system,admin" readxs:"admin"`
	AIModelID              string                     `json:"ai_model_id" validate:"required,min=1,max=64" writexs:"system,admin" readxs:"admin"`
	CreatedAt              int64                      `json:"created_at" writexs:"system" readxs:"admin,owner"`
	UpdatedAt              int64                      `json:"updated_at" writexs:"system" readxs:"admin,owner"`
	ErrorMessage           string                     `json:"error_message" writexs:"system" readxs:"admin,owner"`
	CreatedForFeature      AIModelFeature             `json:"created_for_feature" writexs:"system" readxs:"admin,owner"`
	// MessageCost            float64                 `json:"message_cost" writexs:"system" readxs:"admin,owner"`
	// TokenCost              float64                 `json:"token_cost" writexs:"system" readxs:"admin,owner"`
} //@name AIgencyMessage

func NewAIgencyMessage() *AIgencyMessage {
	ID := CreateAIgencyMessageID()
	now := nuts.TimeToJSTimestamp(time.Now())
	entity := AIgencyMessage{
		ID:                   ID,
		MetaData:             make(map[string]any),
		CompletionParameters: make(map[string]any),
		ChatCompletionConfig: make(map[string]any),
		Attachments:          NewAIgencyMessageFileList(),
		Content:              NewAIgencyMessageContentList(),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	return &entity
}

func CreateAIgencyMessageID() string {
	return nuts.NID(IDPREFIX_AIGENCYMESSAGE, IDLENGTH_AIGENCYMESSAGE)
}

func IsAIgencyMessageID(id string) bool {
	return len(id) == IDLENGTH_AIGENCYMESSAGE+len(IDPREFIX_AIGENCYMESSAGE)+1 && strings.HasPrefix(id, IDPREFIX_AIGENCYMESSAGE)
}

func ValidateAIgencyMessage(entity *AIgencyMessage) error {
	var validate = validator.New()
	err := validate.Struct(entity)
	if err != nil {
		return err
	}
	return nil
}

func (msg *AIgencyMessage) CalculateTokenCount() (tokenCount int) {
	msg.TokenCount = msg.Content.TokenCount()
	return msg.TokenCount
}

type AIgencyMessageContentType string //@name AIgencyMessageContentType

const (
	AIgencyMessageContentTypeText AIgencyMessageContentType = "text"
	AIgencyMessageContentTypeFile AIgencyMessageContentType = "file"
)

type AIgencyMessageContent struct {
	ContentType AIgencyMessageContentType `json:"type" writexs:"system,admin,owner" readxs:"admin,owner,org"`
	Text        string                    `json:"text" validate:"" writexs:"system,admin,owner" readxs:"admin,owner,org"`
	File        *AIgencyMessageFile       `json:"file" validate:"" writexs:"system,admin,owner" readxs:"admin,owner,org"`
} //@name AIgencyMessageContent

func NewAIgencyMessageContent(contentType AIgencyMessageContentType, text *string, file *AIgencyMessageFile) *AIgencyMessageContent {
	entity := AIgencyMessageContent{
		Text:        "",
		File:        nil,
		ContentType: contentType,
	}
	entity.ContentType = contentType
	if text != nil {
		entity.Text = *text
	}
	if contentType == AIgencyMessageContentTypeFile && file != nil {
		entity.File = file
	}
	return &entity
}

type AIgencyMessageContentList struct {
	Data            []*AIgencyMessageContent `json:"data" writexs:"system,admin,owner" readxs:"admin,owner,org"`
	FullText        string                   `json:"full_text" writexs:"system,admin,owner" readxs:"admin,owner,org"`
	separatorString string                   `json:"-"`
	Safety          sync.Mutex               `json:"-"`
} //@name AIgencyMessageContentList

func NewAIgencyMessageContentList() *AIgencyMessageContentList {
	return &AIgencyMessageContentList{
		Data:            []*AIgencyMessageContent{},
		FullText:        "",
		separatorString: "\n",
		Safety:          sync.Mutex{},
	}
}

func (list *AIgencyMessageContentList) SetSeparatorString(separator string) {
	list.Safety.Lock()
	list.separatorString = separator
	list.Safety.Unlock()
}

func (list *AIgencyMessageContentList) GetSeparatorString() (separator string) {
	list.Safety.Lock()
	defer list.Safety.Unlock()
	return list.separatorString
}

func (list *AIgencyMessageContentList) AddContent(content *AIgencyMessageContent) {
	list.Safety.Lock()
	list.Data = append(list.Data, content)
	list.Safety.Unlock()
	list.FullText = list.GetConcatenatedText(true, false)
}

func (list *AIgencyMessageContentList) GetData() []*AIgencyMessageContent {
	listCopy := make([]*AIgencyMessageContent, len(list.Data))
	list.Safety.Lock()
	copy(listCopy, list.Data)
	list.Safety.Unlock()
	return listCopy
}

func (list *AIgencyMessageContentList) GetContentByType(contentType AIgencyMessageContentType) (contents []*AIgencyMessageContent) {
	list.Safety.Lock()
	for _, content := range list.Data {
		if content.ContentType == contentType {
			contents = append(contents, content)
		}
	}
	list.Safety.Unlock()
	return contents
}

func (list *AIgencyMessageContentList) GetConcatenatedText(includeAllContentTypes bool, includeEmptyTexts bool) (text string) {
	list.Safety.Lock()
	for _, content := range list.Data {
		if content.ContentType == AIgencyMessageContentTypeText && (includeAllContentTypes || content.Text != "" || includeEmptyTexts) {
			text += content.Text + list.separatorString
		} else if content.ContentType == AIgencyMessageContentTypeFile && (includeAllContentTypes || content.File != nil) {
			text += content.File.FileName + list.separatorString
		}
	}
	list.Safety.Unlock()
	// remove the last separator
	if len(text) > 0 {
		text = text[:len(text)-len(list.separatorString)]
	}
	return text
}

func (list *AIgencyMessageContentList) GetAIgencyFileList() (files *AIgencyMessageFileList) {
	files = NewAIgencyMessageFileList()
	list.Safety.Lock()
	for _, content := range list.Data {
		if content.ContentType == AIgencyMessageContentTypeFile && content.File != nil {
			files.AddFile(content.File)
		}
	}
	list.Safety.Unlock()
	return files
}

func (list *AIgencyMessageContentList) DeleteContentByType(contentType AIgencyMessageContentType) {
	newData := []*AIgencyMessageContent{}
	list.Safety.Lock()
	for _, content := range list.Data {
		if content.ContentType != contentType {
			newData = append(newData, content)
		}
	}
	list.Data = newData
	list.Safety.Unlock()
	list.FullText = list.GetConcatenatedText(true, false)
}

func (list *AIgencyMessageContentList) DeleteContentByIndex(index int) {
	list.Safety.Lock()
	if index >= 0 && index < len(list.Data) {
		list.Data = append(list.Data[:index], list.Data[index+1:]...)
	}
	list.Safety.Unlock()
	list.FullText = list.GetConcatenatedText(true, false)
}

func (list *AIgencyMessageContentList) ClearData() {
	list.Safety.Lock()
	list.Data = []*AIgencyMessageContent{}
	list.FullText = ""
	list.Safety.Unlock()
}

func (list *AIgencyMessageContentList) GetContentByIndex(index int) *AIgencyMessageContent {
	list.Safety.Lock()
	defer list.Safety.Unlock()
	if index >= 0 && index < len(list.Data) {
		return list.Data[index]
	}
	return nil
}

func (list *AIgencyMessageContentList) IsEmpty() bool {
	list.Safety.Lock()
	defer list.Safety.Unlock()
	return len(list.Data) == 0
}

func (list *AIgencyMessageContentList) Count() int {
	list.Safety.Lock()
	defer list.Safety.Unlock()
	return len(list.Data)
}

func (list *AIgencyMessageContentList) TextLength() (textLength int) {
	allTxt := list.GetConcatenatedText(true, false)
	textLength = len(allTxt)
	return textLength
}

func (list *AIgencyMessageContentList) TokenCount() (tokenCount int) {
	allTxt := list.GetConcatenatedText(true, false)
	tokenCount = CalculateTokenCount(allTxt)
	return tokenCount
}

func CalculateTokenCount(content string) (tokenCount int) {
	var err error
	if content == "" {
		return 0
	}
	if DefaultTextTokenEncoder == nil {
		DefaultTextTokenEncoder, err = tiktoken.EncodingForModel(DefaultTextTokenEncoderModel)
		if err != nil {
			nuts.L.Debugf("Error getting default tokenizer: (%v)", err)
		}
	}
	if DefaultTextTokenEncoder == nil {
		wordCount, nonWordCharCount, err2 := countWordsAndNonWordChars(content)
		if err2 != nil {
			l := math.Round(float64(len(content)) / 4)
			return int(l)
		}
		tokenCount = wordCount + nonWordCharCount
		return tokenCount
	}
	ids := DefaultTextTokenEncoder.Encode(content, nil, nil)
	tokenCount = len(ids)
	return tokenCount
}

func countWordsAndNonWordChars(text string) (wordCount int, nonWordCharCount int, err error) {
	wordRegex, err := regexp.Compile(`\b\w+\b`)
	if err != nil {
		return 0, 0, err
	}

	words := wordRegex.FindAllString(text, -1)
	wordCount = len(words)

	nonWordCharRegex, err := regexp.Compile(`\W`)
	if err != nil {
		return wordCount, 0, err
	}

	nonWordChars := nonWordCharRegex.FindAllString(text, -1)
	nonWordCharCount = len(nonWordChars)

	return wordCount, nonWordCharCount, nil
}
