package commands

import "os"

const envPrefix = "GITHUB_REPO_"
const verbose envVariable = "VERBOSE"

type envVariable string

func getenv(key envVariable) string { return os.Getenv(envPrefix + string(key)) }

func getenvBool(key envVariable) bool {
	switch getenv(verbose) {
	case "true", "TRUE", "True", "1", "Y", "y", "YES", "yes":
		return true

	default:
		return false

	}
}
