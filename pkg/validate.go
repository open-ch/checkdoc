package pkg

import (
	"os"
	"osag/libs/logger"
	"path/filepath"
)

// NodeReport contains some information about the quality of a node
type NodeReport struct {
	Node      LinkGraphNode // The underlying node
	DeadLinks []string      // (local) dead links that this node contain
	IsOrphan  bool          // Does anything point to this node
}

// TODO the whole package needs a little rewrite to use some form of object that contains the config
//  and state (logger, parameters, etc) and on which the methods are called.

// TODO if we ever want to do more fancy things, this part of the lib deserves to be rewritten to use
// a graph library, something like gonum/graph.

// ValidateReports the passed report map. Currently, this checks that:
//  - there are no orphan pages (without inbound links), except for the root README.md
//  - internal links point to existing things (either files, directories or other readmes)
//
// This method returns 'true' if no issues where found, and false otherwise
func ValidateReports(reports map[string]NodeReport, logger logger.Logger) bool {
	// TODO this function could take a logger and print useful things to it.
	// TODO consider adding rules allowing for things like CHANGELOG files not to be linked to
	// TODO add a flag to tolerate or refuse things like README (ie, force the extension)

	var isValid = true
	logger.Infof("Checking for orphaned documents...")
	var orphans []string
	for path, report := range reports {
		// TODO specify the root file via an option
		if report.IsOrphan && path != "README.md" {
			orphans = append(orphans, path)
			isValid = false
		}
	}
	logOrphans(orphans, logger)

	logger.Infof("Checking for dead links...")
	var withDeadLinks []NodeReport
	for _, report := range reports {
		if len(report.DeadLinks) != 0 {
			withDeadLinks = append(withDeadLinks, report)
			isValid = false
		}
	}

	logDeadLinks(withDeadLinks, logger)

	return isValid
}

func logOrphans(orphans []string, logger logger.Logger) {
	if len(orphans) == 0 {
		logger.Infof("No orphans found.")
		return
	}
	logger.Errorf("Located some orphan documents:")
	for _, orphan := range orphans {
		logger.Errorf("\t%s", orphan)
	}
}

func logDeadLinks(withDeadLinks []NodeReport, logger logger.Logger) {
	if len(withDeadLinks) == 0 {
		logger.Infof("No dead links found.")
		return
	}
	logger.Errorf("Located some files with dead links:")
	for _, invalid := range withDeadLinks {
		logger.Errorf("\t%s", invalid.Node.RelativePath)
		for _, deadLink := range invalid.DeadLinks {
			logger.Errorf("\t\t%s", deadLink)
		}
	}
}

// BuildReport will run through the passed nodes, using the specified root to run its checks, and build a report for each node
// that will be container within the returned map
func BuildReport(treeRoot string, nodes []LinkGraphNode, implicitIndexes []string) map[string]NodeReport {
	rawPathSet := buildPathSet(nodes)

	resolvedPaths := resolveImplicitPaths(treeRoot, implicitIndexes, rawPathSet)

	deadLinks := buildDeadLinkReport(resolvedPaths, nodes)
	orphans := buildOrphanReport(resolvedPaths, nodes)

	nodeReports := make(map[string]NodeReport)

	for _, node := range nodes {
		nodeReports[node.RelativePath] = NodeReport{
			Node:      node,
			DeadLinks: deadLinks[node.RelativePath],
			IsOrphan:  orphans[node.RelativePath],
		}
	}

	return nodeReports
}

// buildOrphanReport looks up all passed nodes in the specified pathSet, and returns a map telling if they are present in it,
// ie, have or have not been linked to.
// Note:
//  - this won't catch any node that linked to itself, but this seems like an acceptable corner case, as it
//    would look pretty obvious in a document anyway.
//  - resolvedPathSet is expected to contain only paths to files, not directories.
//
func buildOrphanReport(resolvedPathSet map[string]bool, nodes []LinkGraphNode) map[string]bool {
	orphanReport := make(map[string]bool)
	// resolvedPathSet contains an exhaustive set of all existing local links, relative to the tree root.
	// We check all nodes against it to check if they were pointed to from somewhere.
	for _, node := range nodes {
		// Do a lookup, check if the value is present
		_, present := resolvedPathSet[node.RelativePath]
		// ... and set the value in the report accordingly: if it is present, it's not an orphan
		orphanReport[node.RelativePath] = !present
	}
	return orphanReport
}

// buildDeadLinkReport builds a report of dead links for each passed graph node.
// resolvedPathSet must have been computed beforehand, and is expected to contain a set of all links
// pointed to from nodes, minus any invalid link, so that it may be used to check for wrong links.
func buildDeadLinkReport(resolvedPathSet map[string]bool, nodes []LinkGraphNode) map[string][]string {
	toRet := make(map[string][]string)

	for _, node := range nodes {
		deadLinks := []string{}
		for _, link := range node.NormalizedLocalRelativeLinks {
			if _, present := resolvedPathSet[link]; !present {
				deadLinks = append(deadLinks, link)
			}
		}
		// Keep track of the dead links for that node, identified by its relative path.
		toRet[node.RelativePath] = deadLinks
	}

	return toRet
}

// buildPathSet returns a set of all links found in the passed nodes.
func buildPathSet(nodes []LinkGraphNode) map[string]bool {
	toRet := make(map[string]bool)
	for _, node := range nodes {
		for _, relativePath := range node.NormalizedLocalRelativeLinks {
			toRet[relativePath] = false
		}
	}
	return toRet
}

// checkForNonExistingPaths checks that all keys in the passed set exists, and returns a slice
// of all the ones that don't exist.
func checkForNonExistingPaths(treeRoot string, pathSet map[string]bool) []string {
	var notExisting []string
	for path := range pathSet {
		if _, err := os.Stat(filepath.Join(treeRoot, path)); os.IsNotExist(err) {
			notExisting = append(notExisting, path)
		}
	}
	return notExisting
}

// resolveImplicitLinks will check the passed link set for directories, and for each one of them,
// verify that it contains a file present in 'implicitIndexes'.
// If such a file does not exist, but the directory exists, the directory is still added to the returned map.
//
// Note that any non existing file or directory will not be present in the returned map either.
func resolveImplicitPaths(treeRoot string, implicitIndexes []string, pathSet map[string]bool) map[string]bool {
	toRet := make(map[string]bool)
	for path := range pathSet {
		absPath := filepath.Join(treeRoot, path)
		info, err := os.Stat(absPath)

		if os.IsNotExist(err) {
			// That path does not exist: ignore it and continue
			continue
		}

		if info.IsDir() {
			// This is a directory: we check if it contains any of the expected files:
			for _, indexFile := range implicitIndexes {
				relativeIndexPath := filepath.Join(path, indexFile)
				absIndexPath := filepath.Join(treeRoot, relativeIndexPath)
				indexFileInfo, err := os.Stat(absIndexPath)
				if os.IsNotExist(err) || indexFileInfo.IsDir() {
					// this potential index file does not exists or points to a directory:
					// it can't be used as an implicit path to a file.
					continue
				}
				// the path exists and is a file: keep it as a "resolved implicit"
				toRet[relativeIndexPath] = false
				// don't break, as there could be other existing implicit index files.
			}
		}

		// At this point we know the path exists, and if it was a directory, we checked for
		// possible index files it may contain.
		// We keep it whether it is a file or directory, as links to existing directories
		// even without documentation are currently seen as valid.
		toRet[path] = false
	}
	return toRet
}
