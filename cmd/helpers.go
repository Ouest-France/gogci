package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func convertToEnvName(name string) (string, error) {

	// Remove characters that are not alphanum or . _ -
	r, err := regexp.Compile("[a-zA-Z0-9._-]")
	if err != nil {
		return "", fmt.Errorf("failed to compile regex: %w", err)
	}
	var sanitizedName string
	for _, l := range name {
		if r.MatchString(string(l)) {
			sanitizedName = sanitizedName + string(l)
		}
	}

	// Error if sanitized name is empty
	if sanitizedName == "" {
		return "", errors.New("empty sanitized name")
	}

	// Set name to be in upper case
	capitalizedName := strings.ToUpper(sanitizedName)

	// Replace - and . with _
	substituedName := capitalizedName
	for old, new := range map[string]string{"-": "_", ".": "_"} {
		substituedName = strings.ReplaceAll(substituedName, old, new)
	}

	// Add VAULTENV prefix
	finalName := "VAULTENV_" + substituedName

	return finalName, nil
}

// ErrorToEval print an error string that can be evaluated by a shell to print
func ErrorToEval(err error) {
	fmt.Printf("echo \"echo %s\"", err)
}
