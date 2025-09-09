package runner

import "fmt"

// ValidateContainerName checks the Docker name constraints
func ValidateContainerName(name string) error {
	if !dockerNameRegex.MatchString(name) {
		return fmt.Errorf("invalid container name: %q", name)
	}
	return nil
}
