package slack_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/karuppiah7890/sqs-alerter/pkg/config"
	"github.com/karuppiah7890/sqs-alerter/pkg/slack"
)

// This test requires a valid slack token and a slack channel named test
func TestSendMessage(t *testing.T) {
	c, err := config.NewConfigFromEnvVars()
	if err != nil {
		t.Fatalf("error occurred while getting configuration from environment variables: %v", err)
	}

	testMessage := fmt.Sprintf("This is a test message. Time: %v", time.Now())

	_, err = slack.SendMessage(c.GetSlackToken(), "test", testMessage)
	if err != nil {
		t.Fatalf("error occurred while sending slack message: %v", err)
	}
}
