package bedrock

const awsService = "bedrock"

type BedrockError struct {
	Message string `json:"message"`
}

type BedrockResponseStream struct {
	Bytes string `json:"bytes"`
}
