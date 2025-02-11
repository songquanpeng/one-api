package message

import (
	"fmt"
	"github.com/songquanpeng/one-api/common/config"
)

const (
	ByAll           = "all"
	ByEmail         = "email"
	ByMessagePusher = "message_pusher"
)

func Notify(by string, title string, description string, content string) error {
	if by == ByAll {
		if err := SendEmail(title, config.RootUserEmail, content); err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}
		if err := SendMessage(title, description, content); err != nil {
			return fmt.Errorf("failed to send message: %v", err)
		}
		return nil
	}

	if by == ByEmail {
		return SendEmail(title, config.RootUserEmail, content)
	}
	if by == ByMessagePusher {
		return SendMessage(title, description, content)
	}
	return fmt.Errorf("unknown notify method: %s", by)
}
