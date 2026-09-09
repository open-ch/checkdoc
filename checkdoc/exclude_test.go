package checkdoc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// writeTree creates the given files, relative to a fresh temporary tree root.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	treeRoot := t.TempDir()
	for name, content := range files {
		path := filepath.Join(treeRoot, name)
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		assert.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	return treeRoot
}

func TestReadExclusionsMissingFile(t *testing.T) {
	exclusions, err := readExclusions(t.TempDir())

	assert.NoError(t, err, "A missing ignore file is not an error.")
	assert.Empty(t, exclusions)
}

func TestReadExclusions(t *testing.T) {
	treeRoot := writeTree(t, map[string]string{
		IgnoreFile: "# Content, not documentation\n" +
			"\n" +
			"  content/help-section/  \n" +
			"./other/tree\n" +
			"single-file.md\n",
	})

	exclusions, err := readExclusions(treeRoot)

	assert.NoError(t, err, "Should not fail on valid entries.")
	// Comments and blank lines are dropped; trailing slashes and "./" are cleaned away.
	assert.Equal(t, []string{"content/help-section", "other/tree", "single-file.md"}, exclusions)
}

func TestReadExclusionsRejectsPathsOutsideTheTree(t *testing.T) {
	for _, entry := range []string{"/absolute/path", "..", "../sibling", "."} {
		treeRoot := writeTree(t, map[string]string{IgnoreFile: entry + "\n"})

		exclusions, err := readExclusions(treeRoot)

		assert.Error(t, err, "Should reject %q", entry)
		assert.Nil(t, exclusions, "Nothing should be returned on failure")
	}
}

func TestIsExcluded(t *testing.T) {
	exclusions := []string{"content", "docs/generated.md"}

	assert.True(t, isExcluded("content", exclusions), "The directory itself is excluded.")
	assert.True(t, isExcluded("content/pages/Home.md", exclusions), "Anything below it is excluded.")
	assert.True(t, isExcluded("docs/generated.md", exclusions), "A single file can be excluded.")

	assert.False(t, isExcluded("contents/Home.md", exclusions), "A longer sibling name is not a match.")
	assert.False(t, isExcluded("other/content/Home.md", exclusions), "Entries are anchored at the tree root.")
	assert.False(t, isExcluded("docs/README.md", exclusions), "Siblings of an excluded file are kept.")
}

func TestBuildLinkGraphNodesHonoursExclusions(t *testing.T) {
	treeRoot := writeTree(t, map[string]string{
		"README.md":                "# Root\n\nSee [the tool](tools/README.md).\n",
		"tools/README.md":          "# Tools\n",
		"content/pages/Home.md":    "# Home\n\nSee [routing](/help/routing).\n",
		"content/pages/sub/FAQ.md": "# FAQ\n",
		IgnoreFile:                 "content\n",
	})

	nodes, err := BuildLinkGraphNodes(treeRoot, []string{}, []string{".md"}, false)
	assert.NoError(t, err, "Should not fail on valid input.")

	var paths []string
	for _, node := range nodes {
		paths = append(paths, node.RelativePath)
	}
	assert.ElementsMatch(t, []string{"README.md", "tools/README.md"}, paths,
		"The excluded directory contributes no nodes, at any depth.")

	// Neither rule can fire for content that was never discovered.
	reports := BuildReport(treeRoot, nodes, []string{"README.md"})
	assert.True(t, ValidateReports(reports),
		"Excluded pages are neither orphans nor sources of dead links.")
}

func TestBuildLinkGraphNodesKeepsLinksIntoExcludedTrees(t *testing.T) {
	treeRoot := writeTree(t, map[string]string{
		"README.md":             "# Root\n\nSee [the content](content/pages/Home.md).\n",
		"content/pages/Home.md": "# Home\n",
		IgnoreFile:              "content\n",
	})

	nodes, err := BuildLinkGraphNodes(treeRoot, []string{}, []string{".md"}, false)
	assert.NoError(t, err, "Should not fail on valid input.")

	// Exclusion only stops discovery. The file still exists, so a link to it from
	// a document that is checked resolves as before.
	reports := BuildReport(treeRoot, nodes, []string{"README.md"})
	assert.Empty(t, reports["README.md"].DeadLinks,
		"A link into an excluded tree is not dead: the file is on disk.")
}
