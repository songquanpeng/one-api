package model

import (
	"one-api/common/config"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	TokensPriceType = "tokens"
	TimesPriceType  = "times"
	DefaultPrice    = 30.0
	DollarRate      = 0.002
	RMBRate         = 0.014
)

type Price struct {
	Model       string  `json:"model" gorm:"type:varchar(100)" binding:"required"`
	Type        string  `json:"type"  gorm:"default:'tokens'" binding:"required"`
	ChannelType int     `json:"channel_type" gorm:"default:0" binding:"gte=0"`
	Input       float64 `json:"input" gorm:"default:0" binding:"gte=0"`
	Output      float64 `json:"output" gorm:"default:0" binding:"gte=0"`
}

func GetAllPrices() ([]*Price, error) {
	var prices []*Price
	if err := DB.Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (price *Price) Update(modelName string) error {
	if err := DB.Model(price).Select("*").Where("model = ?", modelName).Updates(price).Error; err != nil {
		return err
	}

	return nil
}

func (price *Price) Insert() error {
	if err := DB.Create(price).Error; err != nil {
		return err
	}

	return nil
}

func (price *Price) GetInput() float64 {
	if price.Input <= 0 {
		return 0
	}
	return price.Input
}

func (price *Price) GetOutput() float64 {
	if price.Output <= 0 || price.Type == TimesPriceType {
		return 0
	}

	return price.Output
}

func (price *Price) FetchInputCurrencyPrice(rate float64) string {
	r := decimal.NewFromFloat(price.GetInput()).Mul(decimal.NewFromFloat(rate))
	return r.String()
}

func (price *Price) FetchOutputCurrencyPrice(rate float64) string {
	r := decimal.NewFromFloat(price.GetOutput()).Mul(decimal.NewFromFloat(rate))
	return r.String()
}

func UpdatePrices(tx *gorm.DB, models []string, prices *Price) error {
	err := tx.Model(Price{}).Where("model IN (?)", models).Select("*").Omit("model").Updates(
		Price{
			Type:        prices.Type,
			ChannelType: prices.ChannelType,
			Input:       prices.Input,
			Output:      prices.Output,
		}).Error

	return err
}

func DeletePrices(tx *gorm.DB, models []string) error {
	err := tx.Where("model IN (?)", models).Delete(&Price{}).Error

	return err
}

func InsertPrices(tx *gorm.DB, prices []*Price) error {
	err := tx.CreateInBatches(prices, 100).Error
	return err
}

func DeleteAllPrices(tx *gorm.DB) error {
	err := tx.Where("1=1").Delete(&Price{}).Error
	return err
}

func (price *Price) Delete() error {
	return DB.Where("model = ?", price.Model).Delete(&Price{}).Error
}

type ModelType struct {
	Ratio []float64
	Type  int
}

// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
func GetDefaultPrice() []*Price {
	ModelTypes := map[string]ModelType{
		// 	$0.03 / 1K tokens	$0.06 / 1K tokens
		"gpt-4":      {[]float64{15, 30}, config.ChannelTypeOpenAI},
		"gpt-4-0314": {[]float64{15, 30}, config.ChannelTypeOpenAI},
		"gpt-4-0613": {[]float64{15, 30}, config.ChannelTypeOpenAI},
		// 	$0.06 / 1K tokens	$0.12 / 1K tokens
		"gpt-4-32k":      {[]float64{30, 60}, config.ChannelTypeOpenAI},
		"gpt-4-32k-0314": {[]float64{30, 60}, config.ChannelTypeOpenAI},
		"gpt-4-32k-0613": {[]float64{30, 60}, config.ChannelTypeOpenAI},
		// 	$0.01 / 1K tokens	$0.03 / 1K tokens
		"gpt-4-preview":          {[]float64{5, 15}, config.ChannelTypeOpenAI},
		"gpt-4-turbo":            {[]float64{5, 15}, config.ChannelTypeOpenAI},
		"gpt-4-turbo-2024-04-09": {[]float64{5, 15}, config.ChannelTypeOpenAI},
		"gpt-4-1106-preview":     {[]float64{5, 15}, config.ChannelTypeOpenAI},
		"gpt-4-0125-preview":     {[]float64{5, 15}, config.ChannelTypeOpenAI},
		"gpt-4-turbo-preview":    {[]float64{5, 15}, config.ChannelTypeOpenAI},
		"gpt-4-vision-preview":   {[]float64{5, 15}, config.ChannelTypeOpenAI},
		// $0.005 / 1K tokens	$0.015 / 1K tokens
		"gpt-4o": {[]float64{2.5, 7.5}, config.ChannelTypeOpenAI},
		// 	$0.0005 / 1K tokens	$0.0015 / 1K tokens
		"gpt-3.5-turbo":      {[]float64{0.25, 0.75}, config.ChannelTypeOpenAI},
		"gpt-3.5-turbo-0125": {[]float64{0.25, 0.75}, config.ChannelTypeOpenAI},
		// 	$0.0015 / 1K tokens	$0.002 / 1K tokens
		"gpt-3.5-turbo-0301":     {[]float64{0.75, 1}, config.ChannelTypeOpenAI},
		"gpt-3.5-turbo-0613":     {[]float64{0.75, 1}, config.ChannelTypeOpenAI},
		"gpt-3.5-turbo-instruct": {[]float64{0.75, 1}, config.ChannelTypeOpenAI},
		// 	$0.003 / 1K tokens	$0.004 / 1K tokens
		"gpt-3.5-turbo-16k":      {[]float64{1.5, 2}, config.ChannelTypeOpenAI},
		"gpt-3.5-turbo-16k-0613": {[]float64{1.5, 2}, config.ChannelTypeOpenAI},
		// 	$0.001 / 1K tokens	$0.002 / 1K tokens
		"gpt-3.5-turbo-1106": {[]float64{0.5, 1}, config.ChannelTypeOpenAI},
		// 	$0.0020 / 1K tokens
		"davinci-002": {[]float64{1, 1}, config.ChannelTypeOpenAI},
		// 	$0.0004 / 1K tokens
		"babbage-002": {[]float64{0.2, 0.2}, config.ChannelTypeOpenAI},
		// $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
		"whisper-1": {[]float64{15, 15}, config.ChannelTypeOpenAI},
		// $0.015 / 1K characters
		"tts-1":      {[]float64{7.5, 7.5}, config.ChannelTypeOpenAI},
		"tts-1-1106": {[]float64{7.5, 7.5}, config.ChannelTypeOpenAI},
		// $0.030 / 1K characters
		"tts-1-hd":               {[]float64{15, 15}, config.ChannelTypeOpenAI},
		"tts-1-hd-1106":          {[]float64{15, 15}, config.ChannelTypeOpenAI},
		"text-embedding-ada-002": {[]float64{0.05, 0.05}, config.ChannelTypeOpenAI},
		// 	$0.00002 / 1K tokens
		"text-embedding-3-small": {[]float64{0.01, 0.01}, config.ChannelTypeOpenAI},
		// 	$0.00013 / 1K tokens
		"text-embedding-3-large": {[]float64{0.065, 0.065}, config.ChannelTypeOpenAI},
		"text-moderation-stable": {[]float64{0.1, 0.1}, config.ChannelTypeOpenAI},
		"text-moderation-latest": {[]float64{0.1, 0.1}, config.ChannelTypeOpenAI},
		// $0.016 - $0.020 / image
		"dall-e-2": {[]float64{8, 8}, config.ChannelTypeOpenAI},
		// $0.040 - $0.120 / image
		"dall-e-3": {[]float64{20, 20}, config.ChannelTypeOpenAI},

		// $0.80/million tokens $2.40/million tokens
		"claude-instant-1.2": {[]float64{0.4, 1.2}, config.ChannelTypeAnthropic},
		// $8.00/million tokens $24.00/million tokens
		"claude-2.0": {[]float64{4, 12}, config.ChannelTypeAnthropic},
		"claude-2.1": {[]float64{4, 12}, config.ChannelTypeAnthropic},
		// $15 / M $75 / M
		"claude-3-opus-20240229": {[]float64{7.5, 22.5}, config.ChannelTypeAnthropic},
		//  $3 / M $15 / M
		"claude-3-sonnet-20240229": {[]float64{1.3, 3.9}, config.ChannelTypeAnthropic},
		//  $0.25 / M $1.25 / M  0.00025$ / 1k tokens 0.00125$ / 1k tokens
		"claude-3-haiku-20240307": {[]float64{0.125, 0.625}, config.ChannelTypeAnthropic},

		// ￥0.004 / 1k tokens ￥0.008 / 1k tokens
		"ERNIE-Speed": {[]float64{0.2857, 0.5714}, config.ChannelTypeBaidu},
		// ￥0.012 / 1k tokens ￥0.012 / 1k tokens
		"ERNIE-Bot":    {[]float64{0.8572, 0.8572}, config.ChannelTypeBaidu},
		"ERNIE-3.5-8K": {[]float64{0.8572, 0.8572}, config.ChannelTypeBaidu},
		// 0.024元/千tokens 0.048元/千tokens
		"ERNIE-Bot-8k": {[]float64{1.7143, 3.4286}, config.ChannelTypeBaidu},
		// ￥0.008 / 1k tokens ￥0.008 / 1k tokens
		"ERNIE-Bot-turbo": {[]float64{0.5715, 0.5715}, config.ChannelTypeBaidu},
		// ￥0.12 / 1k tokens ￥0.12 / 1k tokens
		"ERNIE-Bot-4": {[]float64{8.572, 8.572}, config.ChannelTypeBaidu},
		"ERNIE-4.0":   {[]float64{8.572, 8.572}, config.ChannelTypeBaidu},
		// ￥0.002 / 1k tokens
		"Embedding-V1": {[]float64{0.1429, 0.1429}, config.ChannelTypeBaidu},
		// ￥0.004 / 1k tokens
		"BLOOMZ-7B": {[]float64{0.2857, 0.2857}, config.ChannelTypeBaidu},

		"PaLM-2": {[]float64{1, 1}, config.ChannelTypePaLM},
		// $0.50 / 1 million tokens  $1.50 / 1 million tokens
		// 0.0005$ / 1k tokens 0.0015$ / 1k tokens
		"gemini-pro":        {[]float64{0.25, 0.75}, config.ChannelTypeGemini},
		"gemini-pro-vision": {[]float64{0.25, 0.75}, config.ChannelTypeGemini},
		"gemini-1.0-pro":    {[]float64{0.25, 0.75}, config.ChannelTypeGemini},
		// $7 / 1 million tokens  $21 / 1 million tokens
		"gemini-1.5-pro":          {[]float64{1.75, 5.25}, config.ChannelTypeGemini},
		"gemini-1.5-pro-latest":   {[]float64{1.75, 5.25}, config.ChannelTypeGemini},
		"gemini-1.5-flash":        {[]float64{0.175, 0.265}, config.ChannelTypeGemini},
		"gemini-1.5-flash-latest": {[]float64{0.175, 0.265}, config.ChannelTypeGemini},
		"gemini-ultra":            {[]float64{1, 1}, config.ChannelTypeGemini},

		// ￥0.005 / 1k tokens
		"glm-3-turbo": {[]float64{0.3572, 0.3572}, config.ChannelTypeZhipu},
		// ￥0.1 / 1k tokens
		"glm-4":  {[]float64{7.143, 7.143}, config.ChannelTypeZhipu},
		"glm-4v": {[]float64{7.143, 7.143}, config.ChannelTypeZhipu},
		// ￥0.0005 / 1k tokens
		"embedding-2": {[]float64{0.0357, 0.0357}, config.ChannelTypeZhipu},
		// ￥0.25 / 1张图片
		"cogview-3": {[]float64{17.8571, 17.8571}, config.ChannelTypeZhipu},

		// ￥0.008 / 1k tokens
		"qwen-turbo": {[]float64{0.5715, 0.5715}, config.ChannelTypeAli},
		// ￥0.02 / 1k tokens
		"qwen-plus":   {[]float64{1.4286, 1.4286}, config.ChannelTypeAli},
		"qwen-vl-max": {[]float64{1.4286, 1.4286}, config.ChannelTypeAli},
		// 0.12元/1,000tokens
		"qwen-max":             {[]float64{8.5714, 8.5714}, config.ChannelTypeAli},
		"qwen-max-longcontext": {[]float64{8.5714, 8.5714}, config.ChannelTypeAli},
		// 0.008元/1,000tokens
		"qwen-vl-plus": {[]float64{0.5715, 0.5715}, config.ChannelTypeAli},
		// ￥0.0007 / 1k tokens
		"text-embedding-v1": {[]float64{0.05, 0.05}, config.ChannelTypeAli},

		// ￥0.018 / 1k tokens
		"SparkDesk":      {[]float64{1.2858, 1.2858}, config.ChannelTypeXunfei},
		"SparkDesk-v1.1": {[]float64{1.2858, 1.2858}, config.ChannelTypeXunfei},
		"SparkDesk-v2.1": {[]float64{1.2858, 1.2858}, config.ChannelTypeXunfei},
		"SparkDesk-v3.1": {[]float64{1.2858, 1.2858}, config.ChannelTypeXunfei},
		"SparkDesk-v3.5": {[]float64{1.2858, 1.2858}, config.ChannelTypeXunfei},

		// ¥0.012 / 1k tokens
		"360GPT_S2_V9": {[]float64{0.8572, 0.8572}, config.ChannelType360},
		// ¥0.001 / 1k tokens
		"embedding-bert-512-v1":     {[]float64{0.0715, 0.0715}, config.ChannelType360},
		"embedding_s1_v1":           {[]float64{0.0715, 0.0715}, config.ChannelType360},
		"semantic_similarity_s1_v1": {[]float64{0.0715, 0.0715}, config.ChannelType360},

		// ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
		"hunyuan": {[]float64{7.143, 7.143}, config.ChannelTypeTencent},
		// https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
		// ¥0.01 / 1k tokens
		"ChatStd": {[]float64{0.7143, 0.7143}, config.ChannelTypeTencent},
		//¥0.1 / 1k tokens
		"ChatPro": {[]float64{7.143, 7.143}, config.ChannelTypeTencent},

		"Baichuan2-Turbo":         {[]float64{0.5715, 0.5715}, config.ChannelTypeBaichuan}, // ¥0.008 / 1k tokens
		"Baichuan2-Turbo-192k":    {[]float64{1.143, 1.143}, config.ChannelTypeBaichuan},   // ¥0.016 / 1k tokens
		"Baichuan2-53B":           {[]float64{1.4286, 1.4286}, config.ChannelTypeBaichuan}, // ¥0.02 / 1k tokens
		"Baichuan-Text-Embedding": {[]float64{0.0357, 0.0357}, config.ChannelTypeBaichuan}, // ¥0.0005 / 1k tokens

		"abab5.5s-chat": {[]float64{0.3572, 0.3572}, config.ChannelTypeMiniMax},   // ¥0.005 / 1k tokens
		"abab5.5-chat":  {[]float64{1.0714, 1.0714}, config.ChannelTypeMiniMax},   // ¥0.015 / 1k tokens
		"abab6-chat":    {[]float64{14.2857, 14.2857}, config.ChannelTypeMiniMax}, // ¥0.2 / 1k tokens
		"embo-01":       {[]float64{0.0357, 0.0357}, config.ChannelTypeMiniMax},   // ¥0.0005 / 1k tokens

		"deepseek-coder": {[]float64{0.75, 0.75}, config.ChannelTypeDeepseek}, // 暂定 $0.0015 / 1K tokens
		"deepseek-chat":  {[]float64{0.75, 0.75}, config.ChannelTypeDeepseek}, // 暂定 $0.0015 / 1K tokens

		"moonshot-v1-8k":   {[]float64{0.8572, 0.8572}, config.ChannelTypeMoonshot}, // ¥0.012 / 1K tokens
		"moonshot-v1-32k":  {[]float64{1.7143, 1.7143}, config.ChannelTypeMoonshot}, // ¥0.024 / 1K tokens
		"moonshot-v1-128k": {[]float64{4.2857, 4.2857}, config.ChannelTypeMoonshot}, // ¥0.06 / 1K tokens

		"open-mistral-7b":       {[]float64{0.125, 0.125}, config.ChannelTypeMistral}, // 0.25$ / 1M tokens	0.25$ / 1M tokens  0.00025$ / 1k tokens
		"open-mixtral-8x7b":     {[]float64{0.35, 0.35}, config.ChannelTypeMistral},   // 0.7$ / 1M tokens	0.7$ / 1M tokens  0.0007$ / 1k tokens
		"mistral-small-latest":  {[]float64{1, 3}, config.ChannelTypeMistral},         // 2$ / 1M tokens	6$ / 1M tokens  0.002$ / 1k tokens
		"mistral-medium-latest": {[]float64{1.35, 4.05}, config.ChannelTypeMistral},   // 2.7$ / 1M tokens	8.1$ / 1M tokens  0.0027$ / 1k tokens
		"mistral-large-latest":  {[]float64{4, 12}, config.ChannelTypeMistral},        // 8$ / 1M tokens	24$ / 1M tokens  0.008$ / 1k tokens
		"mistral-embed":         {[]float64{0.05, 0.05}, config.ChannelTypeMistral},   // 0.1$ / 1M tokens 0.1$ / 1M tokens  0.0001$ / 1k tokens

		// $0.70/$0.80 /1M Tokens 0.0007$ / 1k tokens
		"llama2-70b-4096": {[]float64{0.35, 0.4}, config.ChannelTypeGroq},
		// $0.10/$0.10 /1M Tokens 0.0001$ / 1k tokens
		"llama2-7b-2048": {[]float64{0.05, 0.05}, config.ChannelTypeGroq},
		"gemma-7b-it":    {[]float64{0.05, 0.05}, config.ChannelTypeGroq},
		// $0.27/$0.27 /1M Tokens 0.00027$ / 1k tokens
		"mixtral-8x7b-32768": {[]float64{0.135, 0.135}, config.ChannelTypeGroq},

		// 2.5 元 / 1M tokens 0.0025 / 1k tokens
		"yi-34b-chat-0205": {[]float64{0.1786, 0.1786}, config.ChannelTypeLingyi},
		// 12 元 / 1M tokens 0.012 / 1k tokens
		"yi-34b-chat-200k": {[]float64{0.8571, 0.8571}, config.ChannelTypeLingyi},
		// 	6 元 / 1M tokens 0.006 / 1k tokens
		"yi-vl-plus": {[]float64{0.4286, 0.4286}, config.ChannelTypeLingyi},

		"@cf/stabilityai/stable-diffusion-xl-base-1.0": {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@cf/lykon/dreamshaper-8-lcm":                  {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@cf/bytedance/stable-diffusion-xl-lightning":  {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@cf/qwen/qwen1.5-7b-chat-awq":                 {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@cf/qwen/qwen1.5-14b-chat-awq":                {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@hf/thebloke/deepseek-coder-6.7b-base-awq":    {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@hf/google/gemma-7b-it":                       {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@hf/thebloke/llama-2-13b-chat-awq":            {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		"@cf/openai/whisper":                           {[]float64{0, 0}, config.ChannelTypeCloudflareAI},
		//$0.50 /1M TOKENS   $1.50/1M TOKENS
		"command-r": {[]float64{0.25, 0.75}, config.ChannelTypeCohere},
		//$3 /1M TOKENS   $15/1M TOKENS
		"command-r-plus": {[]float64{1.5, 7.5}, config.ChannelTypeCohere},

		// 0.065
		"sd3": {[]float64{32.5, 32.5}, config.ChannelTypeStabilityAI},
		// 0.04
		"sd3-turbo": {[]float64{20, 20}, config.ChannelTypeStabilityAI},
		// 0.03
		"stable-image-core": {[]float64{15, 15}, config.ChannelTypeStabilityAI},

		// hunyuan
		"hunyuan-lite":          {[]float64{0, 0}, config.ChannelTypeHunyuan},
		"hunyuan-standard":      {[]float64{0.3214, 0.3571}, config.ChannelTypeHunyuan},
		"hunyuan-standard-256k": {[]float64{1.0714, 4.2857}, config.ChannelTypeHunyuan},
		"hunyuan-pro":           {[]float64{2.1429, 7.1429}, config.ChannelTypeHunyuan},
	}

	var prices []*Price

	for model, modelType := range ModelTypes {
		prices = append(prices, &Price{
			Model:       model,
			Type:        TokensPriceType,
			ChannelType: modelType.Type,
			Input:       modelType.Ratio[0],
			Output:      modelType.Ratio[1],
		})
	}

	var DefaultMJPrice = map[string]float64{
		"mj_imagine":        50,
		"mj_variation":      50,
		"mj_reroll":         50,
		"mj_blend":          50,
		"mj_modal":          50,
		"mj_zoom":           50,
		"mj_shorten":        50,
		"mj_high_variation": 50,
		"mj_low_variation":  50,
		"mj_pan":            50,
		"mj_inpaint":        0,
		"mj_custom_zoom":    0,
		"mj_describe":       25,
		"mj_upscale":        25,
		"swap_face":         25,
	}

	for model, mjPrice := range DefaultMJPrice {
		prices = append(prices, &Price{
			Model:       model,
			Type:        TimesPriceType,
			ChannelType: config.ChannelTypeMidjourney,
			Input:       mjPrice,
			Output:      mjPrice,
		})
	}

	return prices
}
