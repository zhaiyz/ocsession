package store

import "time"

type Message struct {
    SessionID string     `json:"session_id"`
    Content   string     `json:"content"`
    Timestamp time.Time  `json:"timestamp"`
    Role      string     `json:"role"` // user/assistant
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
    Result    string                 `json:"result,omitempty"`
}
