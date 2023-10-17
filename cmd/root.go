package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Points to the root of the documentation hierarchy to validate
	treeRoot        string
	resolveRepoRoot bool

	respectGitIgnore bool

	verbose bool

	rootCmd = &cobra.Command{
		// Don't show usage when reporting errors.
		// Only show with -h, --help or when subcommands are missing.
		SilenceUsage: true,
		// Don't show errors twice, we handle error in Execute()
		SilenceErrors: true,
		Use:           "checkdoc",
		Short:         "checkdoc is a markdown documentation validator",
		Long: "A markdown documentation validator intended to enforce a healthy documentation " +
			"in settings such as a fat repo.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				// Note NewTextHandler uses key=value pairs unlike the default slog format
				// currently there's no easy way to access the default format as a handler
				logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				})
				slogger := slog.New(logHandler)
				slog.SetDefault(slogger)
			}
		},
	}

	// Some other things that are currently fixed but may at some point be configurable
	baseNames       []string // Currently we only search markdown files based on the extension
	extensions      = []string{".md"}
	implicitIndexes = []string{"README.md"} // When links point to a directory, we check for a readme within it
)

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVarP(&treeRoot, "root", "r", ".",
		"Path to the root of the markdown documentation hierarchy to validate")

	rootCmd.PersistentFlags().BoolVarP(&resolveRepoRoot, "use-git-root", "g", true,
		"from the given root, fall back to the repository's root."+
			" This will cause checkdoc to fail if --root is not pointing to a repository.")

	rootCmd.PersistentFlags().BoolVar(&respectGitIgnore, "respect-git-ignore", true,
		`If true, will check all potential documents against the repository's gitignore files.'`)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Detailed output if true")
}

// Execute runs the whole enchilada, baby!
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("checkdoc failed", "err", err)
		os.Exit(1)
	}
}
