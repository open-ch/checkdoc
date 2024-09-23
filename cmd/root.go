package cmd

import (
	"os"

	"osag/libs/go/observability/logging"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	treeRoot         string //nolint:gochecknoglobals // TODO refactor to avoid global
	resolveRepoRoot  bool   //nolint:gochecknoglobals // TODO refactor to avoid global
	respectGitIgnore bool   //nolint:gochecknoglobals // TODO refactor to avoid global
)

// GetRootCommand returns the root command for the whole CLI execution
func GetRootCommand() *cobra.Command {
	var verbose bool

	rootCmd := &cobra.Command{
		// Don't show usage when reporting errors.
		// Only show with -h, --help or when subcommands are missing.
		SilenceUsage: true,
		// Don't show errors twice, we handle error in Execute()
		SilenceErrors: true,
		Use:           "checkdoc",
		Short:         "checkdoc is a markdown documentation validator",
		Long: "A markdown documentation validator intended to enforce a healthy documentation " +
			"in settings such as a fat repo.",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			handler := log.New(os.Stderr)
			if verbose {
				handler.SetLevel(log.DebugLevel)
			}

			// initialize the observability logging library with the charmbracelet pretty logger
			return logging.Init("info", "plain", logging.WithHandler(handler))
		},
	}

	// The default completions don't work very well, hide them.
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(getCatLinksCommand())
	rootCmd.AddCommand(getVerifyCommand())

	rootCmd.PersistentFlags().StringVarP(&treeRoot, "root", "r", ".",
		"Path to the root of the markdown documentation hierarchy to validate")

	rootCmd.PersistentFlags().BoolVarP(&resolveRepoRoot, "use-git-root", "g", true,
		"from the given root, fall back to the repository's root."+
			" This will cause checkdoc to fail if --root is not pointing to a repository.")

	rootCmd.PersistentFlags().BoolVar(&respectGitIgnore, "respect-git-ignore", true,
		`If true, will check all potential documents against the repository's gitignore files.'`)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Detailed output if true")

	return rootCmd
}
