package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Points to the root of the documentation hierarchy to validate
	treeRoot        string
	resolveRepoRoot bool

	respectGitIgnore bool

	verbose bool

	rootCmd = &cobra.Command{
		Use:   "checkdoc",
		Short: "checkdoc is a markdown documentation validator",
		Long: "A markdown documentation validator intended to enforce a healthy documentation " +
			"in settings such as a fat repo.",
	}

	// Some other things that are currently fixed but may at some point be configurable
	baseNames       []string // Currently we only search markdown files based on the extension
	extensions      = []string{".md"}
	implicitIndexes = []string{"README.md"} // When links point to a directory, we check for a readme within it

	// Logger...
	logger = log.New()
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

	logger.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})
}

// Execute runs the whole enchilada, baby!
func Execute() {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
