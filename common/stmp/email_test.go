package stmp_test

import (
	"fmt"
	"one-api/common/config"
	"testing"

	"one-api/common"
	"one-api/common/stmp"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func InitConfig() {
	viper.AddConfigPath("/one-api")
	viper.SetConfigName("test")
	viper.ReadInConfig()
}

type SMTPConfig struct {
	Name        string `mapstructure:"name"`
	SMTPServer  string `mapstructure:"SMTPServer"`
	SMTPPort    int    `mapstructure:"SMTPPort"`
	SMTPAccount string `mapstructure:"SMTPAccount"`
	SMTPToken   string `mapstructure:"SMTPToken"`
	SMTPFrom    string `mapstructure:"SMTPFrom"`
}

func TestSend(t *testing.T) {
	InitConfig()

	var configs []SMTPConfig
	err := viper.UnmarshalKey("stmp.provider", &configs)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(configs)
	email := viper.GetString("stmp.to")
	fmt.Println(email)

	for _, tt := range configs {
		t.Run(tt.Name, func(t *testing.T) {
			stmpClient := stmp.NewStmp(tt.SMTPServer, tt.SMTPPort, tt.SMTPAccount, tt.SMTPToken, tt.SMTPFrom)
			code := "123456"
			contentTemp := `
				<p>
					您正在进行邮箱验证。您的验证码为: 
				</p>
				
				<p style="text-align: center; font-size: 30px; color: #58a6ff;">
					<strong>%s</strong>
				</p>
				
				<p style="color: #858585; padding-top: 15px;">
					验证码 %d 分钟内有效，如果不是本人操作，请忽略。
				</p>`

			subject := fmt.Sprintf("%s邮箱验证邮件", config.SystemName)
			content := fmt.Sprintf(contentTemp, code, common.VerificationValidMinutes)

			err := stmpClient.Render(email, subject, content)
			assert.NoError(t, err)
		})
	}
}
