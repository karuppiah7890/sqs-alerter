package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
)

// All configuration is through environment variables

const STATE_FILE_PATH_ENV_VAR = "STATE_FILE_PATH"
const DEFAULT_STATE_FILE_PATH = "sqs-alerter-state.yaml"
const AWS_ACCESS_KEY_ID_ENV_VAR = "AWS_ACCESS_KEY_ID"
const AWS_SECRET_ACCESS_KEY_ENV_VAR = "AWS_SECRET_ACCESS_KEY"
const SQS_QUEUE_NAME_ENV_VAR = "SQS_QUEUE_NAME"
const DEFAULT_SQS_QUEUE_NAME = "SQS Queue"
const SQS_QUEUE_URL_ENV_VAR = "SQS_QUEUE_URL"
const ENVIRONMENT_NAME_ENV_VAR = "ENVIRONMENT_NAME"
const DEFAULT_ENVIRONMENT_NAME = "Production"
const SLACK_TOKEN_ENV_VAR = "SLACK_TOKEN"
const SLACK_CHANNEL_ENV_VAR = "SLACK_CHANNEL"

type Config struct {
	awsAccessKeyId     string
	awsSecretAccessKey string
	sqsQueueUrl        string
	sqsQueueName       string
	environmentName    string
	slackToken         string
	slackChannel       string
	stateFilePath      string
}

func NewConfigFromEnvVars() (*Config, error) {
	stateFilePath, err := getStateFilePath()
	if err != nil {
		return nil, fmt.Errorf("error occurred while getting state file path: %v", err)
	}

	awsAccessKeyId, err := getAwsAccessKeyId()
	if err != nil {
		return nil, fmt.Errorf("error occurred while getting AWS Access Key ID: %v", err)
	}

	awsSecretAccessKey, err := getAwsSecretAccessKey()
	if err != nil {
		return nil, fmt.Errorf("error occurred while getting AWS Secret Access Key: %v", err)
	}

	sqsQueueName := getSqsQueueName()

	sqsQueueUrl, err := getSqsQueueUrl()
	if err != nil {
		return nil, fmt.Errorf("error occurred while getting SQS Queue URL: %v", err)
	}

	environmentName := getEnvironmentName()

	slackToken, err := getSlackToken()
	if err != nil {
		return nil, fmt.Errorf("error occurred while getting slack token: %v", err)
	}

	slackChannel, err := getSlackChannel()
	if err != nil {
		return nil, fmt.Errorf("error occurred while getting slack channel: %v", err)
	}

	return &Config{
		stateFilePath:      stateFilePath,
		awsAccessKeyId:     awsAccessKeyId,
		awsSecretAccessKey: awsSecretAccessKey,
		sqsQueueName:       sqsQueueName,
		sqsQueueUrl:        sqsQueueUrl,
		environmentName:    environmentName,
		slackToken:         slackToken,
		slackChannel:       slackChannel,
	}, nil
}

// Get state file path
func getStateFilePath() (string, error) {
	stateFilePath, ok := os.LookupEnv(STATE_FILE_PATH_ENV_VAR)
	if !ok {
		stateFilePath = DEFAULT_STATE_FILE_PATH
	}

	_, err := os.Stat(stateFilePath)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("state file does not exist at path %s", stateFilePath)
		}

		return "", fmt.Errorf("could not find file info of the state file at path %s: %v", stateFilePath, err)
	}

	return stateFilePath, nil
}

// Get optional name for the SQS Queue. Default is "SQS Queue".
// This will be used in the alert messages
func getSqsQueueName() string {
	sqsQueueName, ok := os.LookupEnv(SQS_QUEUE_NAME_ENV_VAR)
	if !ok {
		return DEFAULT_SQS_QUEUE_NAME
	}

	return fmt.Sprintf("%s (SQS Queue)", sqsQueueName)
}

// Get SQS Queue URL
func getSqsQueueUrl() (string, error) {
	sqsQueueUrl, ok := os.LookupEnv(SQS_QUEUE_URL_ENV_VAR)
	if !ok {
		return "", fmt.Errorf("%s environment variable value is a required value. Please provide it", SQS_QUEUE_URL_ENV_VAR)
	}

	_, err := url.Parse(sqsQueueUrl)
	if err != nil {
		return "", fmt.Errorf("error while parsing %s environment variable value (%s): %v", SQS_QUEUE_URL_ENV_VAR, sqsQueueUrl, err)
	}

	return sqsQueueUrl, nil
}

// Get AWS Access Key ID
func getAwsAccessKeyId() (string, error) {
	awsAccessKeyId, ok := os.LookupEnv(AWS_ACCESS_KEY_ID_ENV_VAR)
	if !ok {
		return "", fmt.Errorf("%s environment variable value is a required value. Please define it", AWS_ACCESS_KEY_ID_ENV_VAR)
	}

	return awsAccessKeyId, nil
}

// Get AWS Secret Access Key
func getAwsSecretAccessKey() (string, error) {
	awsSecretAccessKey, ok := os.LookupEnv(AWS_SECRET_ACCESS_KEY_ENV_VAR)
	if !ok {
		return "", fmt.Errorf("%s environment variable value is a required value. Please define it", AWS_SECRET_ACCESS_KEY_ENV_VAR)
	}

	return awsSecretAccessKey, nil
}

// Get optional environment name for the environment where
// the services are running. Default is "Production". This name will
// be used in the alert messages
func getEnvironmentName() string {
	environmentName, ok := os.LookupEnv(ENVIRONMENT_NAME_ENV_VAR)
	if !ok {
		return DEFAULT_ENVIRONMENT_NAME
	}

	return environmentName
}

func getSlackToken() (string, error) {
	slackToken, ok := os.LookupEnv(SLACK_TOKEN_ENV_VAR)
	if !ok {
		return "", fmt.Errorf("%s environment variable is not defined and is required. Please define it", SLACK_TOKEN_ENV_VAR)
	}
	return slackToken, nil
}

func getSlackChannel() (string, error) {
	slackChannel, ok := os.LookupEnv(SLACK_CHANNEL_ENV_VAR)
	if !ok {
		return "", fmt.Errorf("%s environment variable is not defined and is required. Please define it", SLACK_CHANNEL_ENV_VAR)
	}
	return slackChannel, nil
}

func (c *Config) GetStateFilePath() string {
	return c.stateFilePath
}

func (c *Config) GetSqsQueueName() string {
	return c.sqsQueueName
}

func (c *Config) GetSqsQueueUrl() string {
	return c.sqsQueueUrl
}

func (c *Config) GetEnvironmentName() string {
	return c.environmentName
}

func (c *Config) GetSlackToken() string {
	return c.slackToken
}

func (c *Config) GetSlackChanel() string {
	return c.slackChannel
}
