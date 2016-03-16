# Github History

Generates an HTML files summarizing releases created for a Github project.

## Requirements
Git 2.*

## Setup
```sh
go get github.com/robstrong/go-gh-history
```
In the go-gh-history dir
```sh
go build
```

## Usage
To get help on the command, use 
```sh
./go-gh-history --help
usage: go-github-history-linux-amd64 [<flags>] <gen-type> <repo>

Flags:
  --help               Show help.
  -t, --token-path=TOKEN-PATH  
                       Path to file containing token
  -o, --out="gh-history.html"  
                       HTML output file
  --template=TEMPLATE  HTML template file, default is 'releases.html' or 'issues.html' depending on gen-type
  --verbose            Enable verbose output

Args:
  <gen-type>  Generation type ('releases' or 'issues')
  <repo>      Github Repository in the format 'owner/repository'
```


## TODO
- Clean up code a bit
- Create Github token instead of requiring user to create one
