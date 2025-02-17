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
		var errMsgs []string
		if err := SendEmail(title, config.RootUserEmail, content); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("failed to send email: %v", err))
		}
		if err := SendMessage(title, description, content); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("failed to send message: %v", err))
		}

		if len(errMsgs) > 0 {
			return fmt.Errorf("multiple errors occurred: %v", errMsgs)
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
