package model

import (
	"strconv"
	"strings"
	"time"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/ssrf"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
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
	config.OptionMap["OidcEnabled"] = strconv.FormatBool(config.OidcEnabled)
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

	// SSRF Protection Options
	config.OptionMap["SSRFProtectionEnabled"] = strconv.FormatBool(config.SSRFProtectionEnabled)
	config.OptionMap["SSRFAllowPrivateIP"] = strconv.FormatBool(config.SSRFAllowPrivateIP)
	config.OptionMap["SSRFApplyIPFilterForDomain"] = strconv.FormatBool(config.SSRFApplyIPFilterForDomain)
	config.OptionMap["SSRFDomainListMode"] = config.SSRFDomainListMode
	config.OptionMap["SSRFDomainList"] = strings.Join(config.SSRFDomainList, ",")
	config.OptionMap["SSRFIPListMode"] = config.SSRFIPListMode
	config.OptionMap["SSRFIPList"] = strings.Join(config.SSRFIPList, ",")
	config.OptionMap["SSRFAllowedPorts"] = intSliceToString(config.SSRFAllowedPorts)

	config.OptionMapRWMutex.Unlock()
	loadOptionsFromDatabase()
}

// intSliceToString 将整数切片转换为逗号分隔的字符串
func intSliceToString(slice []int) string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = strconv.Itoa(v)
	}
	return strings.Join(strSlice, ",")
}

// stringToIntSlice 将逗号分隔的字符串转换为整数切片
func stringToIntSlice(s string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if v, err := strconv.Atoi(part); err == nil {
			result = append(result, v)
		}
	}
	return result
}

// splitStringList 将逗号分隔的字符串转换为字符串切片
func splitStringList(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func updateSSRFProtectionConfig() {
	ssrfConfig := &ssrf.SSRFProtectionConfig{
		Enabled:                config.SSRFProtectionEnabled,
		AllowPrivateIP:         config.SSRFAllowPrivateIP,
		ApplyIPFilterForDomain: config.SSRFApplyIPFilterForDomain,
		DomainListMode:         ssrf.ListMode(config.SSRFDomainListMode),
		DomainList:             config.SSRFDomainList,
		IPListMode:             ssrf.ListMode(config.SSRFIPListMode),
		IPList:                 config.SSRFIPList,
		AllowedPorts:           config.SSRFAllowedPorts,
		DNSCacheTTL:            60,
		MaxRedirects:           10,
	}
	ssrf.UpdateGlobalConfig(ssrfConfig)
	client.CloseUserContentIdleConnections()
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
		case "OidcEnabled":
			config.OidcEnabled = boolValue
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
	case "OidcClientId":
		config.OidcClientId = value
	case "OidcClientSecret":
		config.OidcClientSecret = value
	case "OidcWellKnown":
		config.OidcWellKnown = value
	case "OidcAuthorizationEndpoint":
		config.OidcAuthorizationEndpoint = value
	case "OidcTokenEndpoint":
		config.OidcTokenEndpoint = value
	case "OidcUserinfoEndpoint":
		config.OidcUserinfoEndpoint = value
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
	// SSRF Protection Options
	case "SSRFProtectionEnabled":
		config.SSRFProtectionEnabled = value == "true"
		updateSSRFProtectionConfig()
	case "SSRFAllowPrivateIP":
		config.SSRFAllowPrivateIP = value == "true"
		updateSSRFProtectionConfig()
	case "SSRFApplyIPFilterForDomain":
		config.SSRFApplyIPFilterForDomain = value == "true"
		updateSSRFProtectionConfig()
	case "SSRFDomainListMode":
		if value == "whitelist" || value == "blacklist" {
			config.SSRFDomainListMode = value
		} else {
			config.SSRFDomainListMode = "blacklist"
		}
		updateSSRFProtectionConfig()
	case "SSRFDomainList":
		config.SSRFDomainList = splitStringList(value)
		updateSSRFProtectionConfig()
	case "SSRFIPListMode":
		if value == "whitelist" || value == "blacklist" {
			config.SSRFIPListMode = value
		} else {
			config.SSRFIPListMode = "blacklist"
		}
		updateSSRFProtectionConfig()
	case "SSRFIPList":
		config.SSRFIPList = splitStringList(value)
		updateSSRFProtectionConfig()
	case "SSRFAllowedPorts":
		config.SSRFAllowedPorts = stringToIntSlice(value)
		updateSSRFProtectionConfig()
	}
	return err
}
