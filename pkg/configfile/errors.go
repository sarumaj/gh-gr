package configfile

import (
	"errors"
)

var (
	msgConfExists = "Configuration file already exists in current directory. " +
		"Please run 'update' if you want to update your settings. " +
		"Alternatively, run 'purge' if you want to initialize the repositories again."
	ErrConfExists = errors.New(msgConfExists)

	msgConfNotExists = "Couldn't find configuration file in current directory or any " +
		"parent directory. Make sure that you are in the correct directory and that init has " +
		"been run successfully."
	ErrConfNotExists = errors.New(msgConfNotExists)
)
