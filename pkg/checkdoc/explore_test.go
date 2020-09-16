package checkdoc

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-ch/checkdoc/pkg/mockrepo"

	"github.com/stretchr/testify/assert"
)

// TODO something like pytest's fixtures probably exists in go, we could save some copy pasta this way.
//  note from Elwin: have a look at https://pkg.go.dev/github.com/stretchr/testify/suite?tab=doc
func getTestDir() string {
	dir := filepath.Join(mockrepo.MockRepo(), "test-data")
	fmt.Printf("dir: %s\n", dir)
	return filepath.Join(mockrepo.MockRepo(), "test-data")
}

func TestBuildLinkGraphNodesFailures(t *testing.T) {
	nodes, err := BuildLinkGraphNodes("/abs/path", []string{}, []string{}, false)
	assert.Nil(t, nodes, "Not expecting any returned value on failure.")
	assert.Error(t, err, "Should fail if both basename and extensions are empty.")

	nodes, err = BuildLinkGraphNodes("rel/path", []string{"README"}, []string{}, false)
	assert.Nil(t, nodes, "Not expecting any returned value on failure.")
	assert.Error(t, err, "Should fail on a relative tree root path.")
}

func TestBuildLinkGraphNodes(t *testing.T) {
	testDir := getTestDir()

	// Simple check...
	singleNode, err := BuildLinkGraphNodes(testDir, []string{"CHANGELOG.md"}, []string{}, false)
	assert.Nil(t, err, "Should not fail on valid input.")
	assert.Equal(t, 1, len(singleNode))

	assert.Equal(t, "sub-dir-a/CHANGELOG.md", singleNode[0].RelativePath)
	assert.Equal(t, []string{"sub-dir-a/nested-sub-dir-a", "sub-dir-a/dead-end", "sub-dir-a/nested-sub-dir-b", "sub-dir-b"}, singleNode[0].NormalizedLocalRelativeLinks)
}

func TestFindRelevantFilesNotExisting(t *testing.T) {
	testDir := getTestDir()

	emptyFind, emptyErr := findMatchingFiles(testDir, []string{}, []string{})
	// Not that returning an error is done from the public method using this function.
	assert.Empty(t, emptyFind, "Should not return anything when no params are passed")
	assert.NoError(t, emptyErr, "Should not fail on empty arguments")

	emptyFind2, err := findMatchingFiles(testDir, []string{"not-existing.md"}, []string{})
	assert.Empty(t, emptyFind2, "Should not return anything on non existing basename and empty extension.")
	assert.NoError(t, err, "Should not fail with valid arguments")

	emptyFind3, err := findMatchingFiles(testDir, []string{}, []string{".yolo"})
	assert.Empty(t, emptyFind3, "Should not return anything on empty basename and non-existing extension")
	assert.NoError(t, err, "Should not fail with valid arguments")
}

func TestFindRelevantFilesByBasename(t *testing.T) {
	testDir := getTestDir()

	singleFind, err := findMatchingFiles(testDir, []string{"some-md-file.md"}, []string{})
	assert.Equal(t, 1, len(singleFind), "expected to find a single file.")
	assert.NoError(t, err, "Should not fail with valid arguments")
	assert.True(t, strings.HasSuffix(singleFind[0], "/some-md-file.md"))

	tripleFind, err := findMatchingFiles(testDir, []string{"README.md"}, []string{})
	assert.Equal(t, 3, len(tripleFind), "expected to find a single file.")
	assert.NoError(t, err, "Should not fail with valid arguments")
	assert.True(t, strings.HasSuffix(tripleFind[0], "/README.md"))
	assert.True(t, strings.HasSuffix(tripleFind[1], "/sub-dir-a/nested-sub-dir-a/README.md"))
	assert.True(t, strings.HasSuffix(tripleFind[2], "/sub-dir-b/README.md"))
}

func TestFindRelevantFilesByExtension(t *testing.T) {
	testDir := getTestDir()
	mdFinds, err := findMatchingFiles(testDir, []string{}, []string{".md"})

	assert.NoError(t, err, "Should not fail with valid arguments")
	assert.Equal(t, 6, len(mdFinds), "Expected to find all test markdown files.")
	assert.True(t, strings.HasSuffix(mdFinds[0], "/README.md"))
	assert.True(t, strings.HasSuffix(mdFinds[1], "/some-md-file.md"))
	assert.True(t, strings.HasSuffix(mdFinds[2], "/sub-dir-a/CHANGELOG.md"))
	assert.True(t, strings.HasSuffix(mdFinds[3], "/sub-dir-a/nested-sub-dir-a/README.md"))
	assert.True(t, strings.HasSuffix(mdFinds[4], "/sub-dir-a/nested-sub-dir-a/some-other-md-file.md"))
	assert.True(t, strings.HasSuffix(mdFinds[5], "/sub-dir-b/README.md"))
}

func TestFindRelevantFilesByNameAndExtension(t *testing.T) {
	testDir := getTestDir()
	allFinds, err := findMatchingFiles(testDir, []string{"README", "CHANGELOG"}, []string{".md"})

	assert.NoError(t, err, "Should not fail with valid arguments")

	assert.Equal(t, 7, len(allFinds), "Expected to find all test markdown files.")
	assert.True(t, strings.HasSuffix(allFinds[0], "/sub-dir-a/README"))
	assert.True(t, strings.HasSuffix(allFinds[1], "/README.md"))
	assert.True(t, strings.HasSuffix(allFinds[2], "/some-md-file.md"))
	assert.True(t, strings.HasSuffix(allFinds[3], "/sub-dir-a/CHANGELOG.md"))
	assert.True(t, strings.HasSuffix(allFinds[4], "/sub-dir-a/nested-sub-dir-a/README.md"))
	assert.True(t, strings.HasSuffix(allFinds[5], "/sub-dir-a/nested-sub-dir-a/some-other-md-file.md"))
	assert.True(t, strings.HasSuffix(allFinds[6], "/sub-dir-b/README.md"))
}

func TestFindRelevantFilesByNameAndExtensionHasDuplicate(t *testing.T) {
	testDir := getTestDir()
	// We explicitely check we obtain duplicates: removing them should be done elsewehere.

	withDupes, err := findMatchingFiles(testDir, []string{"CHANGELOG.md"}, []string{".md"})

	assert.NoError(t, err, "Should not fail with valid arguments")

	assert.Equal(t, 7, len(withDupes), "Expected to find all test markdown files.")
	assert.True(t, strings.HasSuffix(withDupes[0], "/sub-dir-a/CHANGELOG.md"))
	assert.True(t, strings.HasSuffix(withDupes[1], "/README.md"))
	assert.True(t, strings.HasSuffix(withDupes[2], "/some-md-file.md"))
	assert.True(t, strings.HasSuffix(withDupes[3], "/sub-dir-a/CHANGELOG.md"))
	assert.True(t, strings.HasSuffix(withDupes[4], "/sub-dir-a/nested-sub-dir-a/README.md"))
	assert.True(t, strings.HasSuffix(withDupes[5], "/sub-dir-a/nested-sub-dir-a/some-other-md-file.md"))
	assert.True(t, strings.HasSuffix(withDupes[6], "/sub-dir-b/README.md"))
}

func TestNormalizeLinksToRoot(t *testing.T) {
	root := "/path/to/root/"
	filePath := "relative/file"
	relativeLinks := []string{"../back/one/level", "./path-in-same-dir", "same-dir-too", "sub-dir/hello", "/from/project-root"}
	normalizedLinks, err := normalizeLinksToRoot(root, filePath, relativeLinks)
	assert.NoError(t, err, "Should not fail on valid input")

	assert.Equal(t, []string{"back/one/level", "relative/path-in-same-dir", "relative/same-dir-too", "relative/sub-dir/hello", "from/project-root"}, normalizedLinks)
}

func TestNormalizeLinksToRootFailures(t *testing.T) {
	root := "/path/to/root/"
	filePath := "relative/file"
	relativeLinks := []string{"../../back/too/much", "./path-in-same-dir"}
	normalizedLinks, err := normalizeLinksToRoot(root, filePath, relativeLinks)
	assert.Error(t, err, "Should not fail on valid input")
	assert.Nil(t, normalizedLinks, "Nothing should be returned on failure")
}

func TestParseFiles(t *testing.T) {
	testDir := getTestDir()
	testFileA := filepath.Join(testDir, "some-md-file.md")
	testFileB := filepath.Join(testDir, "sub-dir-a/README")

	emptyParse, emptyError := parseFiles([]string{})

	assert.Empty(t, emptyParse)
	assert.NoError(t, emptyError)

	parsedFiles, err := parseFiles([]string{testFileA, testFileB})

	assert.NoError(t, err, "Expected no parsing error")
	assert.Equal(t, 2, len(parsedFiles), "expected one output for each input")
}
