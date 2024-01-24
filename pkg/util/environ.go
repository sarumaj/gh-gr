package util

import "os"

// Prefix for relevant environment variables.
const EnvPrefix = "GITHUB_REPO_"

// Verbose env variable.
const Verbose envVariable = "VERBOSE"

// Custom type for env var names.
type envVariable string

// Retrieve environment variable.
func Getenv(key envVariable) string { return os.Getenv(EnvPrefix + string(key)) }

// Retrieve environment variable of boolean type.
func GetenvBool(key envVariable) bool {
	switch Getenv(key) {
	case "true", "TRUE", "True", "1", "Y", "y", "YES", "yes":
		return true

	default:
		return false

	}
}
