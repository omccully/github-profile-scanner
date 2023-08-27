# GitHub Profile Scanner

Scans a GitHub profile to check if all repositories have a description, readme, and at least one image in the readme.

## Download and install from source

Requires `go` command line tool to compile and install the Go code.

```bash
git clone https://github.com/omccully/github-profile-scanner.git
cd github-profile-scanner
go install .
```

Then make sure the `%GOPATH%/bin` path is part of your PATH environmental variable.

## Usage

`github-profile-scanner [GitHub profile name]`

## Demo

![GitHub Profile Scanner Demo](/demo.gif)
