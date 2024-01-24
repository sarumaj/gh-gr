package commands

import (
	cobra "github.com/spf13/cobra"
)

var initCmd = func() *cobra.Command {
	initCmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"setup"},
		Short:   "Initialize repository mirror",
		Long: "Initialize repository mirror.\n\n" +
			"Automatically generates a list of repositories a given user has permissions to.\n" +
			"Supports filtering by repository blob size and with regular expressions.\n" +
			"Regular expressions support following features:\n\n" +
			"\t- Python-style capture groups (?P<name>re)\n" +
			"\t- .NET-style capture groups (?<name>re) or (?'name're)\n" +
			"\t- comments (?#comment)\n" +
			"\t- possessive match (?>re)\n" +
			"\t- positive lookahead (?=re)\n" +
			"\t- negative lookahead (?!re)\n" +
			"\t- positive lookbehind (?<=re)\n" +
			"\t- negative lookbehind (?<!re)\n" +
			"\t- back reference \\1\n" +
			"\t- named back reference \\k'name'\n" +
			"\t- named ascii character class [[:foo:]]\n" +
			"\t- conditionals (?(expr)yes|no)\n",
		Example: "gh gr init " +
			"--concurrency 100 --timeout \"10s\" " +
			"--dir \"/home/user/github\" --subdirs --sizelimit $((10*1024*1024)) --include \"(ORG1|ORG2)/.*\" --exclude \"ORG1/REPO1\"",
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
	flags.StringArrayVarP(&configFlags.Excluded, "exclude", "e", []string{}, "Regular expressions for repositories to exclude")
	flags.StringArrayVarP(&configFlags.Included, "include", "i", []string{}, "Regular expressions for repositories to include explicitly")

	return initCmd
}()
