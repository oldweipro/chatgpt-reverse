package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type APIRequest struct {
	Messages  []apiMessage `json:"messages"`
	Stream    bool         `json:"stream"`
	Model     string       `json:"model"`
	PluginIDs []string     `json:"plugin_ids"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatgptMessage struct {
	ID      uuid.UUID      `json:"id"`
	Author  chatgptAuthor  `json:"author"`
	Content chatgptContent `json:"content"`
}

type chatgptContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type chatgptAuthor struct {
	Role string `json:"role"`
}

type ChatGPTRequest struct {
	Action                     string           `json:"action"`
	Messages                   []chatgptMessage `json:"messages"`
	ParentMessageID            string           `json:"parent_message_id,omitempty"`
	ConversationID             string           `json:"conversation_id,omitempty"`
	Model                      string           `json:"model"`
	HistoryAndTrainingDisabled bool             `json:"history_and_training_disabled"`
	ArkoseToken                string           `json:"arkose_token,omitempty"`
	PluginIDs                  []string         `json:"plugin_ids,omitempty"`
}

func NewChatGPTRequest() ChatGPTRequest {
	return ChatGPTRequest{
		Action: "next",
		//ArkoseToken:                GetArkoseToken(),
		ParentMessageID:            uuid.NewString(),
		Model:                      "text-davinci-002-render-sha",
		HistoryAndTrainingDisabled: DisableHistory,
	}
}
func (c *ChatGPTRequest) AddMessage(role string, content string) {
	c.Messages = append(c.Messages, chatgptMessage{
		ID:      uuid.New(),
		Author:  chatgptAuthor{Role: role},
		Content: chatgptContent{ContentType: "text", Parts: []string{content}},
	})
}

type ChatCompletion struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}
type Msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Choice struct {
	Index        int         `json:"index"`
	Message      Msg         `json:"message"`
	FinishReason interface{} `json:"finish_reason"`
}
type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewChatCompletion(fullTest string) ChatCompletion {
	return ChatCompletion{
		ID:      "chatcmpl-QXlha2FBbmROaXhpZUFyZUF3ZXNvbWUK",
		Object:  "chat.completion",
		Created: int64(0),
		Model:   "gpt-3.5-turbo-0301",
		Usage: usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
		Choices: []Choice{
			{
				Message: Msg{
					Content: fullTest,
					Role:    "assistant",
				},
				Index: 0,
			},
		},
	}
}

type StringStruct struct {
	Text string `json:"text"`
}

type ChatGPTResponse struct {
	Message        Message     `json:"message"`
	ConversationID string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}

type Message struct {
	ID         string      `json:"id"`
	Author     Author      `json:"author"`
	CreateTime float64     `json:"create_time"`
	UpdateTime interface{} `json:"update_time"`
	Content    Content     `json:"content"`
	EndTurn    interface{} `json:"end_turn"`
	Weight     float64     `json:"weight"`
	Metadata   Metadata    `json:"metadata"`
	Recipient  string      `json:"recipient"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type Author struct {
	Role     string                 `json:"role"`
	Name     interface{}            `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Metadata struct {
	Timestamp     string         `json:"timestamp_"`
	MessageType   string         `json:"message_type"`
	FinishDetails *FinishDetails `json:"finish_details"`
	ModelSlug     string         `json:"model_slug"`
	Recipient     string         `json:"recipient"`
}

type FinishDetails struct {
	Type string `json:"type"`
	Stop string `json:"stop"`
}

type ChatCompletionChunk struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choices `json:"choices"`
}

func (chunk *ChatCompletionChunk) String() string {
	resp, _ := json.Marshal(chunk)
	return string(resp)
}

type Choices struct {
	Delta        Delta       `json:"delta"`
	Index        int         `json:"index"`
	FinishReason interface{} `json:"finish_reason"`
}

type Delta struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}

func NewChatCompletionChunk(text string) ChatCompletionChunk {
	return ChatCompletionChunk{
		ID:      "chatcmpl-QXlha2FBbmROaXhpZUFyZUF3ZXNvbWUK",
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   "gpt-3.5-turbo-0301",
		Choices: []Choices{
			{
				Index: 0,
				Delta: Delta{
					Content: text,
				},
				FinishReason: nil,
			},
		},
	}
}
func StopChunk(reason string) ChatCompletionChunk {
	return ChatCompletionChunk{
		ID:      "chatcmpl-QXlha2FBbmROaXhpZUFyZUF3ZXNvbWUK",
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   "gpt-3.5-turbo-0301",
		Choices: []Choices{
			{
				Index:        0,
				FinishReason: reason,
			},
		},
	}
}

type CommonModel struct {
	ID        uint           `json:"id" form:"id" gorm:"primarykey"` // 主键ID
	CreatedAt time.Time      `json:"createdAt" form:"createdAt"`     // 创建时间
	UpdatedAt time.Time      `json:"updatedAt" form:"updatedAt"`     // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                 // 删除时间
}

// MailAccount 结构体
type MailAccount struct {
	CommonModel
	Username                 string     `json:"username" form:"username" gorm:"column:username;comment:mail账号;"`
	NickName                 string     `json:"nickName" form:"nickName" gorm:"column:nick_name;comment:mail昵称;"`
	Remark                   string     `json:"remark" form:"remark" gorm:"column:remark;comment:mail备注;"`
	Password                 string     `json:"password" form:"password" gorm:"column:password;comment:mail密码;"`
	ClaudeSessionKey         string     `json:"claudeSessionKey" form:"claudeSessionKey" gorm:"column:claude_session_key;comment:claude SessionKey;type:longtext;"`
	ClaudeSessionKeyGetTime  *time.Time `json:"claudeSessionKeyGetTime" form:"claudeSessionKeyGetTime" gorm:"column:claude_session_key_get_time;comment:claude SessionKey 获取时间;"`
	OpenaiPassword           string     `json:"openaiPassword" form:"openaiPassword" gorm:"column:openai_password;comment:openai密码;"`
	OpenaiAccessToken        string     `json:"openaiAccessToken" form:"openaiAccessToken" gorm:"column:openai_access_token;comment:openai AccessToken;type:longtext;"`
	OpenaiPuid               string     `json:"openaiPuid" form:"openaiPuid" gorm:"column:openai_puid;comment:openai puid;type:longtext;"`
	OpenaiAccessTokenUseTime *time.Time `json:"openaiAccessTokenUseTime" form:"openaiAccessTokenUseTime" gorm:"column:openai_access_token_use_time;comment:openai AccessToken 使用时间;"`
	OpenaiAccessTokenGetTime *time.Time `json:"openaiAccessTokenGetTime" form:"openaiAccessTokenGetTime" gorm:"column:openai_access_token_get_time;comment:openai AccessToken 获取时间;"`
	OpenaiSkExpire           *time.Time `json:"openaiSkExpire" form:"openaiSkExpire" gorm:"column:openai_sk_expire;comment:openai sk 过期时间;"`
	SkUsedAt                 *time.Time `json:"skUsedAt" form:"skUsedAt" gorm:"column:sk_used_at;comment:openai sk 使用时间;"`
	OpenaiSk                 string     `json:"openaiSk" form:"openaiSk" gorm:"column:openai_sk;comment:openai密钥;"`
	OpenaiAmount             *uint      `json:"openaiAmount" form:"openaiAmount" gorm:"column:openai_amount;comment:openai余额，使用额度;"`
	OpenaiStatus             *uint      `json:"openaiStatus" form:"openaiStatus" gorm:"column:openai_status;comment:openai状态，是否1启用或0禁用2暂时不可用;"`
}

// TableName MailAccount 表名
func (MailAccount) TableName() string {
	return "mail_account"
}
