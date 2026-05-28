package llm

import (
	"fmt"
	"strings"
	"time"

	"orbit/pkg/models"
)

var communicationStyleHints = map[string]string{
	"warm_short":    "Be warm, empathetic, and keep replies short (1–3 sentences unless they ask for more).",
	"warm_long":     "Be warm and empathetic; give thoughtful, fuller replies when helpful.",
	"direct_short":  "Be clear and direct; keep replies short and practical.",
	"playful_short": "Be light and playful when appropriate; keep replies concise.",
}

var whyHereLabels = map[string]string{
	"companionship":  "companionship and someone to talk to",
	"stress_support": "support during stress or hard times",
	"goals":          "help with goals and accountability",
	"loneliness":     "feeling lonely or disconnected",
	"curiosity":      "curiosity about AI companions",
	"self_growth":    "self-reflection and personal growth",
}

func WithSystemForUser(user models.User, messages []Message) []Message {
	return WithSystemPrompt(messages, SystemPromptForUser(user))
}

func WithSystemForUserAndMemory(user models.User, messages []Message, memoryContext string) []Message {
	if strings.TrimSpace(memoryContext) == "" {
		return WithSystemForUser(user, messages)
	}
	prompt := SystemPromptForUser(user) + "\n\n" + memoryContext
	return WithSystemPrompt(messages, prompt)
}

func SystemPromptForUser(user models.User) string {
	base := `You are ORBIT, a warm and thoughtful AI companion.
Remember context from the conversation, respond naturally, and keep replies concise unless the user wants depth.
You are not a therapist; you offer continuity, curiosity, and support without being manipulative.`

	if !user.OnboardingCompleted {
		return base
	}

	var b strings.Builder
	b.WriteString(base)
	b.WriteString("\n\n## About this user\n")
	b.WriteString(ProfileSummary(user))

	if user.PreferredLanguage != nil && *user.PreferredLanguage != "" {
		b.WriteString(fmt.Sprintf("\nAlways respond in the user's preferred language: %s.", *user.PreferredLanguage))
	}
	if user.CommunicationStyle != nil {
		if hint, ok := communicationStyleHints[*user.CommunicationStyle]; ok {
			b.WriteString("\n" + hint)
		}
	}
	return b.String()
}

func ProfileSummary(user models.User) string {
	var lines []string

	if name := user.DisplayName(); name != "" {
		lines = append(lines, fmt.Sprintf("- Call them: %s", name))
	}
	if user.Name != nil && user.Nickname != nil && *user.Nickname != "" && *user.Name != *user.Nickname {
		lines = append(lines, fmt.Sprintf("- Full name: %s", *user.Name))
	}
	if user.Gender != nil {
		lines = append(lines, fmt.Sprintf("- Gender: %s", *user.Gender))
	}
	if user.Birthdate != nil {
		lines = append(lines, fmt.Sprintf("- Age: %d", AgeYears(*user.Birthdate)))
	}
	if user.Occupation != nil {
		lines = append(lines, fmt.Sprintf("- Occupation: %s", *user.Occupation))
	}
	if user.Timezone != nil {
		lines = append(lines, fmt.Sprintf("- Timezone: %s", *user.Timezone))
	}
	if len(user.WhyHere) > 0 {
		var reasons []string
		for _, r := range user.WhyHere {
			if label, ok := whyHereLabels[r]; ok {
				reasons = append(reasons, label)
			}
		}
		if len(reasons) > 0 {
			lines = append(lines, fmt.Sprintf("- They came to ORBIT for: %s", strings.Join(reasons, "; ")))
		}
	}
	if user.PreferredLanguage != nil {
		lines = append(lines, fmt.Sprintf("- Preferred language: %s", *user.PreferredLanguage))
	}
	return strings.Join(lines, "\n")
}

func GreetingPrompt(user models.User) []Message {
	prompt := SystemPromptForUser(user) + `

This is the first message in a new conversation after onboarding.
Greet the user warmly and personally using what you know about them — especially why they joined ORBIT.
Keep it to 2–4 sentences. Do not list all their profile fields back at them.`

	return []Message{
		{Role: "system", Content: prompt},
		{Role: "user", Content: "Start the conversation with your opening greeting."},
	}
}

func AgeYears(birthdate time.Time) int {
	now := time.Now()
	age := now.Year() - birthdate.Year()
	if now.YearDay() < birthdate.YearDay() {
		age--
	}
	return age
}
