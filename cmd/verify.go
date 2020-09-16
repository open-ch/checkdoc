package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-ch/checkdoc/pkg/checkdoc"
	"github.com/open-ch/go-libs/fsutils"
	"github.com/spf13/cobra"
)

// A file or dir name telling us we are at the root of a git repo
const gitRootIndicator = ".git"

func init() {

	var respectGitIgnore bool

	// verifyCmd represents the verify command
	var verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Runs sanity checks on the documentation",
		Long: `Run some checks against the markdown documentation found in a directory hierarchy.

Currently, verify will check for two things:
 - orphan README.md files: these are files that are not linked to 
   from the repo's root directory, either directly or indirectly.
 - broken links.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := runVerify(respectGitIgnore)
			if err != nil {
				logger.Errorf("Verify failed: %s", err)
				os.Exit(1)
			}
		},
	}

	verifyCmd.Flags().BoolVar(&respectGitIgnore, "respect-git-ignore", true,
		`If true, will check all potential documents against the repository's gitignore files.'`)

	// 'verify' only needs the global flags.
	// Viper seems not to tolerate a CLI with a single, root command, so this file defines the
	// only thing that you can currently do.
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(respectGitIgnore bool) error {
	absTreeRoot, err := getCorrectPathToTreeRoot(treeRoot, resolveRepoRoot)
	if err != nil {
		return err
	}
	logger.Infof("Running verify on tree root %s", absTreeRoot)
	return verifyTree(absTreeRoot, respectGitIgnore)
}

func verifyTree(treeRoot string, respectGitIgnore bool) error {
	logger.Infof("Considering basenames %v and extensions %v", baseNames, extensions)
	nodes, err := checkdoc.BuildLinkGraphNodes(treeRoot, baseNames, extensions, respectGitIgnore)

	if err != nil {
		logger.Errorf("Could not build the link graph for tree root %s: %s", treeRoot, err)
		return err
	}

	logNodes(nodes)

	reports := checkdoc.BuildReport(treeRoot, nodes, implicitIndexes)
	if !checkdoc.ValidateReports(reports, logger) {
		logger.Errorf("Verify failed on tree root %s", treeRoot)
		return fmt.Errorf("verify failed on tree root %s", treeRoot)
	}
	logger.Infof("Validated doc tree root %s", treeRoot)
	return nil
}

func logNodes(nodes []checkdoc.LinkGraphNode) {
	logger.Debugf("Found %d nodes at:", len(nodes))
	for _, node := range nodes {
		logger.Debugf("\t%s:", node.RelativePath)
	}
}

// getCorrectPathToTreeRoot is in charge of returning an absolute path to the 'correct' tree root, depending on the
// specified 'resolveRepoRoot' flag:
//  - if 'resolveRepoRoot' is true, the hierarchy above the passedRootPath will be explored
//    for a git repository root, and that path will be used as the root
//    from which to check documentation consistency.
//  - otherwise, the passed path will be returned after having made sure it is absolute,
//    calling Abs() if required.
func getCorrectPathToTreeRoot(passedRootPath string, resolveRepoRoot bool) (string, error) {
	absPath, err := filepath.Abs(passedRootPath)
	if err != nil {
		logger.Errorf("Could not convert %s to an absolute path: %s", passedRootPath, err)
		return "", err
	}
	if resolveRepoRoot {
		repoRoot, err := fsutils.SearchClosestParentContaining(absPath, gitRootIndicator)
		if err != nil {
			logger.Errorf("Failed to find git repo root from path %s:%s", passedRootPath, err)
			return "", err
		}
		return repoRoot, nil
	}
	return absPath, nil
}
