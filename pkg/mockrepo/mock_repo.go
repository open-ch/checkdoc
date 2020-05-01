package mockrepo

import (
	"io/ioutil"
	"os"
	"osag/libs/untar"
)

// Some functions to provide a mock test environment to the checking logic.

// MockRepo creates a mock repository in a temporary directory
func MockRepo() string {
	// Ignore the error: if it fails things blow up shortly after anyway
	tempDir, _ := ioutil.TempDir("", "mock-repo-*")
	// This method will be called from places that have imported this package as a dependency:
	// in Bazel, it means that its content will be present in a directory with the same name.
	f, _ := os.Open("./mockrepo/test-data.tar.gz")
	untar.Untar(f, tempDir)
	f.Close()
	return tempDir
}
