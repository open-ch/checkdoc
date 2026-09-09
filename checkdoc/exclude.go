package checkdoc

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreFile is the optional file at the tree root listing paths that checkdoc
// does not inspect. It lives in the repository rather than in flags so that a
// local run and the CI run apply the same exclusions without having to repeat
// them on every invocation.
const IgnoreFile = ".checkdocignore"

// readExclusions returns the paths listed in IgnoreFile at treeRoot, cleaned and
// relative to it. A missing file excludes nothing.
//
// Entries are plain paths rather than patterns, and a directory excludes
// everything below it. That covers the case this exists for: trees of markdown
// that are content rather than documentation, which no README links to and which
// may link among themselves, so both checkdoc rules would report them.
func readExclusions(treeRoot string) ([]string, error) {
	// Read through an os.Root so the lookup stays inside the tree: IgnoreFile is a
	// fixed name, but it may well be a symlink pointing somewhere else.
	root, err := os.OpenRoot(treeRoot)
	if err != nil {
		return nil, fmt.Errorf("open tree root %s: %w", treeRoot, err)
	}
	defer root.Close()

	file, err := root.Open(IgnoreFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var exclusions []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if filepath.IsAbs(line) {
			return nil, fmt.Errorf("%s: %q must be relative to the tree root", IgnoreFile, line)
		}
		// Clean drops a trailing slash and a leading "./", so a directory may be
		// written either way.
		cleaned := filepath.Clean(line)
		if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("%s: %q does not point inside the tree root", IgnoreFile, line)
		}
		exclusions = append(exclusions, cleaned)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", IgnoreFile, err)
	}
	return exclusions, nil
}

// isExcluded reports whether relativePath is an excluded path itself or sits
// below an excluded directory.
func isExcluded(relativePath string, exclusions []string) bool {
	cleaned := filepath.Clean(relativePath)
	for _, exclusion := range exclusions {
		if cleaned == exclusion || strings.HasPrefix(cleaned, exclusion+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// dropExcluded removes the absolute paths that IgnoreFile excludes.
//
// Only discovery is affected: an excluded file is neither required to be linked
// to nor checked for dead links, but it still exists on disk, so links pointing
// at it from documents that are checked continue to resolve.
func dropExcluded(treeRoot string, absPaths, exclusions []string) ([]string, error) {
	if len(exclusions) == 0 {
		return absPaths, nil
	}
	kept := make([]string, 0, len(absPaths))
	for _, absPath := range absPaths {
		relativePath, err := filepath.Rel(treeRoot, absPath)
		if err != nil {
			return nil, fmt.Errorf("relate %s to tree root %s: %w", absPath, treeRoot, err)
		}
		if isExcluded(relativePath, exclusions) {
			continue
		}
		kept = append(kept, absPath)
	}
	return kept, nil
}
