package mockrepo

import (
	"io/ioutil"
	"os"
	"testing"
)

// Some functions to provide a mock test environment to the checking logic.

// MockRepo creates a mock repository in a temporary directory
func MockRepo(t *testing.T) string {
	t.Helper()
	tempDir, err := ioutil.TempDir("", "mock-repo-*")
	if err != nil {
		t.Error(err)
	}
	// This method will be called from places that have imported this package as a dependency:
	// in Bazel, it means that its content will be present in a directory with the same name.
	// This is ugly as hell and will probably break outside of Bazel
	f, err := os.Open("../mockrepo/test-data.tar.gz")
	if err != nil {
		t.Error(err)
	}
	// TODO refactor test to generate fixtures instead of using an archive
	untar(f, tempDir)
	err = f.Close()
	if err != nil {
		t.Error(err)
	}
	return tempDir
}
