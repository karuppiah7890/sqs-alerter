package state

import (
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

type State struct {
	QueueMessageCount   int `yaml:"queueMessageCount"`
	LastThreadTimestamp string
}

func New(stateFilePath string) (*State, error) {
	stateFileData, err := os.ReadFile(stateFilePath)
	if err != nil {
		return nil, fmt.Errorf("error occurred while reading state file at path %s: %v", stateFilePath, err)
	}

	var state State

	err = yaml.Unmarshal(stateFileData, &state)
	if err != nil {
		return nil, fmt.Errorf("error occurred while parsing yaml state file at path %s: %v", stateFilePath, err)
	}
	return &state, nil
}

func (old *State) SendAlert(newQueueMessageCount int) bool {
	return newQueueMessageCount != old.QueueMessageCount
}

func (s *State) StoreToFile(stateFilePath string) error {
	stateFileData, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("error occurred while storing yaml state file at path %s: %v", stateFilePath, err)
	}

	err = os.WriteFile(stateFilePath, stateFileData, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("error occurred while storing yaml state file at path %s: %v", stateFilePath, err)
	}
	return nil
}
