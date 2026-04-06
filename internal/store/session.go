package store

type Session struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Created   int64  `json:"created"`
	Updated   int64  `json:"updated"`
	ProjectID string `json:"project_id"`
	Directory string `json:"directory"`
}

type SessionDetail struct {
	Session      Session
	LastMessages []Message
	Stats        SessionStats
}

type SessionStats struct {
	TokenCount    int
	MessageCount  int
	ToolCallCount int
	Cost          float64
	Duration      int64 // 会话持续时间（秒）
}

type FilterCriteria struct {
	Project  string
	DateFrom int64
	DateTo   int64
	Query    string
}
