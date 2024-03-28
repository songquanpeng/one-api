package model

import (
	"one-api/common"

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
	Model       string  `json:"model" gorm:"type:varchar(30);primaryKey" binding:"required"`
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

func (price *Price) FetchInputCurrencyPrice(rate float64) float64 {
	return price.GetInput() * rate
}

func (price *Price) FetchOutputCurrencyPrice(rate float64) float64 {
	return price.GetOutput() * rate
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
	err := DB.Delete(price).Error
	if err != nil {
		return err
	}
	return err
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
		"gpt-4":      {[]float64{15, 30}, common.ChannelTypeOpenAI},
		"gpt-4-0314": {[]float64{15, 30}, common.ChannelTypeOpenAI},
		"gpt-4-0613": {[]float64{15, 30}, common.ChannelTypeOpenAI},
		// 	$0.06 / 1K tokens	$0.12 / 1K tokens
		"gpt-4-32k":      {[]float64{30, 60}, common.ChannelTypeOpenAI},
		"gpt-4-32k-0314": {[]float64{30, 60}, common.ChannelTypeOpenAI},
		"gpt-4-32k-0613": {[]float64{30, 60}, common.ChannelTypeOpenAI},
		// 	$0.01 / 1K tokens	$0.03 / 1K tokens
		"gpt-4-preview":        {[]float64{5, 15}, common.ChannelTypeOpenAI},
		"gpt-4-1106-preview":   {[]float64{5, 15}, common.ChannelTypeOpenAI},
		"gpt-4-0125-preview":   {[]float64{5, 15}, common.ChannelTypeOpenAI},
		"gpt-4-turbo-preview":  {[]float64{5, 15}, common.ChannelTypeOpenAI},
		"gpt-4-vision-preview": {[]float64{5, 15}, common.ChannelTypeOpenAI},
		// 	$0.0005 / 1K tokens	$0.0015 / 1K tokens
		"gpt-3.5-turbo":      {[]float64{0.25, 0.75}, common.ChannelTypeOpenAI},
		"gpt-3.5-turbo-0125": {[]float64{0.25, 0.75}, common.ChannelTypeOpenAI},
		// 	$0.0015 / 1K tokens	$0.002 / 1K tokens
		"gpt-3.5-turbo-0301":     {[]float64{0.75, 1}, common.ChannelTypeOpenAI},
		"gpt-3.5-turbo-0613":     {[]float64{0.75, 1}, common.ChannelTypeOpenAI},
		"gpt-3.5-turbo-instruct": {[]float64{0.75, 1}, common.ChannelTypeOpenAI},
		// 	$0.003 / 1K tokens	$0.004 / 1K tokens
		"gpt-3.5-turbo-16k":      {[]float64{1.5, 2}, common.ChannelTypeOpenAI},
		"gpt-3.5-turbo-16k-0613": {[]float64{1.5, 2}, common.ChannelTypeOpenAI},
		// 	$0.001 / 1K tokens	$0.002 / 1K tokens
		"gpt-3.5-turbo-1106": {[]float64{0.5, 1}, common.ChannelTypeOpenAI},
		// 	$0.0020 / 1K tokens
		"davinci-002": {[]float64{1, 1}, common.ChannelTypeOpenAI},
		// 	$0.0004 / 1K tokens
		"babbage-002": {[]float64{0.2, 0.2}, common.ChannelTypeOpenAI},
		// $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
		"whisper-1": {[]float64{15, 15}, common.ChannelTypeOpenAI},
		// $0.015 / 1K characters
		"tts-1":      {[]float64{7.5, 7.5}, common.ChannelTypeOpenAI},
		"tts-1-1106": {[]float64{7.5, 7.5}, common.ChannelTypeOpenAI},
		// $0.030 / 1K characters
		"tts-1-hd":               {[]float64{15, 15}, common.ChannelTypeOpenAI},
		"tts-1-hd-1106":          {[]float64{15, 15}, common.ChannelTypeOpenAI},
		"text-embedding-ada-002": {[]float64{0.05, 0.05}, common.ChannelTypeOpenAI},
		// 	$0.00002 / 1K tokens
		"text-embedding-3-small": {[]float64{0.01, 0.01}, common.ChannelTypeOpenAI},
		// 	$0.00013 / 1K tokens
		"text-embedding-3-large": {[]float64{0.065, 0.065}, common.ChannelTypeOpenAI},
		"text-moderation-stable": {[]float64{0.1, 0.1}, common.ChannelTypeOpenAI},
		"text-moderation-latest": {[]float64{0.1, 0.1}, common.ChannelTypeOpenAI},
		// $0.016 - $0.020 / image
		"dall-e-2": {[]float64{8, 8}, common.ChannelTypeOpenAI},
		// $0.040 - $0.120 / image
		"dall-e-3": {[]float64{20, 20}, common.ChannelTypeOpenAI},

		// $0.80/million tokens $2.40/million tokens
		"claude-instant-1.2": {[]float64{0.4, 1.2}, common.ChannelTypeAnthropic},
		// $8.00/million tokens $24.00/million tokens
		"claude-2.0": {[]float64{4, 12}, common.ChannelTypeAnthropic},
		"claude-2.1": {[]float64{4, 12}, common.ChannelTypeAnthropic},
		// $15 / M $75 / M
		"claude-3-opus-20240229": {[]float64{7.5, 22.5}, common.ChannelTypeAnthropic},
		//  $3 / M $15 / M
		"claude-3-sonnet-20240229": {[]float64{1.3, 3.9}, common.ChannelTypeAnthropic},
		//  $0.25 / M $1.25 / M  0.00025$ / 1k tokens 0.00125$ / 1k tokens
		"claude-3-haiku-20240307": {[]float64{0.125, 0.625}, common.ChannelTypeAnthropic},

		// ￥0.004 / 1k tokens ￥0.008 / 1k tokens
		"ERNIE-Speed": {[]float64{0.2857, 0.5714}, common.ChannelTypeBaidu},
		// ￥0.012 / 1k tokens ￥0.012 / 1k tokens
		"ERNIE-Bot":    {[]float64{0.8572, 0.8572}, common.ChannelTypeBaidu},
		"ERNIE-3.5-8K": {[]float64{0.8572, 0.8572}, common.ChannelTypeBaidu},
		// 0.024元/千tokens 0.048元/千tokens
		"ERNIE-Bot-8k": {[]float64{1.7143, 3.4286}, common.ChannelTypeBaidu},
		// ￥0.008 / 1k tokens ￥0.008 / 1k tokens
		"ERNIE-Bot-turbo": {[]float64{0.5715, 0.5715}, common.ChannelTypeBaidu},
		// ￥0.12 / 1k tokens ￥0.12 / 1k tokens
		"ERNIE-Bot-4": {[]float64{8.572, 8.572}, common.ChannelTypeBaidu},
		"ERNIE-4.0":   {[]float64{8.572, 8.572}, common.ChannelTypeBaidu},
		// ￥0.002 / 1k tokens
		"Embedding-V1": {[]float64{0.1429, 0.1429}, common.ChannelTypeBaidu},
		// ￥0.004 / 1k tokens
		"BLOOMZ-7B": {[]float64{0.2857, 0.2857}, common.ChannelTypeBaidu},

		"PaLM-2":            {[]float64{1, 1}, common.ChannelTypePaLM},
		"gemini-pro":        {[]float64{1, 1}, common.ChannelTypeGemini},
		"gemini-pro-vision": {[]float64{1, 1}, common.ChannelTypeGemini},
		"gemini-1.0-pro":    {[]float64{1, 1}, common.ChannelTypeGemini},
		"gemini-1.5-pro":    {[]float64{1, 1}, common.ChannelTypeGemini},

		// ￥0.005 / 1k tokens
		"glm-3-turbo": {[]float64{0.3572, 0.3572}, common.ChannelTypeZhipu},
		// ￥0.1 / 1k tokens
		"glm-4":  {[]float64{7.143, 7.143}, common.ChannelTypeZhipu},
		"glm-4v": {[]float64{7.143, 7.143}, common.ChannelTypeZhipu},
		// ￥0.0005 / 1k tokens
		"embedding-2": {[]float64{0.0357, 0.0357}, common.ChannelTypeZhipu},
		// ￥0.25 / 1张图片
		"cogview-3": {[]float64{17.8571, 17.8571}, common.ChannelTypeZhipu},

		// ￥0.008 / 1k tokens
		"qwen-turbo": {[]float64{0.5715, 0.5715}, common.ChannelTypeAli},
		// ￥0.02 / 1k tokens
		"qwen-plus":   {[]float64{1.4286, 1.4286}, common.ChannelTypeAli},
		"qwen-vl-max": {[]float64{1.4286, 1.4286}, common.ChannelTypeAli},
		// 0.12元/1,000tokens
		"qwen-max":             {[]float64{8.5714, 8.5714}, common.ChannelTypeAli},
		"qwen-max-longcontext": {[]float64{8.5714, 8.5714}, common.ChannelTypeAli},
		// 0.008元/1,000tokens
		"qwen-vl-plus": {[]float64{0.5715, 0.5715}, common.ChannelTypeAli},
		// ￥0.0007 / 1k tokens
		"text-embedding-v1": {[]float64{0.05, 0.05}, common.ChannelTypeAli},

		// ￥0.018 / 1k tokens
		"SparkDesk":      {[]float64{1.2858, 1.2858}, common.ChannelTypeXunfei},
		"SparkDesk-v1.1": {[]float64{1.2858, 1.2858}, common.ChannelTypeXunfei},
		"SparkDesk-v2.1": {[]float64{1.2858, 1.2858}, common.ChannelTypeXunfei},
		"SparkDesk-v3.1": {[]float64{1.2858, 1.2858}, common.ChannelTypeXunfei},
		"SparkDesk-v3.5": {[]float64{1.2858, 1.2858}, common.ChannelTypeXunfei},

		// ¥0.012 / 1k tokens
		"360GPT_S2_V9": {[]float64{0.8572, 0.8572}, common.ChannelType360},
		// ¥0.001 / 1k tokens
		"embedding-bert-512-v1":     {[]float64{0.0715, 0.0715}, common.ChannelType360},
		"embedding_s1_v1":           {[]float64{0.0715, 0.0715}, common.ChannelType360},
		"semantic_similarity_s1_v1": {[]float64{0.0715, 0.0715}, common.ChannelType360},

		// ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
		"hunyuan": {[]float64{7.143, 7.143}, common.ChannelTypeTencent},
		// https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
		// ¥0.01 / 1k tokens
		"ChatStd": {[]float64{0.7143, 0.7143}, common.ChannelTypeTencent},
		//¥0.1 / 1k tokens
		"ChatPro": {[]float64{7.143, 7.143}, common.ChannelTypeTencent},

		"Baichuan2-Turbo":         {[]float64{0.5715, 0.5715}, common.ChannelTypeBaichuan}, // ¥0.008 / 1k tokens
		"Baichuan2-Turbo-192k":    {[]float64{1.143, 1.143}, common.ChannelTypeBaichuan},   // ¥0.016 / 1k tokens
		"Baichuan2-53B":           {[]float64{1.4286, 1.4286}, common.ChannelTypeBaichuan}, // ¥0.02 / 1k tokens
		"Baichuan-Text-Embedding": {[]float64{0.0357, 0.0357}, common.ChannelTypeBaichuan}, // ¥0.0005 / 1k tokens

		"abab5.5s-chat": {[]float64{0.3572, 0.3572}, common.ChannelTypeMiniMax},   // ¥0.005 / 1k tokens
		"abab5.5-chat":  {[]float64{1.0714, 1.0714}, common.ChannelTypeMiniMax},   // ¥0.015 / 1k tokens
		"abab6-chat":    {[]float64{14.2857, 14.2857}, common.ChannelTypeMiniMax}, // ¥0.2 / 1k tokens
		"embo-01":       {[]float64{0.0357, 0.0357}, common.ChannelTypeMiniMax},   // ¥0.0005 / 1k tokens

		"deepseek-coder": {[]float64{0.75, 0.75}, common.ChannelTypeDeepseek}, // 暂定 $0.0015 / 1K tokens
		"deepseek-chat":  {[]float64{0.75, 0.75}, common.ChannelTypeDeepseek}, // 暂定 $0.0015 / 1K tokens

		"moonshot-v1-8k":   {[]float64{0.8572, 0.8572}, common.ChannelTypeMoonshot}, // ¥0.012 / 1K tokens
		"moonshot-v1-32k":  {[]float64{1.7143, 1.7143}, common.ChannelTypeMoonshot}, // ¥0.024 / 1K tokens
		"moonshot-v1-128k": {[]float64{4.2857, 4.2857}, common.ChannelTypeMoonshot}, // ¥0.06 / 1K tokens

		"open-mistral-7b":       {[]float64{0.125, 0.125}, common.ChannelTypeMistral}, // 0.25$ / 1M tokens	0.25$ / 1M tokens  0.00025$ / 1k tokens
		"open-mixtral-8x7b":     {[]float64{0.35, 0.35}, common.ChannelTypeMistral},   // 0.7$ / 1M tokens	0.7$ / 1M tokens  0.0007$ / 1k tokens
		"mistral-small-latest":  {[]float64{1, 3}, common.ChannelTypeMistral},         // 2$ / 1M tokens	6$ / 1M tokens  0.002$ / 1k tokens
		"mistral-medium-latest": {[]float64{1.35, 4.05}, common.ChannelTypeMistral},   // 2.7$ / 1M tokens	8.1$ / 1M tokens  0.0027$ / 1k tokens
		"mistral-large-latest":  {[]float64{4, 12}, common.ChannelTypeMistral},        // 8$ / 1M tokens	24$ / 1M tokens  0.008$ / 1k tokens
		"mistral-embed":         {[]float64{0.05, 0.05}, common.ChannelTypeMistral},   // 0.1$ / 1M tokens 0.1$ / 1M tokens  0.0001$ / 1k tokens

		// $0.70/$0.80 /1M Tokens 0.0007$ / 1k tokens
		"llama2-70b-4096": {[]float64{0.35, 0.4}, common.ChannelTypeGroq},
		// $0.10/$0.10 /1M Tokens 0.0001$ / 1k tokens
		"llama2-7b-2048": {[]float64{0.05, 0.05}, common.ChannelTypeGroq},
		"gemma-7b-it":    {[]float64{0.05, 0.05}, common.ChannelTypeGroq},
		// $0.27/$0.27 /1M Tokens 0.00027$ / 1k tokens
		"mixtral-8x7b-32768": {[]float64{0.135, 0.135}, common.ChannelTypeGroq},

		// 2.5 元 / 1M tokens 0.0025 / 1k tokens
		"yi-34b-chat-0205": {[]float64{0.1786, 0.1786}, common.ChannelTypeLingyi},
		// 12 元 / 1M tokens 0.012 / 1k tokens
		"yi-34b-chat-200k": {[]float64{0.8571, 0.8571}, common.ChannelTypeLingyi},
		// 	6 元 / 1M tokens 0.006 / 1k tokens
		"yi-vl-plus": {[]float64{0.4286, 0.4286}, common.ChannelTypeLingyi},
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

	return prices
}
