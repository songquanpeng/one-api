package siftq

const (
	ChannelName       = "siftq"
	ModelName         = "MiniMax-H3"
	DefaultBaseURL    = "https://siftq.com/api/minimax/"
	CreateVideoPath   = "v2/video_generation"
	QueryVideoPath    = "v2/query/video_generation"
	CreateContextPath = "v2/h3_context_ir"
	MaxRequestBytes   = 64 << 20
)

var ModelList = []string{ModelName}
