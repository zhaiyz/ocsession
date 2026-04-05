package store

type Session struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    Created   int64  `json:"created"`   // Unix timestamp
    Updated   int64  `json:"updated"`   // Unix timestamp
    ProjectID string `json:"project_id"`
    Directory string `json:"directory"`
    
    // Extended fields from config
    Tags      []string
    Alias     string
    Notes     string
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
	Duration      int64  // 会话持续时间（秒）
}

type FilterCriteria struct {
    Tags       []string
    Project    string
    DateFrom   int64  // Unix timestamp
    DateTo     int64  // Unix timestamp
    Query      string
}
