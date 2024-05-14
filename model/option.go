package model

import (
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"strconv"
	"strings"
	"time"
)

type Option struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value"`
}

func AllOption() ([]*Option, error) {
	var options []*Option
	var err error
	err = DB.Find(&options).Error
	return options, err
}

func InitOptionMap() {
	config.OptionMapRWMutex.Lock()
	config.OptionMap = make(map[string]string)
	config.OptionMap["PasswordLoginEnabled"] = strconv.FormatBool(config.PasswordLoginEnabled)
	config.OptionMap["PasswordRegisterEnabled"] = strconv.FormatBool(config.PasswordRegisterEnabled)
	config.OptionMap["EmailVerificationEnabled"] = strconv.FormatBool(config.EmailVerificationEnabled)
	config.OptionMap["GitHubOAuthEnabled"] = strconv.FormatBool(config.GitHubOAuthEnabled)
	config.OptionMap["WeChatAuthEnabled"] = strconv.FormatBool(config.WeChatAuthEnabled)
	config.OptionMap["TurnstileCheckEnabled"] = strconv.FormatBool(config.TurnstileCheckEnabled)
	config.OptionMap["RegisterEnabled"] = strconv.FormatBool(config.RegisterEnabled)
	config.OptionMap["AutomaticDisableChannelEnabled"] = strconv.FormatBool(config.AutomaticDisableChannelEnabled)
	config.OptionMap["AutomaticEnableChannelEnabled"] = strconv.FormatBool(config.AutomaticEnableChannelEnabled)
	config.OptionMap["ApproximateTokenEnabled"] = strconv.FormatBool(config.ApproximateTokenEnabled)
	config.OptionMap["LogConsumeEnabled"] = strconv.FormatBool(config.LogConsumeEnabled)
	config.OptionMap["DisplayInCurrencyEnabled"] = strconv.FormatBool(config.DisplayInCurrencyEnabled)
	config.OptionMap["DisplayTokenStatEnabled"] = strconv.FormatBool(config.DisplayTokenStatEnabled)
	config.OptionMap["ChannelDisableThreshold"] = strconv.FormatFloat(config.ChannelDisableThreshold, 'f', -1, 64)
	config.OptionMap["EmailDomainRestrictionEnabled"] = strconv.FormatBool(config.EmailDomainRestrictionEnabled)
	config.OptionMap["EmailDomainWhitelist"] = strings.Join(config.EmailDomainWhitelist, ",")
	config.OptionMap["SMTPServer"] = ""
	config.OptionMap["SMTPFrom"] = ""
	config.OptionMap["SMTPPort"] = strconv.Itoa(config.SMTPPort)
	config.OptionMap["SMTPAccount"] = ""
	config.OptionMap["SMTPToken"] = ""
	config.OptionMap["Notice"] = ""
	config.OptionMap["About"] = ""
	config.OptionMap["HomePageContent"] = ""
	config.OptionMap["Footer"] = config.Footer
	config.OptionMap["SystemName"] = config.SystemName
	config.OptionMap["Logo"] = config.Logo
	config.OptionMap["ServerAddress"] = ""
	config.OptionMap["GitHubClientId"] = ""
	config.OptionMap["GitHubClientSecret"] = ""
	config.OptionMap["WeChatServerAddress"] = ""
	config.OptionMap["WeChatServerToken"] = ""
	config.OptionMap["WeChatAccountQRCodeImageURL"] = ""
	config.OptionMap["MessagePusherAddress"] = ""
	config.OptionMap["MessagePusherToken"] = ""
	config.OptionMap["TurnstileSiteKey"] = ""
	config.OptionMap["TurnstileSecretKey"] = ""
	config.OptionMap["QuotaForNewUser"] = strconv.FormatInt(config.QuotaForNewUser, 10)
	config.OptionMap["QuotaForInviter"] = strconv.FormatInt(config.QuotaForInviter, 10)
	config.OptionMap["QuotaForInvitee"] = strconv.FormatInt(config.QuotaForInvitee, 10)
	config.OptionMap["QuotaRemindThreshold"] = strconv.FormatInt(config.QuotaRemindThreshold, 10)
	config.OptionMap["PreConsumedQuota"] = strconv.FormatInt(config.PreConsumedQuota, 10)
	config.OptionMap["ModelRatio"] = billingratio.ModelRatio2JSONString()
	config.OptionMap["GroupRatio"] = billingratio.GroupRatio2JSONString()
	config.OptionMap["CompletionRatio"] = billingratio.CompletionRatio2JSONString()
	config.OptionMap["TopUpLink"] = config.TopUpLink
	config.OptionMap["ChatLink"] = config.ChatLink
	config.OptionMap["QuotaPerUnit"] = strconv.FormatFloat(config.QuotaPerUnit, 'f', -1, 64)
	config.OptionMap["RetryTimes"] = strconv.Itoa(config.RetryTimes)
	config.OptionMap["Theme"] = config.Theme
	config.OptionMapRWMutex.Unlock()
	loadOptionsFromDatabase()
}

func loadOptionsFromDatabase() {
	options, _ := AllOption()
	for _, option := range options {
		if option.Key == "ModelRatio" {
			option.Value = billingratio.AddNewMissingRatio(option.Value)
		}
		err := updateOptionMap(option.Key, option.Value)
		if err != nil {
			logger.SysError("failed to update option map: " + err.Error())
		}
	}
}

func SyncOptions(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		logger.SysLog("syncing options from database")
		loadOptionsFromDatabase()
	}
}

func UpdateOption(key string, value string) error {
	// Save to database first
	option := Option{
		Key: key,
	}
	// https://gorm.io/docs/update.html#Save-All-Fields
	DB.FirstOrCreate(&option, Option{Key: key})
	option.Value = value
	// Save is a combination function.
	// If save value does not contain primary key, it will execute Create,
	// otherwise it will execute Update (with all fields).
	DB.Save(&option)
	// Update OptionMap
	return updateOptionMap(key, value)
}

func updateOptionMap(key string, value string) (err error) {
	config.OptionMapRWMutex.Lock()
	defer config.OptionMapRWMutex.Unlock()
	config.OptionMap[key] = value
	if strings.HasSuffix(key, "Enabled") {
		boolValue := value == "true"
		switch key {
		case "PasswordRegisterEnabled":
			config.PasswordRegisterEnabled = boolValue
		case "PasswordLoginEnabled":
			config.PasswordLoginEnabled = boolValue
		case "EmailVerificationEnabled":
			config.EmailVerificationEnabled = boolValue
		case "GitHubOAuthEnabled":
			config.GitHubOAuthEnabled = boolValue
		case "WeChatAuthEnabled":
			config.WeChatAuthEnabled = boolValue
		case "TurnstileCheckEnabled":
			config.TurnstileCheckEnabled = boolValue
		case "RegisterEnabled":
			config.RegisterEnabled = boolValue
		case "EmailDomainRestrictionEnabled":
			config.EmailDomainRestrictionEnabled = boolValue
		case "AutomaticDisableChannelEnabled":
			config.AutomaticDisableChannelEnabled = boolValue
		case "AutomaticEnableChannelEnabled":
			config.AutomaticEnableChannelEnabled = boolValue
		case "ApproximateTokenEnabled":
			config.ApproximateTokenEnabled = boolValue
		case "LogConsumeEnabled":
			config.LogConsumeEnabled = boolValue
		case "DisplayInCurrencyEnabled":
			config.DisplayInCurrencyEnabled = boolValue
		case "DisplayTokenStatEnabled":
			config.DisplayTokenStatEnabled = boolValue
		}
	}
	switch key {
	case "EmailDomainWhitelist":
		config.EmailDomainWhitelist = strings.Split(value, ",")
	case "SMTPServer":
		config.SMTPServer = value
	case "SMTPPort":
		intValue, _ := strconv.Atoi(value)
		config.SMTPPort = intValue
	case "SMTPAccount":
		config.SMTPAccount = value
	case "SMTPFrom":
		config.SMTPFrom = value
	case "SMTPToken":
		config.SMTPToken = value
	case "ServerAddress":
		config.ServerAddress = value
	case "GitHubClientId":
		config.GitHubClientId = value
	case "GitHubClientSecret":
		config.GitHubClientSecret = value
	case "LarkClientId":
		config.LarkClientId = value
	case "LarkClientSecret":
		config.LarkClientSecret = value
	case "Footer":
		config.Footer = value
	case "SystemName":
		config.SystemName = value
	case "Logo":
		config.Logo = value
	case "WeChatServerAddress":
		config.WeChatServerAddress = value
	case "WeChatServerToken":
		config.WeChatServerToken = value
	case "WeChatAccountQRCodeImageURL":
		config.WeChatAccountQRCodeImageURL = value
	case "MessagePusherAddress":
		config.MessagePusherAddress = value
	case "MessagePusherToken":
		config.MessagePusherToken = value
	case "TurnstileSiteKey":
		config.TurnstileSiteKey = value
	case "TurnstileSecretKey":
		config.TurnstileSecretKey = value
	case "QuotaForNewUser":
		config.QuotaForNewUser, _ = strconv.ParseInt(value, 10, 64)
	case "QuotaForInviter":
		config.QuotaForInviter, _ = strconv.ParseInt(value, 10, 64)
	case "QuotaForInvitee":
		config.QuotaForInvitee, _ = strconv.ParseInt(value, 10, 64)
	case "QuotaRemindThreshold":
		config.QuotaRemindThreshold, _ = strconv.ParseInt(value, 10, 64)
	case "PreConsumedQuota":
		config.PreConsumedQuota, _ = strconv.ParseInt(value, 10, 64)
	case "RetryTimes":
		config.RetryTimes, _ = strconv.Atoi(value)
	case "ModelRatio":
		err = billingratio.UpdateModelRatioByJSONString(value)
	case "GroupRatio":
		err = billingratio.UpdateGroupRatioByJSONString(value)
	case "CompletionRatio":
		err = billingratio.UpdateCompletionRatioByJSONString(value)
	case "TopUpLink":
		config.TopUpLink = value
	case "ChatLink":
		config.ChatLink = value
	case "ChannelDisableThreshold":
		config.ChannelDisableThreshold, _ = strconv.ParseFloat(value, 64)
	case "QuotaPerUnit":
		config.QuotaPerUnit, _ = strconv.ParseFloat(value, 64)
	case "Theme":
		config.Theme = value
	}
	return err
}
