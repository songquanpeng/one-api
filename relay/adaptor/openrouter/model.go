package openrouter

// RequestProvider customize how your requests are routed using the provider object
// in the request body for Chat Completions and Completions.
//
// https://openrouter.ai/docs/features/provider-routing
type RequestProvider struct {
	// Order is list of provider names to try in order (e.g. ["Anthropic", "OpenAI"]). Default: empty
	Order []string `json:"order,omitempty"`
	// AllowFallbacks is whether to allow backup providers when the primary is unavailable. Default: true
	AllowFallbacks bool `json:"allow_fallbacks,omitempty"`
	// RequireParameters is only use providers that support all parameters in your request. Default: false
	RequireParameters bool `json:"require_parameters,omitempty"`
	// DataCollection is control whether to use providers that may store data ("allow" or "deny"). Default: "allow"
	DataCollection string `json:"data_collection,omitempty" binding:"omitempty,oneof=allow deny"`
	// Ignore is list of provider names to skip for this request. Default: empty
	Ignore []string `json:"ignore,omitempty"`
	// Quantizations is list of quantization levels to filter by (e.g. ["int4", "int8"]). Default: empty
	Quantizations []string `json:"quantizations,omitempty"`
	// Sort is sort providers by price or throughput (e.g. "price" or "throughput"). Default: empty
	Sort string `json:"sort,omitempty" binding:"omitempty,oneof=price throughput latency"`
}
