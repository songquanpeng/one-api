package stmp

import (
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/utils"
	"strings"

	"github.com/wneessen/go-mail"
)

type StmpConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewStmp(host string, port int, username string, password string, from string) *StmpConfig {
	if from == "" {
		from = username
	}

	return &StmpConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *StmpConfig) Send(to, subject, body string) error {
	message := mail.NewMsg()
	message.From(s.From)
	message.To(to)
	message.Subject(subject)
	message.SetGenHeader("References", s.getReferences())
	message.SetBodyString(mail.TypeTextHTML, body)
	message.SetUserAgent(fmt.Sprintf("One API %s // https://github.com/MartialBE/one-api", config.Version))

	client, err := mail.NewClient(
		s.Host,
		mail.WithPort(s.Port),
		mail.WithUsername(s.Username),
		mail.WithPassword(s.Password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain))

	if err != nil {
		return err
	}

	switch s.Port {
	case 465:
		client.SetSSL(true)
	case 587:
		client.SetTLSPolicy(mail.TLSMandatory)
		client.SetSMTPAuth(mail.SMTPAuthLogin)
	}

	if err := client.DialAndSend(message); err != nil {
		return err
	}

	return nil
}

func (s *StmpConfig) getReferences() string {
	froms := strings.Split(s.From, "@")
	return fmt.Sprintf("<%s.%s@%s>", froms[0], utils.GetUUID(), froms[1])
}

func (s *StmpConfig) Render(to, subject, content string) error {
	body := getDefaultTemplate(content)

	return s.Send(to, subject, body)
}

func GetSystemStmp() (*StmpConfig, error) {
	if config.SMTPServer == "" || config.SMTPPort == 0 || config.SMTPAccount == "" || config.SMTPToken == "" {
		return nil, fmt.Errorf("SMTP 信息未配置")
	}

	return NewStmp(config.SMTPServer, config.SMTPPort, config.SMTPAccount, config.SMTPToken, config.SMTPFrom), nil
}

func SendPasswordResetEmail(userName, email, link string) error {
	stmp, err := GetSystemStmp()

	if err != nil {
		return err
	}

	contentTemp := `<p style="font-size: 30px">Hi <strong>%s,</strong></p>
	<p>
		您正在进行密码重置。点击下方按钮以重置密码。
	</p>
	
	<p style="text-align: center; font-size: 13px;">
		<a target="__blank" href="%s" class="button" style="color: #ffffff;">重置密码</a>
	</p>
	
	<p style="color: #858585; padding-top: 15px;">
		如果链接无法点击，请尝试点击下面的链接或将其复制到浏览器中打开<br> %s
	</p>
	<p style="color: #858585;">重置链接 %d 分钟内有效，如果不是本人操作，请忽略。</p>`

	subject := fmt.Sprintf("%s密码重置", config.SystemName)
	content := fmt.Sprintf(contentTemp, userName, link, link, common.VerificationValidMinutes)

	return stmp.Render(email, subject, content)
}

func SendVerificationCodeEmail(email, code string) error {
	stmp, err := GetSystemStmp()

	if err != nil {
		return err
	}

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

	return stmp.Render(email, subject, content)
}

func SendQuotaWarningCodeEmail(userName, email string, quota int, noMoreQuota bool) error {
	stmp, err := GetSystemStmp()

	if err != nil {
		return err
	}

	contentTemp := `<p style="font-size: 30px">Hi <strong>%s,</strong></p>
		<p>
			%s，当前剩余额度为 %d，为了不影响您的使用，请及时充值。
		</p>
		
		<p style="text-align: center; font-size: 13px;">
			<a target="__blank" href="%s" class="button" style="color: #ffffff;">点击充值</a>
		</p>
		
		<p style="color: #858585; padding-top: 15px;">
			如果链接无法点击，请尝试点击下面的链接或将其复制到浏览器中打开<br> %s
		</p>`

	subject := "您的额度即将用尽"
	if noMoreQuota {
		subject = "您的额度已用尽"
	}
	topUpLink := fmt.Sprintf("%s/topup", config.ServerAddress)

	content := fmt.Sprintf(contentTemp, userName, subject, quota, topUpLink, topUpLink)

	return stmp.Render(email, subject, content)
}
