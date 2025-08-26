package cohere

var ModelList = []string{
	"command", "command-nightly",
	"command-light", "command-light-nightly",
	"command-r", "command-r-plus",
	// Latest models
	"command-a-03-2025",
	"command-r-plus-04-2024",
	"command-r-08-2024",
}

func init() {
	num := len(ModelList)
	for i := 0; i < num; i++ {
		ModelList = append(ModelList, ModelList[i]+"-internet")
	}
}
