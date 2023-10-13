package util

import "os"

const EnvPrefix = "GITHUB_REPO_"
const Verbose envVariable = "VERBOSE"

type envVariable string

func Getenv(key envVariable) string { return os.Getenv(EnvPrefix + string(key)) }

func GetenvBool(key envVariable) bool {
	switch Getenv(Verbose) {
	case "true", "TRUE", "True", "1", "Y", "y", "YES", "yes":
		return true

	default:
		return false

	}
}
