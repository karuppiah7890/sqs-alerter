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

func main() {
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
	input := sqs.GetQueueAttributesInput{
		QueueUrl: &queueUrl,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
		},
	}

	// get ApproximateNumberOfMessages attribute using get queue attributes for SQS queue
	output, err := sqsClient.GetQueueAttributes(context.TODO(), &input)
	if err != nil {
		log.Fatalf("error occurred while getting sqs queue attributes: %v", err)
	}

	approxNumberOfMessagesStr := output.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]
	approxNumberOfMessages, err := strconv.Atoi(approxNumberOfMessagesStr)
	if err != nil {
		log.Fatalf("error occurred while parsing approximate number of messages count (%s) into integer: %v", approxNumberOfMessagesStr, err)
	}

	lastThreadTimestamp := oldState.LastThreadTimestamp

	// check existing state and current state.
	// if there's a change in state, go ahead and send alert
	if oldState.SendAlert(approxNumberOfMessages) {

		message := fmt.Sprintf("Warning alert :warning:! %d messages are present in %s in %s environment :warning:", approxNumberOfMessages, c.GetSqsQueueName(), c.GetEnvironmentName())

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
	}

	// store current state
	newState := state.State{
		QueueMessageCount:   approxNumberOfMessages,
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
