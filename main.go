package main

import (
	"context"
	"fmt"
	"log"

	"github.com/karuppiah7890/sqs-alerter/pkg/config"

	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func main() {
	c, err := config.NewConfigFromEnvVars()
	if err != nil {
		log.Fatalf("error occurred while getting configuration from environment variables: %v", err)
	}

	_ = c

	// login to AWS

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

	output, err := sqsClient.GetQueueAttributes(context.TODO(), &input)
	if err != nil {
		log.Fatalf("error occurred while getting sqs queue attributes: %v", err)
	}

	fmt.Printf("%+v", output.Attributes)

	// get ApproximateNumberOfMessages attribute using get queue attributes

	// check existing state and current state
	// if there's a change in state, go ahead or else stop

	// store current state
	// send alerts

	// message := fmt.Sprintf("Warning alert :warning:! %s messages are present in %s in %s environment :warning:", c.GetSqsQueueName(), c.GetEnvironmentName())
	// // TODO: Use Mocks to test the integration with ease for different cases with unit tests
	// err = slack.SendMessage(c.GetSlackToken(), c.GetSlackChanel(), message)
	// if err != nil {
	// 	log.Fatalf("error occurred while sending slack alert message: %v", err)
	// }

}
