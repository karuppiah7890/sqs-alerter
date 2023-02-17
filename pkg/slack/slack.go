package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

func SendMessage(slackToken string, channel string, message string) (string, error) {
	api := slack.New(slackToken, slack.OptionDebug(true))

	_, timestamp, _, err := api.SendMessage(channel, slack.MsgOptionText(message, false))
	if err != nil {
		return "", fmt.Errorf("error occurred while sending slack message: %v", err)
	}

	return timestamp, nil
}

func SendMessageToThread(slackToken string, channel string, message string, threadTimestamp string) (string, error) {
	api := slack.New(slackToken, slack.OptionDebug(true))

	_, timestamp, _, err := api.SendMessage(channel, slack.MsgOptionText(message, false), slack.MsgOptionTS(threadTimestamp))
	if err != nil {
		return "", fmt.Errorf("error occurred while sending slack message: %v", err)
	}

	return timestamp, nil
}
