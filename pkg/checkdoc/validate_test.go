package checkdoc

import (
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestValidateSimple(t *testing.T) {
	treeRoot := filepath.Join(getTestDir(), "sub-dir-a", "nested-sub-dir-a")
	extensions := []string{".md"}
	nodes, err := BuildLinkGraphNodes(treeRoot, []string{}, extensions, false)
	assert.NoError(t, err)

	logger := log.New()

	reports := BuildReport(treeRoot, nodes, []string{})
	assert.True(t, ValidateReports(reports, logger), "The specified directory is expected to be valid.")
}

func TestBuildLinkGraphNodesWithGitIgnore(t *testing.T) {
	treeRoot := getTestDir()
	extensions := []string{".md"}
	nodes, err := BuildLinkGraphNodes(treeRoot, []string{}, extensions, true)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(nodes), "Expecting To have only three nodes due to gitignore.")
}

func TestValidateFail(t *testing.T) {
	treeRoot := getTestDir()
	implicitIndexes := []string{"README.md", "README"}
	baseNames := []string{"README"}
	extensions := []string{".md"}
	nodes, err := BuildLinkGraphNodes(treeRoot, baseNames, extensions, false)
	assert.NoError(t, err)

	logger := log.New()

	reports := BuildReport(treeRoot, nodes, implicitIndexes)
	assert.False(t, ValidateReports(reports, logger), "The complete test setup is expected to fail validation")
}

func TestBuildReport(t *testing.T) {
	treeRoot := getTestDir()
	implicitIndexes := []string{"README.md", "README"}
	baseNames := []string{"README"}
	extensions := []string{".md"}
	nodes, err := BuildLinkGraphNodes(treeRoot, baseNames, extensions, false)
	assert.NoError(t, err)

	reports := BuildReport(treeRoot, nodes, implicitIndexes)

	assert.Equal(t, 7, len(reports))
	// Quick sanity check...
	for _, node := range nodes {
		assert.Contains(t, reports, node.RelativePath)
	}

	assert.True(t, reports["README.md"].IsOrphan, "The top level README file is expected to be orphan.")
	assert.Equal(t, 0, len(reports["README.md"].DeadLinks), "No dead link expected here")

	assert.False(t, reports["some-md-file.md"].IsOrphan, "This file should be linked to from the root README")
	assert.Equal(t, 0, len(reports["README.md"].DeadLinks), "No dead link expected here")

	assert.False(t, reports["sub-dir-b/README.md"].IsOrphan, "This file should be linked to")
	assert.Equal(t, 0, len(reports["sub-dir-b/README.md"].DeadLinks), "No dead link expected here")

	assert.False(t, reports["sub-dir-a/README"].IsOrphan, "This file should be linked to")
	assert.Equal(t, []string{"sub-dir-a/not-here"}, reports["sub-dir-a/README"].DeadLinks)

	// the changelog is not linked from anyone
	assert.True(t, reports["sub-dir-a/CHANGELOG.md"].IsOrphan, "This file should be linked to")
	assert.Equal(t, []string{"sub-dir-a/dead-end"}, reports["sub-dir-a/CHANGELOG.md"].DeadLinks)

	assert.False(t, reports["sub-dir-a/nested-sub-dir-a/README.md"].IsOrphan, "This file should be linked to")
	assert.Equal(t, 0, len(reports["sub-dir-a/nested-sub-dir-a/README.md"].DeadLinks))

	assert.False(t, reports["sub-dir-a/nested-sub-dir-a/some-other-md-file.md"].IsOrphan, "This file should be linked to")
	assert.Equal(t, 0, len(reports["sub-dir-a/nested-sub-dir-a/some-other-md-file.md"].DeadLinks))

}

func TestBuildPathSet(t *testing.T) {
	nodeA := LinkGraphNode{"some/path", nil, []string{"path/a", "path/b"}}
	nodeB := LinkGraphNode{"some/path", nil, []string{"path/b", "path/c"}}

	pathSet := BuildLocalPathSet([]LinkGraphNode{nodeA, nodeB})
	expected := map[string]bool{
		"path/a": false,
		"path/b": false,
		"path/c": false,
	}

	assert.Equal(t, expected, pathSet, "Should build a proper path set")
}

func TestCheckForNonExistingPaths(t *testing.T) {
	treeRoot := getTestDir()
	pathSet := map[string]bool{
		"non-existing":      false, // something non existing
		"sub-dir-a/neither": false, // not existing either
		"sub-dir-a":         false, // a directory
		"README.md":         false,
	}

	nonExisting := checkForNonExistingPaths(treeRoot, pathSet)

	assert.Contains(t, nonExisting, "non-existing")
	assert.Contains(t, nonExisting, "sub-dir-a/neither")
}

func TestResolveImplicitPaths(t *testing.T) {
	treeRoot := getTestDir()
	pathSet := map[string]bool{
		"non-existing":               false, // something non existing
		"sub-dir-a":                  false, // a directory with two matching implicits
		"sub-dir-a/nested-sub-dir-b": false, // a directory with no matching implicit
		"README.md":                  false, // a plain file
	}

	resolved := resolveImplicitPaths(treeRoot, []string{"README", "CHANGELOG.md", "nested-sub-dir-a"}, pathSet)

	assert.Equal(t, map[string]bool{
		"sub-dir-a/README":           false, // Exists, resolved implicitly
		"sub-dir-a":                  false, // Exists, points to a directory with an implicit file
		"sub-dir-a/CHANGELOG.md":     false, // Exists, resolved implicitly
		"sub-dir-a/nested-sub-dir-b": false, // Exists, points to a directory with no implicit file
		"README.md":                  false, // a plain file
	}, resolved)
}

func TestResolveImplicitPathsTestData(t *testing.T) {
	treeRoot := getTestDir()
	implicitIndexes := []string{"README.md", "README"}
	baseNames := []string{"CHANGELOG", "README"}
	extensions := []string{".md"}

	nodes, err := BuildLinkGraphNodes(treeRoot, baseNames, extensions, false)
	assert.NoError(t, err)
	rawPathSet := BuildLocalPathSet(nodes)

	// At this point we have not done any implicit resolution:
	// hence, only files with an explicit reference are present in the map.
	assert.Equal(t, 8, len(rawPathSet))
	assert.Equal(t, map[string]bool{
		"some-md-file.md":            false,
		"sub-dir-b":                  false,
		"sub-dir-a/nested-sub-dir-a": false,
		"sub-dir-a/nested-sub-dir-a/some-other-md-file.md": false,
		"sub-dir-a/README":           false,
		"sub-dir-a/nested-sub-dir-b": false,
		"sub-dir-a/dead-end":         false,
		"sub-dir-a/not-here":         false,
	}, rawPathSet)

	// The resolution step will add any files 'implicitly' linked to,
	// while also removing non-existing paths.
	resolvedPaths := resolveImplicitPaths(treeRoot, implicitIndexes, rawPathSet)

	assert.Equal(t, 8, len(resolvedPaths))
	assert.Equal(t, map[string]bool{
		"some-md-file.md":                                  false,
		"sub-dir-b":                                        false,
		"sub-dir-b/README.md":                              false,
		"sub-dir-a/nested-sub-dir-a":                       false,
		"sub-dir-a/nested-sub-dir-a/README.md":             false,
		"sub-dir-a/nested-sub-dir-a/some-other-md-file.md": false,
		"sub-dir-a/nested-sub-dir-b":                       false,
		"sub-dir-a/README":                                 false,
	}, resolvedPaths)

}

func TestSearchOrphansAllOrphans(t *testing.T) {

	nodeA := LinkGraphNode{"README.md", nil, []string{"non-existing"}}
	nodeB := LinkGraphNode{"sub-dir-a/README.md", nil, []string{}}

	pathSet := map[string]bool{
		"non-existing": false, // something non existing, thus not implicit
	}

	orphans := buildOrphanReport(pathSet, []LinkGraphNode{nodeA, nodeB})

	assert.Equal(t, map[string]bool{
		"README.md":           true, // nothing should point to the root node
		"sub-dir-a/README.md": true, // the root node points to this one.
	}, orphans, "expected to find only orphans")

}

func TestSearchOrphansOnlyRootOrphan(t *testing.T) {

	nodeA := LinkGraphNode{"README.md", nil,
		[]string{"non-existing", "sub-dir-a/README.md"}}
	nodeB := LinkGraphNode{"sub-dir-a/README.md", nil, []string{}}

	pathSet := map[string]bool{
		"non-existing":        false, // non implicit link
		"sub-dir-a/README.md": false, // non implicit link
	}

	orphans := buildOrphanReport(pathSet, []LinkGraphNode{nodeA, nodeB})

	assert.Equal(t, map[string]bool{
		"README.md":           true,  // nothing should point to the root node
		"sub-dir-a/README.md": false, // the root node points to this one.
	}, orphans, "expected to find only the root as orphan")

}

func TestBuildDeadLinkReport(t *testing.T) {

	pathSet := map[string]bool{
		"README.md":        false,
		"sub-dir-a":        false, // points to a dir
		"sub-dir-a/README": true,  // points to an index file in a dir, implicitly resolved
	}

	nodeA := LinkGraphNode{"README.md", nil,
		[]string{"sub-dir-a"}}
	nodeB := LinkGraphNode{"sub-dir-a/README", nil,
		[]string{"README.md", "sub-dir-c", "sub-dir-c/some-file"}}

	deadLinkReport := buildDeadLinkReport(pathSet, []LinkGraphNode{nodeA, nodeB})
	assert.Equal(t, map[string][]string{
		"README.md":        {},
		"sub-dir-a/README": {"sub-dir-c", "sub-dir-c/some-file"},
	}, deadLinkReport)
}

func TestEnsureDirectoriesEndWithSlash(t *testing.T) {
	treeRoot := getTestDir()

	pathSet := map[string]bool{
		"README.md":        false,
		"sub-dir-a":        false,
		"sub-dir-a/README": false,
	}

	processed, err := EnsureDirectoriesEndWithSlash(treeRoot, pathSet)

	assert.Nil(t, err, "Not expecting a failure")
	assert.Equal(t, map[string]bool{
		"README.md":        false,
		"sub-dir-a/":        false,
		"sub-dir-a/README": false,
	}, processed)
}
