package awsconfig

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

func WriteCredentials(profile, accessKey, secretKey, sessionToken string) error {

	// Expand home config dir path
	configPath, err := homedir.Expand("~/.aws")
	if err != nil {
		return fmt.Errorf("failed to calculate aws file config path: %w", err)
	}

	// Create config dir
	err = createConfigDir(configPath)
	if err != nil {
		return fmt.Errorf("failed to create aws config dir: %w", err)
	}

	// Expand home credentials path
	credentialsPath, err := homedir.Expand("~/.aws/credentials")
	if err != nil {
		return fmt.Errorf("failed to calculate aws credentials file path: %w", err)
	}

	// Create credentials file
	err = createCredentialsFile(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to create aws credentials file: %w", err)
	}

	// Load credentials file
	credentials, err := ini.Load(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to load aws credentials file: %w", err)
	}

	// Get profile section
	profileSection := credentials.Section(profile)

	// Set credentials
	profileSection.Key("aws_access_key_id").SetValue(accessKey)
	profileSection.Key("aws_secret_access_key").SetValue(secretKey)
	profileSection.Key("aws_session_token").SetValue(sessionToken)

	// Save credentials file
	err = credentials.SaveTo(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to save aws credentials file to disk: %w", err)
	}

	return nil
}

func createConfigDir(path string) error {

	// Check if config dir already exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	// Create dir if it doesn't exist
	err := os.Mkdir(path, 0755)

	return err
}

func createCredentialsFile(path string) error {

	// Check if credentials file already exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	// Create file if it doesn't exist
	credentialsFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer credentialsFile.Close()

	return err
}
