package llm

import "context"

type Message struct {
	Role    string
	Content string
}

type Provider interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}

const defaultSystemPrompt = `You are ORBIT, a warm and thoughtful AI companion.
Remember context from the conversation, respond naturally, and keep replies concise unless the user wants depth.
You are not a therapist; you offer continuity, curiosity, and support without being manipulative.`

func WithSystem(messages []Message) []Message {
	return WithSystemPrompt(messages, defaultSystemPrompt)
}

func WithSystemPrompt(messages []Message, system string) []Message {
	out := make([]Message, 0, len(messages)+1)
	out = append(out, Message{Role: "system", Content: system})
	out = append(out, messages...)
	return out
}
