package model

import (
	"one-api/common"
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

func GetOption(key string) (option Option, err error) {
	err = DB.First(&option, Option{Key: key}).Error
	return
}

func InitOptionMap() {
	common.OptionMapRWMutex.Lock()
	common.OptionMap = make(map[string]string)
	common.OptionMap["PasswordLoginEnabled"] = strconv.FormatBool(common.PasswordLoginEnabled)
	common.OptionMap["PasswordRegisterEnabled"] = strconv.FormatBool(common.PasswordRegisterEnabled)
	common.OptionMap["EmailVerificationEnabled"] = strconv.FormatBool(common.EmailVerificationEnabled)
	common.OptionMap["GitHubOAuthEnabled"] = strconv.FormatBool(common.GitHubOAuthEnabled)
	common.OptionMap["WeChatAuthEnabled"] = strconv.FormatBool(common.WeChatAuthEnabled)
	common.OptionMap["TurnstileCheckEnabled"] = strconv.FormatBool(common.TurnstileCheckEnabled)
	common.OptionMap["RegisterEnabled"] = strconv.FormatBool(common.RegisterEnabled)
	common.OptionMap["AutomaticDisableChannelEnabled"] = strconv.FormatBool(common.AutomaticDisableChannelEnabled)
	common.OptionMap["AutomaticEnableChannelEnabled"] = strconv.FormatBool(common.AutomaticEnableChannelEnabled)
	common.OptionMap["ApproximateTokenEnabled"] = strconv.FormatBool(common.ApproximateTokenEnabled)
	common.OptionMap["LogConsumeEnabled"] = strconv.FormatBool(common.LogConsumeEnabled)
	common.OptionMap["DisplayInCurrencyEnabled"] = strconv.FormatBool(common.DisplayInCurrencyEnabled)
	common.OptionMap["DisplayTokenStatEnabled"] = strconv.FormatBool(common.DisplayTokenStatEnabled)
	common.OptionMap["ChannelDisableThreshold"] = strconv.FormatFloat(common.ChannelDisableThreshold, 'f', -1, 64)
	common.OptionMap["EmailDomainRestrictionEnabled"] = strconv.FormatBool(common.EmailDomainRestrictionEnabled)
	common.OptionMap["EmailDomainWhitelist"] = strings.Join(common.EmailDomainWhitelist, ",")
	common.OptionMap["SMTPServer"] = ""
	common.OptionMap["SMTPFrom"] = ""
	common.OptionMap["SMTPPort"] = strconv.Itoa(common.SMTPPort)
	common.OptionMap["SMTPAccount"] = ""
	common.OptionMap["SMTPToken"] = ""
	common.OptionMap["Notice"] = ""
	common.OptionMap["About"] = ""
	common.OptionMap["HomePageContent"] = ""
	common.OptionMap["Footer"] = common.Footer
	common.OptionMap["SystemName"] = common.SystemName
	common.OptionMap["Logo"] = common.Logo
	common.OptionMap["ServerAddress"] = ""
	common.OptionMap["GitHubClientId"] = ""
	common.OptionMap["GitHubClientSecret"] = ""
	common.OptionMap["WeChatServerAddress"] = ""
	common.OptionMap["WeChatServerToken"] = ""
	common.OptionMap["WeChatAccountQRCodeImageURL"] = ""
	common.OptionMap["TurnstileSiteKey"] = ""
	common.OptionMap["TurnstileSecretKey"] = ""
	common.OptionMap["QuotaForNewUser"] = strconv.Itoa(common.QuotaForNewUser)
	common.OptionMap["QuotaForInviter"] = strconv.Itoa(common.QuotaForInviter)
	common.OptionMap["QuotaForInvitee"] = strconv.Itoa(common.QuotaForInvitee)
	common.OptionMap["QuotaRemindThreshold"] = strconv.Itoa(common.QuotaRemindThreshold)
	common.OptionMap["PreConsumedQuota"] = strconv.Itoa(common.PreConsumedQuota)
	common.OptionMap["GroupRatio"] = common.GroupRatio2JSONString()
	common.OptionMap["TopUpLink"] = common.TopUpLink
	common.OptionMap["ChatLink"] = common.ChatLink
	common.OptionMap["QuotaPerUnit"] = strconv.FormatFloat(common.QuotaPerUnit, 'f', -1, 64)
	common.OptionMap["RetryTimes"] = strconv.Itoa(common.RetryTimes)
	common.OptionMap["RetryCooldownSeconds"] = strconv.Itoa(common.RetryCooldownSeconds)

	common.OptionMap["MjNotifyEnabled"] = strconv.FormatBool(common.MjNotifyEnabled)

	common.OptionMapRWMutex.Unlock()
	loadOptionsFromDatabase()
}

func loadOptionsFromDatabase() {
	options, _ := AllOption()
	for _, option := range options {
		err := updateOptionMap(option.Key, option.Value)
		if err != nil {
			common.SysError("failed to update option map: " + err.Error())
		}
	}
}

func SyncOptions(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		common.SysLog("syncing options from database")
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
	"SMTPPort":             &common.SMTPPort,
	"QuotaForNewUser":      &common.QuotaForNewUser,
	"QuotaForInviter":      &common.QuotaForInviter,
	"QuotaForInvitee":      &common.QuotaForInvitee,
	"QuotaRemindThreshold": &common.QuotaRemindThreshold,
	"PreConsumedQuota":     &common.PreConsumedQuota,
	"RetryTimes":           &common.RetryTimes,
	"RetryCooldownSeconds": &common.RetryCooldownSeconds,
}

var optionBoolMap = map[string]*bool{
	"PasswordRegisterEnabled":        &common.PasswordRegisterEnabled,
	"PasswordLoginEnabled":           &common.PasswordLoginEnabled,
	"EmailVerificationEnabled":       &common.EmailVerificationEnabled,
	"GitHubOAuthEnabled":             &common.GitHubOAuthEnabled,
	"WeChatAuthEnabled":              &common.WeChatAuthEnabled,
	"TurnstileCheckEnabled":          &common.TurnstileCheckEnabled,
	"RegisterEnabled":                &common.RegisterEnabled,
	"EmailDomainRestrictionEnabled":  &common.EmailDomainRestrictionEnabled,
	"AutomaticDisableChannelEnabled": &common.AutomaticDisableChannelEnabled,
	"AutomaticEnableChannelEnabled":  &common.AutomaticEnableChannelEnabled,
	"ApproximateTokenEnabled":        &common.ApproximateTokenEnabled,
	"LogConsumeEnabled":              &common.LogConsumeEnabled,
	"DisplayInCurrencyEnabled":       &common.DisplayInCurrencyEnabled,
	"DisplayTokenStatEnabled":        &common.DisplayTokenStatEnabled,
	"MjNotifyEnabled":                &common.MjNotifyEnabled,
}

var optionStringMap = map[string]*string{
	"SMTPServer":                  &common.SMTPServer,
	"SMTPAccount":                 &common.SMTPAccount,
	"SMTPFrom":                    &common.SMTPFrom,
	"SMTPToken":                   &common.SMTPToken,
	"ServerAddress":               &common.ServerAddress,
	"GitHubClientId":              &common.GitHubClientId,
	"GitHubClientSecret":          &common.GitHubClientSecret,
	"Footer":                      &common.Footer,
	"SystemName":                  &common.SystemName,
	"Logo":                        &common.Logo,
	"WeChatServerAddress":         &common.WeChatServerAddress,
	"WeChatServerToken":           &common.WeChatServerToken,
	"WeChatAccountQRCodeImageURL": &common.WeChatAccountQRCodeImageURL,
	"TurnstileSiteKey":            &common.TurnstileSiteKey,
	"TurnstileSecretKey":          &common.TurnstileSecretKey,
	"TopUpLink":                   &common.TopUpLink,
	"ChatLink":                    &common.ChatLink,
}

func updateOptionMap(key string, value string) (err error) {
	common.OptionMapRWMutex.Lock()
	defer common.OptionMapRWMutex.Unlock()
	common.OptionMap[key] = value
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
		common.EmailDomainWhitelist = strings.Split(value, ",")
	case "GroupRatio":
		err = common.UpdateGroupRatioByJSONString(value)
	case "ChannelDisableThreshold":
		common.ChannelDisableThreshold, _ = strconv.ParseFloat(value, 64)
	case "QuotaPerUnit":
		common.QuotaPerUnit, _ = strconv.ParseFloat(value, 64)
	}
	return err
}
