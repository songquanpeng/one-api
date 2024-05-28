package channel

import (
	"context"
	"errors"
	"one-api/common/config"
	"one-api/common/stmp"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Email struct {
	To string
}

func NewEmail(to string) *Email {
	return &Email{
		To: to,
	}
}

func (e *Email) Name() string {
	return "Email"
}

func (e *Email) Send(_ context.Context, title, message string) error {
	to := e.To
	if to == "" {
		to = config.RootUserEmail
	}

	if config.SMTPServer == "" || config.SMTPAccount == "" || config.SMTPToken == "" || to == "" {
		return errors.New("smtp config is not set, skip send email notifier")
	}

	p := parser.NewWithExtensions(parser.CommonExtensions | parser.DefinitionLists | parser.OrderedListStart)
	doc := p.Parse([]byte(message))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	body := markdown.Render(doc, renderer)

	emailClient := stmp.NewStmp(config.SMTPServer, config.SMTPPort, config.SMTPAccount, config.SMTPToken, config.SMTPFrom)

	return emailClient.Send(to, title, string(body))
}
