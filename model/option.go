package model

import (
	"one-api/common"
	"one-api/common/config"
	"one-api/common/logger"
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
	err := DB.Find(&options).Error
	return options, err
}

func GetOption(key string) (option Option, err error) {
	err = DB.First(&option, Option{Key: key}).Error
	return
}

func InitOptionMap() {
	config.OptionMapRWMutex.Lock()
	config.OptionMap = make(map[string]string)
	config.OptionMap["PasswordLoginEnabled"] = strconv.FormatBool(config.PasswordLoginEnabled)
	config.OptionMap["PasswordRegisterEnabled"] = strconv.FormatBool(config.PasswordRegisterEnabled)
	config.OptionMap["EmailVerificationEnabled"] = strconv.FormatBool(config.EmailVerificationEnabled)
	config.OptionMap["GitHubOAuthEnabled"] = strconv.FormatBool(config.GitHubOAuthEnabled)
	config.OptionMap["WeChatAuthEnabled"] = strconv.FormatBool(config.WeChatAuthEnabled)
	config.OptionMap["LarkAuthEnabled"] = strconv.FormatBool(config.LarkAuthEnabled)
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
	config.OptionMap["TurnstileSiteKey"] = ""
	config.OptionMap["TurnstileSecretKey"] = ""
	config.OptionMap["QuotaForNewUser"] = strconv.Itoa(config.QuotaForNewUser)
	config.OptionMap["QuotaForInviter"] = strconv.Itoa(config.QuotaForInviter)
	config.OptionMap["QuotaForInvitee"] = strconv.Itoa(config.QuotaForInvitee)
	config.OptionMap["QuotaRemindThreshold"] = strconv.Itoa(config.QuotaRemindThreshold)
	config.OptionMap["PreConsumedQuota"] = strconv.Itoa(config.PreConsumedQuota)
	config.OptionMap["GroupRatio"] = common.GroupRatio2JSONString()
	config.OptionMap["TopUpLink"] = config.TopUpLink
	config.OptionMap["ChatLink"] = config.ChatLink
	config.OptionMap["ChatLinks"] = config.ChatLinks
	config.OptionMap["QuotaPerUnit"] = strconv.FormatFloat(config.QuotaPerUnit, 'f', -1, 64)
	config.OptionMap["RetryTimes"] = strconv.Itoa(config.RetryTimes)
	config.OptionMap["RetryCooldownSeconds"] = strconv.Itoa(config.RetryCooldownSeconds)

	config.OptionMap["MjNotifyEnabled"] = strconv.FormatBool(config.MjNotifyEnabled)

	config.OptionMap["ChatCacheEnabled"] = strconv.FormatBool(config.ChatCacheEnabled)
	config.OptionMap["ChatCacheExpireMinute"] = strconv.Itoa(config.ChatCacheExpireMinute)

	config.OptionMap["ChatImageRequestProxy"] = ""

	config.OptionMapRWMutex.Unlock()
	loadOptionsFromDatabase()
}

func loadOptionsFromDatabase() {
	options, _ := AllOption()
	for _, option := range options {
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

var optionIntMap = map[string]*int{
	"SMTPPort":              &config.SMTPPort,
	"QuotaForNewUser":       &config.QuotaForNewUser,
	"QuotaForInviter":       &config.QuotaForInviter,
	"QuotaForInvitee":       &config.QuotaForInvitee,
	"QuotaRemindThreshold":  &config.QuotaRemindThreshold,
	"PreConsumedQuota":      &config.PreConsumedQuota,
	"RetryTimes":            &config.RetryTimes,
	"RetryCooldownSeconds":  &config.RetryCooldownSeconds,
	"ChatCacheExpireMinute": &config.ChatCacheExpireMinute,
}

var optionBoolMap = map[string]*bool{
	"PasswordRegisterEnabled":        &config.PasswordRegisterEnabled,
	"PasswordLoginEnabled":           &config.PasswordLoginEnabled,
	"EmailVerificationEnabled":       &config.EmailVerificationEnabled,
	"GitHubOAuthEnabled":             &config.GitHubOAuthEnabled,
	"WeChatAuthEnabled":              &config.WeChatAuthEnabled,
	"LarkAuthEnabled":                &config.LarkAuthEnabled,
	"TurnstileCheckEnabled":          &config.TurnstileCheckEnabled,
	"RegisterEnabled":                &config.RegisterEnabled,
	"EmailDomainRestrictionEnabled":  &config.EmailDomainRestrictionEnabled,
	"AutomaticDisableChannelEnabled": &config.AutomaticDisableChannelEnabled,
	"AutomaticEnableChannelEnabled":  &config.AutomaticEnableChannelEnabled,
	"ApproximateTokenEnabled":        &config.ApproximateTokenEnabled,
	"LogConsumeEnabled":              &config.LogConsumeEnabled,
	"DisplayInCurrencyEnabled":       &config.DisplayInCurrencyEnabled,
	"DisplayTokenStatEnabled":        &config.DisplayTokenStatEnabled,
	"MjNotifyEnabled":                &config.MjNotifyEnabled,
	"ChatCacheEnabled":               &config.ChatCacheEnabled,
}

var optionStringMap = map[string]*string{
	"SMTPServer":                  &config.SMTPServer,
	"SMTPAccount":                 &config.SMTPAccount,
	"SMTPFrom":                    &config.SMTPFrom,
	"SMTPToken":                   &config.SMTPToken,
	"ServerAddress":               &config.ServerAddress,
	"GitHubClientId":              &config.GitHubClientId,
	"GitHubClientSecret":          &config.GitHubClientSecret,
	"Footer":                      &config.Footer,
	"SystemName":                  &config.SystemName,
	"Logo":                        &config.Logo,
	"WeChatServerAddress":         &config.WeChatServerAddress,
	"WeChatServerToken":           &config.WeChatServerToken,
	"WeChatAccountQRCodeImageURL": &config.WeChatAccountQRCodeImageURL,
	"TurnstileSiteKey":            &config.TurnstileSiteKey,
	"TurnstileSecretKey":          &config.TurnstileSecretKey,
	"TopUpLink":                   &config.TopUpLink,
	"ChatLink":                    &config.ChatLink,
	"ChatLinks":                   &config.ChatLinks,
	"LarkClientId":                &config.LarkClientId,
	"LarkClientSecret":            &config.LarkClientSecret,
	"ChatImageRequestProxy":       &config.ChatImageRequestProxy,
}

func updateOptionMap(key string, value string) (err error) {
	config.OptionMapRWMutex.Lock()
	defer config.OptionMapRWMutex.Unlock()
	config.OptionMap[key] = value
	if ptr, ok := optionIntMap[key]; ok {
		*ptr, _ = strconv.Atoi(value)
		return
	}

	if ptr, ok := optionBoolMap[key]; ok {
		*ptr = value == "true"
		return
	}

	if ptr, ok := optionStringMap[key]; ok {
		*ptr = value
		return
	}

	switch key {
	case "EmailDomainWhitelist":
		config.EmailDomainWhitelist = strings.Split(value, ",")
	case "GroupRatio":
		err = common.UpdateGroupRatioByJSONString(value)
	case "ChannelDisableThreshold":
		config.ChannelDisableThreshold, _ = strconv.ParseFloat(value, 64)
	case "QuotaPerUnit":
		config.QuotaPerUnit, _ = strconv.ParseFloat(value, 64)
	}
	return err
}
