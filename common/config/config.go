package config

import (
	"github.com/songquanpeng/one-api/common/helper"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

var SystemName = "One API"
var ServerAddress = "http://localhost:3000"
var Footer = ""
var Logo = ""
var TopUpLink = ""
var ChatLink = ""
var QuotaPerUnit = 500 * 1000.0 // $0.002 / 1K tokens
var DisplayInCurrencyEnabled = true
var DisplayTokenStatEnabled = true

// Any options with "Secret", "Token" in its key won't be return by GetOptions

var SessionSecret = uuid.New().String()

var OptionMap map[string]string
var OptionMapRWMutex sync.RWMutex

var ItemsPerPage = 10
var MaxRecentItems = 100

var PasswordLoginEnabled = true
var PasswordRegisterEnabled = true
var EmailVerificationEnabled = false
var GitHubOAuthEnabled = false
var WeChatAuthEnabled = false
var TurnstileCheckEnabled = false
var RegisterEnabled = true

var EmailDomainRestrictionEnabled = false
var EmailDomainWhitelist = []string{
	"gmail.com",
	"163.com",
	"126.com",
	"qq.com",
	"outlook.com",
	"hotmail.com",
	"icloud.com",
	"yahoo.com",
	"foxmail.com",
}

var DebugEnabled = os.Getenv("DEBUG") == "true"
var MemoryCacheEnabled = os.Getenv("MEMORY_CACHE_ENABLED") == "true"

var LogConsumeEnabled = true

var SMTPServer = ""
var SMTPPort = 587
var SMTPAccount = ""
var SMTPFrom = ""
var SMTPToken = ""

var GitHubClientId = ""
var GitHubClientSecret = ""

var WeChatServerAddress = ""
var WeChatServerToken = ""
var WeChatAccountQRCodeImageURL = ""

var TurnstileSiteKey = ""
var TurnstileSecretKey = ""

var QuotaForNewUser = 0
var QuotaForInviter = 0
var QuotaForInvitee = 0
var ChannelDisableThreshold = 5.0
var AutomaticDisableChannelEnabled = false
var AutomaticEnableChannelEnabled = false
var QuotaRemindThreshold = 1000
var PreConsumedQuota = 500
var ApproximateTokenEnabled = false
var RetryTimes = 0

var RootUserEmail = ""

var IsMasterNode = os.Getenv("NODE_TYPE") != "slave"

var requestInterval, _ = strconv.Atoi(os.Getenv("POLLING_INTERVAL"))
var RequestInterval = time.Duration(requestInterval) * time.Second

var SyncFrequency = helper.GetOrDefaultEnvInt("SYNC_FREQUENCY", 10*60) // unit is second

var BatchUpdateEnabled = false
var BatchUpdateInterval = helper.GetOrDefaultEnvInt("BATCH_UPDATE_INTERVAL", 5)

var RelayTimeout = helper.GetOrDefaultEnvInt("RELAY_TIMEOUT", 0) // unit is second

var GeminiSafetySetting = helper.GetOrDefaultEnvString("GEMINI_SAFETY_SETTING", "BLOCK_NONE")

var Theme = helper.GetOrDefaultEnvString("THEME", "default")
var ValidThemes = map[string]bool{
	"default": true,
	"berry":   true,
}

// All duration's unit is seconds
// Shouldn't larger then RateLimitKeyExpirationDuration
var (
	GlobalApiRateLimitNum            = helper.GetOrDefaultEnvInt("GLOBAL_API_RATE_LIMIT", 180)
	GlobalApiRateLimitDuration int64 = 3 * 60

	GlobalWebRateLimitNum            = helper.GetOrDefaultEnvInt("GLOBAL_WEB_RATE_LIMIT", 60)
	GlobalWebRateLimitDuration int64 = 3 * 60

	UploadRateLimitNum            = 10
	UploadRateLimitDuration int64 = 60

	DownloadRateLimitNum            = 10
	DownloadRateLimitDuration int64 = 60

	CriticalRateLimitNum            = 20
	CriticalRateLimitDuration int64 = 20 * 60
)

var RateLimitKeyExpirationDuration = 20 * time.Minute
