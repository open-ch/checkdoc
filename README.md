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

## Installing from Github

For some reasons we need to clarify, cloning this repository and doing a `go install` from it works,
but a `go get -u github.com/open-ch/checkdoc` does not.

This may be due to [this issue](https://github.com/russross/blackfriday/issues/500) with the [blackfriday](https://github.com/russross/blackfriday) dependency,
but as long as a manual installation is easy we have other priorities.

Please follow the manual installation steps below.

### Manual Installation

```
# Move to the GOPATH source directory
cd `go env GOPATH`/src
mkdir -p github.com/open-ch/checkdoc
cd github.com/open-ch
git clone https://github.com/open-ch/checkdoc
cd checkdoc
go install
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

## CHANGELOG

See the [CHANGELOG](CHANGELOG.md) file.

## License

Please see the LICENSE file.
