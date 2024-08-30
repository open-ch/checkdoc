package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-ch/checkdoc/checkdoc"
)

func getCatLinksCommand() *cobra.Command {
	var outputPath string

	catLinksCmd := &cobra.Command{
		Use:   "catlinks",
		Short: "Searches and dumps internal links found in the documentation files",
		Long: `Searches and dumps internal links found int documentation files:
This only includes links to local files, and does not include any HTTP, FTP or any other such link.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runCatLinks(outputPath)
		},
	}

	catLinksCmd.Flags().StringVarP(&outputPath, "output", "o", "",
		"File to write the output to. Will output to STDOUT if not set.")

	return catLinksCmd
}

func runCatLinks(outputPath string) error {
	// TODO avoid globals treeRoot and resolveRepoRoot
	absTreeRoot, err := filepath.Abs(treeRoot)
	if err != nil {
		return fmt.Errorf("could not convert %s to an absolute path: %w", treeRoot, err)
	}

	if resolveRepoRoot {
		repoRoot, err := getRepositoryRoot(absTreeRoot)
		if err != nil {
			return fmt.Errorf("failed to find git repo root from path %s: %w", absTreeRoot, err)
		}
		absTreeRoot = repoRoot
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
	extensions := []string{".md"}
	baseNames := []string{}
	slog.Debug("building links",
		"extensions", extensions, "baseNames", baseNames)
	nodes, err := checkdoc.BuildLinkGraphNodes(treeRoot, baseNames, extensions, respectGitIgnore)
	if err != nil {
		return err
	}
	localPaths := checkdoc.BuildLocalPathSet(nodes)
	localPathsWithSlash, err := checkdoc.EnsureDirectoriesEndWithSlash(treeRoot, localPaths)
	filteredPaths := filterLinks(localPathsWithSlash)

	for path := range filteredPaths {
		_, printerr := fmt.Fprintln(output, path)
		err = errors.Join(err, printerr)
	}

	return err
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
