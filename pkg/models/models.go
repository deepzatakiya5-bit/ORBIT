package models

import "time"

type User struct {
	ID                  string     `json:"id"`
	Name                *string    `json:"name,omitempty"`
	Nickname            *string    `json:"nickname,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	Birthdate           *time.Time `json:"birthdate,omitempty"`
	Occupation          *string    `json:"occupation,omitempty"`
	PreferredLanguage   *string    `json:"preferred_language,omitempty"`
	Timezone            *string    `json:"timezone,omitempty"`
	CommunicationStyle  *string    `json:"communication_style,omitempty"`
	WhyHere             []string   `json:"why_here,omitempty"`
	Email               *string    `json:"email,omitempty"`
	Phone               *string    `json:"phone,omitempty"`
	DeviceID            *string    `json:"device_id,omitempty"`
	OnboardingCompleted bool       `json:"onboarding_completed"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// DisplayName returns nickname if set, otherwise name.
func (u User) DisplayName() string {
	if u.Nickname != nil && *u.Nickname != "" {
		return *u.Nickname
	}
	if u.Name != nil {
		return *u.Name
	}
	return ""
}

type OnboardingInput struct {
	Name               string   `json:"name"`
	Nickname           string   `json:"nickname"`
	Gender             string   `json:"gender"`
	Birthdate          string   `json:"birthdate"`
	Occupation         string   `json:"occupation"`
	PreferredLanguage  string   `json:"preferred_language"`
	Timezone           string   `json:"timezone"`
	CommunicationStyle string   `json:"communication_style"`
	WhyHere            []string `json:"why_here"`
	Email              string   `json:"email,omitempty"`
	Phone              string   `json:"phone,omitempty"`
	DeviceID           string   `json:"device_id,omitempty"`
}

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     *string   `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type ConversationSummary struct {
	Conversation
	MessageCount    int     `json:"message_count"`
	LastMessage     *string `json:"last_message,omitempty"`
	LastMessageRole *string `json:"last_message_role,omitempty"`
}

type CreateConversationResponse struct {
	Conversation Conversation `json:"conversation"`
	Greeting     *Message     `json:"greeting,omitempty"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type SendMessageResponse struct {
	UserMessage      Message `json:"user_message"`
	AssistantMessage Message `json:"assistant_message"`
}
