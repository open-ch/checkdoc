package markdown

import (
	"os"
	"path/filepath"
	"testing"

	blackfriday "github.com/russross/blackfriday/v2"
	"github.com/stretchr/testify/assert"
)

func getTestDir() string {
	workDir, _ := os.Getwd()
	return filepath.Join(workDir, "test-data")
}

func getTestAst(testFileName string) *blackfriday.Node {
	testFile := filepath.Join(getTestDir(), testFileName)
	ast, _ := ParseFileToAst(testFile)
	return ast
}

func TestParseFileToAst(t *testing.T) {
	testFile := filepath.Join(getTestDir(), "test-file.md-ext")
	ast, err := ParseFileToAst(testFile)
	assert.Nil(t, err, "Expected to successfully parse test file.")
	assert.NotNil(t, ast, "Expected to successfully parse test file.")
}

func TestExtractAllLinks(t *testing.T) {
	ast := getTestAst("test-file.md-ext")
	links := ExtractAllLinks(ast)

	assert.Equal(t, 11, len(links), "Expected 10 links. Is the Autolink extension enabled?")

	assert.Equal(t, "https://google.ch", string(links[0].Destination))
	assert.Equal(t, "https://open.ch", string(links[1].Destination))
	assert.Equal(t, "relative/internal", string(links[2].Destination))
	assert.Equal(t, "/absolute/internal", string(links[3].Destination))
	assert.Equal(t, "https://sqooba.io", string(links[4].Destination))
	assert.Equal(t, "nested/relative", string(links[5].Destination))
	assert.Equal(t, "/nested/absolute", string(links[6].Destination))
	assert.Equal(t, "../sibling", string(links[7].Destination))
	assert.Equal(t, "./sub-dir", string(links[8].Destination))
	assert.Equal(t, "mailto:julien@sqooba.io", string(links[9].Destination))
	assert.Equal(t, "#anchor-id", string(links[10].Destination))
}

func TestFilterLocalLinks(t *testing.T) {
	ast := getTestAst("test-file.md-ext")
	localLinks := FilterLocalLinks(ExtractAllLinks(ast))

	assert.Equal(t, 6, len(localLinks))
	assert.Equal(t, "relative/internal", string(localLinks[0].Destination))
	assert.Equal(t, "/absolute/internal", string(localLinks[1].Destination))
	assert.Equal(t, "nested/relative", string(localLinks[2].Destination))
	assert.Equal(t, "/nested/absolute", string(localLinks[3].Destination))
	assert.Equal(t, "../sibling", string(localLinks[4].Destination))
	assert.Equal(t, "./sub-dir", string(localLinks[5].Destination))
}
