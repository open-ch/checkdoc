package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"osag/libs/go/observability/logging"
	"github.com/open-ch/checkdoc/checkdoc"

	"github.com/spf13/cobra"
)

func getVerifyCommand() *cobra.Command {
	verifyCmd := cobra.Command{
		Use:   "verify",
		Short: "Runs sanity checks on the documentation",
		Long: `Run some checks against the markdown documentation found in a directory hierarchy.

Currently, verify will check for two things:
 - orphan README.md files: these are files that are not linked to
   from the repo's root directory, either directly or indirectly.
 - broken links.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runVerify(respectGitIgnore)
		},
	}

	return &verifyCmd
}

func runVerify(respectGitIgnore bool) error {
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

	logging.Infow("Running verify on tree root", "rootpath", absTreeRoot)
	return verifyTree(absTreeRoot, respectGitIgnore)
}

func verifyTree(treeRoot string, respectGitIgnore bool) error {
	extensions := []string{".md"}
	baseNames := []string{}
	logging.Debugw("building links",
		"extensions", extensions, "baseNames", baseNames)
	nodes, err := checkdoc.BuildLinkGraphNodes(treeRoot, baseNames, extensions, respectGitIgnore)

	if err != nil {
		return fmt.Errorf("could not build the link graph for tree root %s: %w", treeRoot, err)
	}

	logNodes(nodes)

	reports := checkdoc.BuildReport(treeRoot, nodes, []string{"README.md"})
	if !checkdoc.ValidateReports(reports) {
		return fmt.Errorf("verify failed on tree root %s", treeRoot)
	}
	logging.Info("Validated doc tree root successfully")
	return nil
}

func logNodes(nodes []checkdoc.LinkGraphNode) {
	logging.Debugw("Found nodes", "nodescount", len(nodes))
	for _, node := range nodes {
		logging.Debug(fmt.Sprintf("\t%s:", node.RelativePath))
	}
}

func getRepositoryRoot(path string) (string, error) {
	gitCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitCmd.Dir = path
	output, err := gitCmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}
