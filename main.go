package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/karuppiah7890/sqs-alerter/pkg/config"
	"github.com/karuppiah7890/sqs-alerter/pkg/slack"
	"github.com/karuppiah7890/sqs-alerter/pkg/state"

	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// TODO: Write tests for all of this

var version string

func main() {
	log.Printf("version: %v", version)
	c, err := config.NewConfigFromEnvVars()
	if err != nil {
		log.Fatalf("error occurred while getting configuration from environment variables: %v", err)
	}

	oldState, err := state.New(c.GetStateFilePath())
	if err != nil {
		log.Fatalf("error occurred while initializing state from state file at %s: %v", c.GetStateFilePath(), err)
	}

	awsconfig, err := awsconf.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("error occurred while loading aws configuration: %v", err)
	}

	sqsClient := sqs.NewFromConfig(awsconfig)

	queueUrl := c.GetSqsQueueUrl()
	numberOfMessages, err := getNumberOfMessagesInSqs(queueUrl, sqsClient)
	if err != nil {
		log.Fatalf("error occurred while getting number of messages in sqs queue: %v", err)
	}

	lastThreadTimestamp := oldState.LastThreadTimestamp

	// check existing state and current state.
	// if there's a change in state, go ahead and send alert
	if oldState.SendAlert(numberOfMessages) {

		message := fmt.Sprintf("Warning alert :warning:! %d messages are present in %s in %s environment :warning:", numberOfMessages, c.GetSqsQueueName(), c.GetEnvironmentName())

		if createNewThread(lastThreadTimestamp, time.Now(), c.GetNewThreadMinInterval()) {
			// TODO: Use Mocks to test the integration with ease for different cases with unit tests
			lastThreadTimestamp, err = slack.SendMessage(c.GetSlackToken(), c.GetSlackChanel(), message)
			if err != nil {
				log.Fatalf("error occurred while sending slack alert message: %v", err)
			}
		} else {
			// ignore the existing thread's new message's timestamp
			_, err = slack.SendMessageToThread(c.GetSlackToken(), c.GetSlackChanel(), message, lastThreadTimestamp)
			if err != nil {
				log.Fatalf("error occurred while sending slack alert message: %v", err)
			}
		}

		if numberOfMessages > 0 {
			messages, err := getFewMessagesFromQueue(queueUrl, sqsClient, numberOfMessages)
			if err != nil {
				log.Fatalf("error occurred while getting few messages from sqs queue: %v", err)
			}

			if len(messages) > 0 {
				slackMessage := formSlackMessageForQueueMessages(messages)
				_, err = slack.SendMessageToThread(c.GetSlackToken(), c.GetSlackChanel(), slackMessage, lastThreadTimestamp)
				if err != nil {
					log.Fatalf("error occurred while sending slack alert message: %v", err)
				}
			}
		}
	}

	// store current state
	newState := state.State{
		QueueMessageCount:   numberOfMessages,
		LastThreadTimestamp: lastThreadTimestamp,
	}

	err = newState.StoreToFile(c.GetStateFilePath())
	if err != nil {
		log.Fatalf("error occurred while storing new state to state file at %s: %v", c.GetStateFilePath(), err)
	}
}

func createNewThread(lastThreadTimestampStr string, now time.Time, newThreadMinInterval time.Duration) bool {
	if lastThreadTimestampStr == "" {
		return true
	}

	lastThreadTimestamp, err := strconv.ParseFloat(lastThreadTimestampStr, 64)
	if err != nil {
		log.Printf("error occurred while parsing last thread timestamp string value (%s) to float: %v", lastThreadTimestampStr, err)
		return true
	}

	lastThreadTime := time.Unix(int64(lastThreadTimestamp), 0)
	duration := now.Sub(lastThreadTime)

	return duration > newThreadMinInterval
}

func getNumberOfMessagesInSqs(queueUrl string, sqsClient *sqs.Client) (int, error) {
	input := sqs.GetQueueAttributesInput{
		QueueUrl: &queueUrl,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
		},
	}

	// get ApproximateNumberOfMessages attribute using get queue attributes for SQS queue
	output, err := sqsClient.GetQueueAttributes(context.TODO(), &input)
	if err != nil {
		return 0, fmt.Errorf("error occurred while getting sqs queue attributes: %v", err)
	}

	approxNumberOfMessagesStr := output.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]
	approxNumberOfMessages, err := strconv.Atoi(approxNumberOfMessagesStr)
	if err != nil {
		return 0, fmt.Errorf("error occurred while parsing approximate number of messages count (%s) into integer: %v", approxNumberOfMessagesStr, err)
	}

	return approxNumberOfMessages, nil
}

// Get the body of a few (max 10) messages from the queue
func getFewMessagesFromQueue(queueUrl string, sqsClient *sqs.Client, totalNumberOfMessages int) ([]*types.Message, error) {
	messages := make([]*types.Message, 0)

	messageCount := minOf(10, totalNumberOfMessages)

	for i := 0; messageCount > 0 && i < 10; i++ {
		input := sqs.ReceiveMessageInput{
			QueueUrl:              &queueUrl,
			MaxNumberOfMessages:   10,
			MessageAttributeNames: []string{"All"},
		}

		output, err := sqsClient.ReceiveMessage(context.TODO(), &input)
		if err != nil {
			return nil, fmt.Errorf("error occurred while receiving sqs queue message: %v", err)
		}

		for _, message := range output.Messages {
			messages = append(messages, &message)
		}

		messageCount -= len(output.Messages)
	}

	return messages, nil
}

func minOf(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

func formSlackMessageForQueueMessages(messages []*types.Message) string {
	slackMessage := "Below are the bodies of a few messages:"

	for _, message := range messages {
		slackMessage = fmt.Sprintf("%s\n\n```%s\n\nReason: %s```", slackMessage, *message.Body, getErrorReason(*message))
	}

	return slackMessage
}

func getErrorReason(message types.Message) string {
	sqsErrorMessage, ok := message.MessageAttributes["SQS-ErrorMessage"]

	if !ok {
		return ""
	}

	return *sqsErrorMessage.StringValue
}
