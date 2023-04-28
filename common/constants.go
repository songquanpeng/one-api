package common

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

var StartTime = time.Now().Unix() // unit: second
var Version = "v0.0.0"            // this hard coding will be replaced automatically when building, no need to manually change
var SystemName = "One API"
var ServerAddress = "http://localhost:3000"
var Footer = ""
var TopUpLink = ""

var UsingSQLite = false

// Any options with "Secret", "Token" in its key won't be return by GetOptions

var SessionSecret = uuid.New().String()
var SQLitePath = "one-api.db"

var OptionMap map[string]string
var OptionMapRWMutex sync.RWMutex

var ItemsPerPage = 10

var PasswordLoginEnabled = true
var PasswordRegisterEnabled = true
var EmailVerificationEnabled = false
var GitHubOAuthEnabled = false
var WeChatAuthEnabled = false
var TurnstileCheckEnabled = false
var RegisterEnabled = true

var SMTPServer = ""
var SMTPAccount = ""
var SMTPToken = ""

var GitHubClientId = ""
var GitHubClientSecret = ""

var WeChatServerAddress = ""
var WeChatServerToken = ""
var WeChatAccountQRCodeImageURL = ""

var TurnstileSiteKey = ""
var TurnstileSecretKey = ""

var QuotaForNewUser = 100

// https://platform.openai.com/docs/models/model-endpoint-compatibility
var RatioGPT3dot5 float64 = 2
var RatioGPT4 float64 = 30
var RatioGPT4_32k float64 = 60

const (
	RoleGuestUser  = 0
	RoleCommonUser = 1
	RoleAdminUser  = 10
	RoleRootUser   = 100
)

var (
	FileUploadPermission    = RoleGuestUser
	FileDownloadPermission  = RoleGuestUser
	ImageUploadPermission   = RoleGuestUser
	ImageDownloadPermission = RoleGuestUser
)

// All duration's unit is seconds
// Shouldn't larger then RateLimitKeyExpirationDuration
var (
	GlobalApiRateLimitNum            = 180
	GlobalApiRateLimitDuration int64 = 3 * 60

	GlobalWebRateLimitNum            = 60
	GlobalWebRateLimitDuration int64 = 3 * 60

	UploadRateLimitNum            = 10
	UploadRateLimitDuration int64 = 60

	DownloadRateLimitNum            = 10
	DownloadRateLimitDuration int64 = 60

	CriticalRateLimitNum            = 20
	CriticalRateLimitDuration int64 = 20 * 60
)

var RateLimitKeyExpirationDuration = 20 * time.Minute

const (
	UserStatusEnabled  = 1 // don't use 0, 0 is the default value!
	UserStatusDisabled = 2 // also don't use 0
)

const (
	TokenStatusEnabled   = 1 // don't use 0, 0 is the default value!
	TokenStatusDisabled  = 2 // also don't use 0
	TokenStatusExpired   = 3
	TokenStatusExhausted = 4
)

const (
	RedemptionCodeStatusEnabled  = 1 // don't use 0, 0 is the default value!
	RedemptionCodeStatusDisabled = 2 // also don't use 0
	RedemptionCodeStatusUsed     = 3 // also don't use 0
)

const (
	ChannelStatusUnknown  = 0
	ChannelStatusEnabled  = 1 // don't use 0, 0 is the default value!
	ChannelStatusDisabled = 2 // also don't use 0
)

const (
	ChannelTypeUnknown   = 0
	ChannelTypeOpenAI    = 1
	ChannelTypeAPI2D     = 2
	ChannelTypeAzure     = 3
	ChannelTypeCloseAI   = 4
	ChannelTypeOpenAISB  = 5
	ChannelTypeOpenAIMax = 6
	ChannelTypeOhMyGPT   = 7
	ChannelTypeCustom    = 8
)

var ChannelBaseURLs = []string{
	"",                            // 0
	"https://api.openai.com",      // 1
	"https://openai.api2d.net",    // 2
	"",                            // 3
	"https://api.openai-asia.com", // 4
	"https://api.openai-sb.com",   // 5
	"https://api.openaimax.com",   // 6
	"https://api.ohmygpt.com",     // 7
	"",                            // 8
}
