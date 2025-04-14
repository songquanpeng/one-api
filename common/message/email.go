package message

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/wneessen/go-mail"
)

// shouldAuth returns true if SMTP authentication credentials are provided
func shouldAuth() bool {
	return config.SMTPAccount != "" || config.SMTPToken != ""
}

// formatHeloName formats a string to be HELO compatible:
//   - Converts to lowercase (for English compatibility)
//   - Replaces spaces with hyphens
func formatHeloName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}

// getHeloName returns a HELO identifier combining system name and pod name
func getHeloName() string {
	// Get pod name from environment (Kubernetes sets this automatically)
	podName := os.Getenv("HOSTNAME")

	// Format the system name to be HELO compatible
	formattedName := formatHeloName(config.SystemName)

	if podName != "" {
		return fmt.Sprintf("%s-%s", formattedName, podName)
	}

	// Fallback if not running in Kubernetes
	return formattedName
}

// SendEmail sends an email with the given subject, receiver, and content
func SendEmail(subject, receiver, content string) error {
	if receiver == "" {
		return fmt.Errorf("receiver is empty")
	}

	// For compatibility
	if config.SMTPFrom == "" {
		config.SMTPFrom = config.SMTPAccount
	}

	// Get the improved HELO name
	heloName := getHeloName()

	// Create a new mail client
	client, err := mail.NewClient(config.SMTPServer,
		mail.WithPort(config.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithHELO(heloName),
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	// Configure TLS policy based on port
	switch config.SMTPPort {
	case 465:
		client.SetTLSPolicy(mail.TLSMandatory) // Implicit TLS/SSL
	case 587:
		client.SetTLSPolicy(mail.TLSOpportunistic) // STARTTLS
	default:
		// For other ports, decide what's appropriate
		client.SetTLSPolicy(mail.TLSOpportunistic) // Try STARTTLS if available
	}

	// Set authentication if credentials are provided
	if shouldAuth() {
		client.SetUsername(config.SMTPAccount)
		client.SetPassword(config.SMTPToken)
	}

	// Create a new message
	msg := mail.NewMsg(
		mail.WithNoDefaultUserAgent(),
	)

	// Extract domain from SMTPFrom for Message-ID safely
	var domain string
	atIndex := strings.LastIndex(config.SMTPFrom, "@")
	if atIndex >= 0 && atIndex < len(config.SMTPFrom)-1 {
		domain = config.SMTPFrom[atIndex+1:]
	} else {
		// Fallback if no valid domain found
		domain = "localhost"
	}

	// Generate a unique Message-ID
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("failed to generate message ID: %w", err)
	}
	messageID := fmt.Sprintf("%x@%s", buf, domain)

	// Set message headers and content
	if err := msg.FromFormat(config.SystemName, config.SMTPFrom); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Handle multiple recipients
	receivers := strings.Split(receiver, ";")
	for _, rcv := range receivers {
		rcv = strings.TrimSpace(rcv)
		if rcv != "" {
			if err := msg.AddTo(rcv); err != nil {
				logger.SysWarnf("Failed to add recipient %s: %v", rcv, err)
			}
		}
	}
	msg.SetMessageIDWithValue(messageID)
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, content)

	// Send the email
	if err = client.DialAndSend(msg); err != nil {
		// Check for "short response" error which might indicate successful delivery
		if strings.Contains(err.Error(), "short response") {
			logger.SysWarnf("short response from SMTP server, return nil instead of error: %s", err.Error())
			return nil
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
