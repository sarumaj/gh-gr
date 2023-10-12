package commands

import (
	cobra "github.com/spf13/cobra"
)

var initCmd = func() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize repository mirror",
		Run: func(*cobra.Command, []string) {
			// call copy to initialize all empty config fields
			initializeOrUpdateConfig(configFlags.Copy(), false)
		},
		PostRun: func(*cobra.Command, []string) {
			updateConfigFlags()
		},
	}

	flags := initCmd.Flags()
	flags.StringVarP(&configFlags.BaseDirectory, "dir", "d", ".", "Directory in which repositories will be stored (either absolute or relative)")
	flags.BoolVarP(&configFlags.SubDirectories, "subdirs", "s", false, "Enable creation of separate subdirectories for each org/user")
	flags.Uint64VarP(&configFlags.SizeLimit, "sizelimit", "l", 0, "Exclude repositories with size exceeded the limit (\"0\": no limit, e.g. limit of 52,428,800 corresponds with 50 MB)")
	flags.StringArrayVarP(&configFlags.Excluded, "exclude", "e", []string{}, "Regular expressions of repositories to exclude")
	flags.StringArrayVarP(&configFlags.Included, "include", "i", []string{}, "Regular expressions of repositories to include (bind stronger than exclusion list)")

	return initCmd
}()
