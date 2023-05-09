package checkdoc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/open-ch/go-libs/fsutils"
	"github.com/open-ch/go-libs/mdutils"

	blackfriday "github.com/russross/blackfriday/v2"
)

// parsedAST is a tuple of an absolute path to a markdown file, and its parsed abstract syntax tree (AST).
type parsedAST struct {
	AbsPath   string
	ParsedAST *blackfriday.Node
}

// RelativeAST is a tuple of a relative path to a markdown file, its parsed representation
// as well as relative links (normalized to the root!) found in links
type RelativeAST struct {
	RelativePath string
	ParsedAST    *blackfriday.Node
	// All local links found in the markdown. "Absolute" links (starting with a '/') are relative to the root,
	// relative links are relative to RelativePath.
	// Relative links may contain things like '.' and '..'
	LocalRelativeLinks []blackfriday.LinkData
}

// LinkGraphNode represents a markdown file, by its relative path from a root,
// as well as the local links it contains, normalized relative to the root.
// Ie, there should not be any . or .. in any path anymore.
type LinkGraphNode struct {
	RelativePath                 string            // Path of the file from the root
	ParsedAST                    *blackfriday.Node // The parsed AST from the file referred by this node
	NormalizedLocalRelativeLinks []string          // links to other files, relative from the root
}

// BuildLinkGraphNodes takes a path to a directory, the content of which will be explored recursively.
// Markdown files will be searched for based on the specified baseNames or fileExtensions:
// both can be used together, ie, search for {README, CHANGELOG} and "*.md".
// All matching files will have a corresponding node, but they may well have internal links that point to files
// that do not have a corresponding node, or files that may not even exist.
// TODO deduplicate when/where relevant (if matches occur via basename and extension)
func BuildLinkGraphNodes(
	treeRoot string,
	baseNames []string,
	fileExtensions []string,
	respectGitIgnore bool,
) ([]LinkGraphNode, error) {
	// Input validation
	if len(baseNames) == 0 && len(fileExtensions) == 0 {
		return nil, fmt.Errorf("need to specify at least one base name or extension")
	}

	if !filepath.IsAbs(treeRoot) {
		return nil, fmt.Errorf("treeRoot must be absolute, was: %s", treeRoot)
	}

	// Get to work finding relevant files
	results, err := findMatchingFiles(treeRoot, baseNames, fileExtensions)
	if err != nil {
		return nil, err
	}

	var filteredResults []string
	if respectGitIgnore {
		// Filter out anything that matches a gitignore (if required)
		// Respect the gitignore
		gitIgnore, err := gitignore.NewRepository(treeRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to build up a gitignore from a git repository. "+
				"Is treeRoot pointing to a git repository? It was: %s - %s", treeRoot, err)
		}
		for _, path := range results {
			// match is nil if the path does not match the gitignore
			match := gitIgnore.Absolute(path, false)
			if match == nil {
				filteredResults = append(filteredResults, path)
			}
		}
	} else {
		filteredResults = results
	}

	return parseFilesAndBuildGraph(filteredResults, treeRoot)
}

func parseFilesAndBuildGraph(absFilePaths []string, treeRoot string) ([]LinkGraphNode, error) {
	parsedFiles, err := parseFiles(absFilePaths)
	if err != nil {
		return nil, err
	}

	// We already checked the root is an absolute path. Now we make sure it ends with a slash.
	sanitizedRoot := strings.TrimSuffix(treeRoot, "/") + "/"

	var graphNodes []LinkGraphNode
	for _, parsedFile := range parsedFiles {
		filePathFromTreeRoot := strings.TrimPrefix(parsedFile.AbsPath, sanitizedRoot)
		normalizedRelLinks, err :=
			normalizeLinksToRoot(
				sanitizedRoot,
				filePathFromTreeRoot,
				keepLinksAsStrings(
					mdutils.FilterLocalLinks(
						mdutils.ExtractAllLinks(parsedFile.ParsedAST)),
					true,
				),
			)

		if err != nil {
			return nil, fmt.Errorf("failed to normalize relative links in %s from root %s:%s", normalizedRelLinks, treeRoot, err)
		}

		graphNodes = append(graphNodes,
			LinkGraphNode{
				RelativePath:                 filePathFromTreeRoot,
				ParsedAST:                    parsedFile.ParsedAST,
				NormalizedLocalRelativeLinks: normalizedRelLinks,
			})
	}

	return graphNodes, nil
}

func keepLinksAsStrings(linkDatas []blackfriday.LinkData, trimAnchors bool) []string {
	var toRet []string
	for _, linkData := range linkDatas {
		// TODO validate existence of Anchor at destination?
		var linkStr = string(linkData.Destination)

		if trimAnchors {
			linkStr = strings.Split(linkStr, "#")[0]
		}

		toRet = append(toRet, linkStr)
	}
	return toRet
}

// normalizeRelativeAstTo normalizes the passed relative links according to the treeRoot, based on the filePath
// where they were found.
func normalizeLinksToRoot(treeRoot string, filePath string, relativeLinks []string) ([]string, error) {
	absFilePath := filepath.Join(treeRoot, filePath)
	// We are interested in building links relative to the directory containing the file.
	absDirPath := filepath.Dir(absFilePath)
	var normalizedRelativePaths []string
	for _, relativeLink := range relativeLinks {
		projectAbsoluteLink := filepath.Join(absDirPath, relativeLink)
		// Found a reference starting with "/", where "/" refers to the project root.
		if strings.HasPrefix(relativeLink, "/") {
			projectAbsoluteLink = filepath.Join(treeRoot, relativeLink)
		}

		absoluteNormalizedPath, err := filepath.Abs(projectAbsoluteLink)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize relative links from root %s, file %s:%s", treeRoot, filePath, err)
		}
		if !strings.HasPrefix(absoluteNormalizedPath, treeRoot) {
			return nil, fmt.Errorf("relative link %s points outside of the tree root %s for file %s", string(relativeLink), treeRoot, filePath)
		}
		relativeNormalizedPath := strings.TrimPrefix(absoluteNormalizedPath, treeRoot)
		normalizedRelativePaths = append(normalizedRelativePaths, relativeNormalizedPath)
	}
	return normalizedRelativePaths, nil
}

func findMatchingFiles(treeRoot string, baseNames []string, fileExtensions []string) ([]string, error) {
	var collectedFiles []string
	for _, baseName := range baseNames {
		results, err := fsutils.SearchByFileName(treeRoot, baseName)
		if err != nil {
			return nil, err
		}

		collectedFiles = append(collectedFiles, results...)
	}
	for _, ext := range fileExtensions {
		results, err := fsutils.SearchByExtension(treeRoot, ext)
		if err != nil {
			return nil, err
		}
		collectedFiles = append(collectedFiles, results...)
	}
	return collectedFiles, nil
}

// parseFiles parses the filePaths, expecting them all to point to markdown files.
// It returns a map of the paths to their corresponding AST's.
func parseFiles(mdFilePaths []string) ([]*parsedAST, error) {
	var asts []*parsedAST
	for _, mdFilePath := range mdFilePaths {
		if !filepath.IsAbs(mdFilePath) {
			return nil, fmt.Errorf("will not parse a relative path: %s", mdFilePath)
		}
		ast, err := mdutils.ParseFileToAst(mdFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse markdow file %s: %s", mdFilePath, err)
		}
		asts = append(asts, &parsedAST{mdFilePath, ast})
	}
	return asts, nil
}
