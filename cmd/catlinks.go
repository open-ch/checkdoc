package cmd

import (
	"bufio"
	"fmt"
	"github.com/open-ch/checkdoc/pkg/checkdoc"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var outputPath string

func init() {
	// verifyCmd represents the verify command
	var catLinksCmd = &cobra.Command{
		Use:   "catlinks",
		Short: "Searches and dumps internal links found in the documentation files",
		Long: `Searches and dumps internal links found int documentation files:
This only includes links to local files, and does not include any HTTP, FTP or any other such link.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := runCatLinks()
			if err != nil {
				logger.Errorf("Verify failed: %s", err)
				os.Exit(1)
			}
		},
	}

	catLinksCmd.Flags().StringVarP(&outputPath, "output", "o", "",
		"File to write the output to. Will output to STDOUT if not set.")

	rootCmd.AddCommand(catLinksCmd)
}

func runCatLinks() error {
	absTreeRoot, err := getCorrectPathToTreeRoot(treeRoot, resolveRepoRoot)
	if err != nil {
		return err
	}
	var outputWriter io.Writer
	if outputPath == "" {
		outputWriter = io.Writer(os.Stdout)
	} else {
		outfile, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer outfile.Close()
		buff := bufio.NewWriter(outfile)
		defer buff.Flush()
		outputWriter = buff
	}
	return catLinks(absTreeRoot, respectGitIgnore, outputWriter)
}

func catLinks(treeRoot string, respectGitIgnore bool, output io.Writer) error {
	logger.Debugf("Considering basenames %v and extensions %v", baseNames, extensions)
	nodes, err := checkdoc.BuildLinkGraphNodes(treeRoot, baseNames, extensions, respectGitIgnore)
	if err != nil {
		return err
	}
	localPaths := checkdoc.BuildLocalPathSet(nodes)
	localPathsWithSlash, err := checkdoc.EnsureDirectoriesEndWithSlash(treeRoot, localPaths)
	filteredPaths := filterLinks(localPathsWithSlash)

	for path := range filteredPaths {
		fmt.Fprintln(output, path)
	}

	return nil
}

// We're interested in:
//   - excluding links to markdown files (implicit or explicit)
//   - excluding links to big files (?)
//   - including links to any local source file
//   - exclude non-file links (not relevant as long as we only deal with files)
//
// This means:
//   - filter out anything that ends in .md, README or CHANGELOG
//   - filter out anything that points to a directory (implicitly that's a README)
//     -
func filterLinks(paths map[string]bool) map[string]bool {
	filtered := make(map[string]bool)
	for path := range paths {
		if !discardPath(path) {
			filtered[path] = false
		}
	}
	return filtered
}

func discardPath(path string) bool {
	return strings.HasSuffix(path, ".md") ||
		strings.HasSuffix(path, "/") ||
		strings.HasSuffix(path, "README") ||
		strings.HasSuffix(path, "CHANGELOG")
}
