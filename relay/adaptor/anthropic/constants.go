package anthropic

// ModelList contains all supported Claude models
var ModelList = []string{
	"claude-instant-1.2", "claude-2.0", "claude-2.1",
	"claude-3-haiku-20240307",
	"claude-3-5-haiku-20241022",
	"claude-3-5-haiku-latest",
	"claude-3-sonnet-20240229",
	"claude-3-opus-20240229",
	"claude-3-5-sonnet-20240620",
	"claude-3-5-sonnet-20241022",
	"claude-3-5-sonnet-latest",
}

// Performance optimization constants for streaming
const (
	// StreamBufferInitialSize is the initial buffer size for streaming scanner
	// 64KB is sufficient for most SSE events while minimizing memory usage
	StreamBufferInitialSize = 64 * 1024

	// StreamBufferMaxSize is the maximum buffer size for streaming scanner
	// 512KB allows handling large message blocks without frequent allocations
	StreamBufferMaxSize = 512 * 1024

	// StreamFlushThreshold is the data accumulation threshold before flushing
	// 4KB provides a good balance between latency and efficiency
	StreamFlushThreshold = 4 * 1024

	// MinEventDataLength is the minimum data length to check for event types
	// 50 bytes is enough to contain "data:{"type":"message_start"...}" prefix
	MinEventDataLength = 50

	// EventTypeCheckRange is the range to check for event type in data
	// 100 bytes covers most event types without scanning entire string
	EventTypeCheckRange = 100

	// MinDataPrefixLength is the minimum length for data prefix check
	// 6 bytes is the length of "data:"
	MinDataPrefixLength = 6
)
