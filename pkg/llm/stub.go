package llm

import (
	"context"
	"fmt"

	"orbit/pkg/models"
)

type Stub struct{}

func NewStub() *Stub {
	return &Stub{}
}

func (s *Stub) Chat(ctx context.Context, messages []Message) (string, error) {
	_ = ctx

	if len(messages) >= 2 && messages[len(messages)-1].Content == "Start the conversation with your opening greeting." {
		return "Hi! I'm ORBIT — glad you're here. I've got a sense of who you are now, and I'm looking forward to growing alongside you. What's on your mind today?", nil
	}

	var last string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			last = messages[i].Content
			break
		}
	}
	if last == "" {
		last = "..."
	}
	return fmt.Sprintf("I'm ORBIT (dev mode — set OPENAI_API_KEY for real replies). You said: %s", last), nil
}

func StubGreeting(user models.User) string {
	name := user.DisplayName()
	if name == "" {
		name = "there"
	}

	lang := "your preferred language"
	if user.PreferredLanguage != nil && *user.PreferredLanguage != "" {
		lang = *user.PreferredLanguage
	}

	var why string
	if len(user.WhyHere) > 0 {
		if label, ok := whyHereLabels[user.WhyHere[0]]; ok {
			why = fmt.Sprintf(" I know you're here for %s —", label)
		}
	}

	return fmt.Sprintf(
		"Hi %s! I'm ORBIT, your companion.%s I'll talk with you in %s. How are you feeling today?",
		name, why, lang,
	)
}
