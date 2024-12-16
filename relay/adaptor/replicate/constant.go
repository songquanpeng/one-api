package replicate

// ModelList is a list of models that can be used with Replicate.
//
// https://replicate.com/pricing
var ModelList = []string{
	// -------------------------------------
	// image model
	// -------------------------------------
	"black-forest-labs/flux-1.1-pro",
	"black-forest-labs/flux-1.1-pro-ultra",
	"black-forest-labs/flux-canny-dev",
	"black-forest-labs/flux-canny-pro",
	"black-forest-labs/flux-depth-dev",
	"black-forest-labs/flux-depth-pro",
	"black-forest-labs/flux-dev",
	"black-forest-labs/flux-dev-lora",
	"black-forest-labs/flux-fill-dev",
	"black-forest-labs/flux-fill-pro",
	"black-forest-labs/flux-pro",
	"black-forest-labs/flux-redux-dev",
	"black-forest-labs/flux-redux-schnell",
	"black-forest-labs/flux-schnell",
	"black-forest-labs/flux-schnell-lora",
	"ideogram-ai/ideogram-v2",
	"ideogram-ai/ideogram-v2-turbo",
	"recraft-ai/recraft-v3",
	"recraft-ai/recraft-v3-svg",
	"stability-ai/stable-diffusion-3",
	"stability-ai/stable-diffusion-3.5-large",
	"stability-ai/stable-diffusion-3.5-large-turbo",
	"stability-ai/stable-diffusion-3.5-medium",
	// -------------------------------------
	// language model
	// -------------------------------------
	// "ibm-granite/granite-20b-code-instruct-8k",  // TODO: implement the adaptor
	// "ibm-granite/granite-3.0-2b-instruct",  // TODO: implement the adaptor
	// "ibm-granite/granite-3.0-8b-instruct",  // TODO: implement the adaptor
	// "ibm-granite/granite-8b-code-instruct-128k",  // TODO: implement the adaptor
	// "meta/llama-2-13b",  // TODO: implement the adaptor
	// "meta/llama-2-13b-chat",  // TODO: implement the adaptor
	// "meta/llama-2-70b",  // TODO: implement the adaptor
	// "meta/llama-2-70b-chat",  // TODO: implement the adaptor
	// "meta/llama-2-7b",  // TODO: implement the adaptor
	// "meta/llama-2-7b-chat",  // TODO: implement the adaptor
	// "meta/meta-llama-3.1-405b-instruct",  // TODO: implement the adaptor
	// "meta/meta-llama-3-70b",  // TODO: implement the adaptor
	// "meta/meta-llama-3-70b-instruct",  // TODO: implement the adaptor
	// "meta/meta-llama-3-8b",  // TODO: implement the adaptor
	// "meta/meta-llama-3-8b-instruct",  // TODO: implement the adaptor
	// "mistralai/mistral-7b-instruct-v0.2",  // TODO: implement the adaptor
	// "mistralai/mistral-7b-v0.1",  // TODO: implement the adaptor
	// "mistralai/mixtral-8x7b-instruct-v0.1",  // TODO: implement the adaptor
	// -------------------------------------
	// video model
	// -------------------------------------
	// "minimax/video-01",  // TODO: implement the adaptor
}
