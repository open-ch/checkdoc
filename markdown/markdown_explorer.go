package markdown

import (
	"io/ioutil"
	"regexp"

	blackfriday "github.com/russross/blackfriday/v2"
)

// Using 'var' as 'const' not allowed here.
// This is a private field, so no one should be messing with it.
var urlPrefixMatcher = regexp.MustCompile(`://`)
var mailToMatcher = regexp.MustCompile(`^mailto:`)
var sameFileAnchorMatcher = regexp.MustCompile(`^#`)

// ParseFileToAst parses a file living at the path specified at marcdownFile
// and returns an abstract syntax tree
func ParseFileToAst(markdownFile string) (*blackfriday.Node, error) {
	input, err := ioutil.ReadFile(markdownFile)
	if err != nil {
		return nil, err
	}
	parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.Autolink))

	return parser.Parse(input), nil
}

// ExtractAllLinks will extract all links from the passed ast.
func ExtractAllLinks(ast *blackfriday.Node) []blackfriday.LinkData {
	var links []blackfriday.LinkData
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		// The visitor is called twice: once when the node is first visited, with entering = 'true',
		// then again once all the children are done, with entering = 'false'.
		// We check for links and collect them upon first visit.
		if entering && node.Type == blackfriday.Link {
			links = append(links, node.LinkData)
		}
		// We want to visit every node.
		return blackfriday.GoToNext
	})
	return links
}

// FilterLocalLinks will extract all "local" links, ie, links pointing to the local file system
// and not starting with "http[s]://...". This is done by looking at the link's destination.
// Links of the form "/absolute-link", "../sibling-dir/something", "sub-dir/something", nothing else.
// Note on anchors: links pointing to anchors in the same file will not be returned. Links pointing to other files
// while also containing an anchor (ie, in the form <path_to_file>#<anchor-name> are returned.
func FilterLocalLinks(links []blackfriday.LinkData) []blackfriday.LinkData {
	var localLinks []blackfriday.LinkData
	for _, link := range links {
		// We eliminate links starting with something like <prefix>://, ie http://, ftp://,
		// or with a mailto:
		if urlPrefixMatcher.Match(link.Destination) ||
			mailToMatcher.Match(link.Destination) ||
			sameFileAnchorMatcher.Match(link.Destination) {
			continue
		}
		// At this point we assume the link to be local: any corner cases will blow up somewhere else
		// and will be addressed in due time.
		localLinks = append(localLinks, link)
	}
	return localLinks
}
