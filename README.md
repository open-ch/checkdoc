# checkdoc

`checkdoc` is a tool that helps you assess if a markdown documentation tree is in good shape, at least from a linking perspective.

Its main goal is to enforce minimal quality standards in a repository's documentation.

It will tell you if:

  - Markdown files are not referenced (directly or through other files) from a readme in the root directory
    of a repository
  - There are broken internal links

## Sample Usage

Used on this repository, checkdoc yields the following:
```
$ checkdoc verify
INFO Running verify on tree root /tmp/checkdoc
INFO Considering basenames [] and extensions [.md]
DEBU Found 1 nodes at:
DEBU    README.md:
INFO Checking for orphaned documents...
INFO No orphans found.
INFO Checking for dead links...
ERRO Located some files with dead links:
ERRO    README.md
ERRO       CHANGELOG.md
ERRO Verify failed on tree root /tmp/checkdoc
```

As shown above, it detects that we have a dead link to a non-existing file.

## Excluding Paths

Some markdown in a repository is content rather than documentation: pages served
by an application, generated exports, imported fixtures. Nothing links to it from
a README, and it may link among its own pages, so both checks would report it
even though the tree is healthy.

List those paths in a `.checkdocignore` file at the root of the tree being
checked:

```
# One path per line, relative to this file. A directory excludes everything
# below it. Lines starting with '#' are comments.
applications/customer-portal/apps/backends/markdown-proxy/content/
docs/generated-api-reference.md
```

Entries are plain paths, not glob patterns, and must stay inside the tree. The
file lives in the repository rather than in flags so that a local run and the CI
run apply the same exclusions; checkdoc logs the paths it loaded on every run.

Exclusion only affects discovery. An excluded file is neither required to be
linked to nor checked for dead links, but it is still on disk, so links pointing
into an excluded tree from documents that *are* checked keep resolving.

Files matched by a `.gitignore` are skipped as well, unless
`--respect-git-ignore=false` is passed.

## Installation

```
go install github.com/open-ch/checkdoc@latest
```

Then run it with `checkdoc`, assuming your `$GOPATH/bin` is on your `PATH`. You should see something along these lines:
```
checkdoc
A markdown documentation validator intended to enforce a healthy documentation in settings such as a fat repo.

Usage:
  checkdoc [command]

Available Commands:
  help        Help about any command
  verify      Runs sanity checks on the documentation

Flags:
  -h, --help           help for checkdoc
  -r, --root string    Path to the root of the markdown documentation hierarchy to validate (default ".")
  -g, --use-git-root   from the given root, fall back to the repository's root. This will cause checkdoc to fail if --root is not pointing to a repository. (default true)

Use "checkdoc [command] --help" for more information about a command.
```

## Note For GitHub Readers

While the content of this module is managed in an internal repository,
you may still submit PR's.

## License

Please see the LICENSE file.
