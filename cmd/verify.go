package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"

	"github.com/open-ch/checkdoc/checkdoc"
)

// A file or dir name telling us we are at the root of a git repo
const gitRootIndicator = ".git"

func init() {
	var verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Runs sanity checks on the documentation",
		Long: `Run some checks against the markdown documentation found in a directory hierarchy.

Currently, verify will check for two things:
 - orphan README.md files: these are files that are not linked to
   from the repo's root directory, either directly or indirectly.
 - broken links.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(respectGitIgnore)
		},
	}

	rootCmd.AddCommand(verifyCmd)
}

func runVerify(respectGitIgnore bool) error {
	// TODO avoid globals treeRoot and resolveRepoRoot
	absTreeRoot, err := filepath.Abs(treeRoot)
	if err != nil {
		return fmt.Errorf("Could not convert %s to an absolute path: %w", treeRoot, err)
	}

	if resolveRepoRoot {
		repoRoot, err := getRepositoryRoot(absTreeRoot)
		if err != nil {
			return fmt.Errorf("Failed to find git repo root from path %s: %w", absTreeRoot, err)
		}
		absTreeRoot = repoRoot
	}

	slog.Info("Running verify on tree root", "rootpath", absTreeRoot)
	return verifyTree(absTreeRoot, respectGitIgnore)
}

func verifyTree(treeRoot string, respectGitIgnore bool) error {
	slog.Debug("building links using configured basenames and extensions",
		"basenames", baseNames, "extensions", extensions)
	nodes, err := checkdoc.BuildLinkGraphNodes(treeRoot, baseNames, extensions, respectGitIgnore)

	if err != nil {
		return fmt.Errorf("Could not build the link graph for tree root %s: %w", treeRoot, err)
	}

	logNodes(nodes)

	reports := checkdoc.BuildReport(treeRoot, nodes, implicitIndexes)
	if !checkdoc.ValidateReports(reports) {
		return fmt.Errorf("verify failed on tree root %s", treeRoot)
	}
	slog.Info("Validated doc tree root successfully")
	return nil
}

func logNodes(nodes []checkdoc.LinkGraphNode) {
	slog.Debug("Found nodes", "nodescount", len(nodes))
	for _, node := range nodes {
		slog.Debug(fmt.Sprintf("\t%s:", node.RelativePath))
	}
}

func getRepositoryRoot(path string) (string, error) {
	gitCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitCmd.Dir = path
	output, err := gitCmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}
