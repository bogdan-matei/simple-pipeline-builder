# Description

Simple Pipeliner Builder (SPB) is an utility tool that allows local definition of jobs/pipelines by using a plugin approach.

Plugins define actions that can be run by a job.

## How to use

Prerequisites:
* golang (1.18)
* Docker (currently using 20.10.12, API Version 1.41)

Vscode Configuration:
* `GO111MODULE=auto` needs to be set up at the workspace level
  * `GOPATH` must not be used as it's deprecated

The tool can be called directly by using `go run <main.go path>`. It supports the following flags:
* `-image` (string)-> specifies the image of the job; by default hub.docker.com is used
  * image value defaults to `golang:1.18.0-alpine3.15`
* `-d` (bool) -> debug flag that changes how the job is running
  * by default (no debug) the job runs as a `run container`
  * with default flag the job runs as a combination of `create/exec/debug container`
    * this is usefull when you can't identify which part of the command fails; the cli won't delete the job unless it succedes
* `-e/-env` (string array) -> allows the definition of multiple `key=value` strings that will be used as the env variables inside the container

> Important:
> * The job currently mounts ONLY the current directory via the `PWD` env variable.
> * The command needs to be the last element of the cli command


```bash
# Generic command
`go run main.go run build [-image <string>] [-e/-env <key=value>] [-d] "command <string>" `


# Example of cli usage to build current code project 
# Workdir is src/ folder of this repository
`go run main.go run build -e GOPATH -e GO111MODULE=auto -e GOOS=linux -e GOARCH=386 "go mod tidy && go build"`

# Example with debug
`go run main.go run build -e GOPATH -e GO111MODULE=auto -e GOOS=linux -e GOARCH=386 -d "go mod tidy && go build"`


# GOPATH is provided without value to unset the env var; it's deprecated but stil set by the default image 'golang:1.18.0-alpine3.15'
# GO111MODULE=auto is requried to run link the dependencies between the project
# GOOS=linux and GOARCH=386 is used like this because WSL runs on linux with x86_64 arch
# go mod tidy -> installs dependencies (go.sum) based on the go.mod file
# go build -> builds the code in an executable with default values

# Command to view the combinations between OS & ARCH
`go tool dist list`
```

## To do's

* Job run
  * create proper struct for storing data
  * create proper interface object for sub-commands
    * the interface needs to be implemented for function to run
  * restructure code for first try of the project
    * structure project and code in go-way
  * add -h to show existing flags and usage

* Next steps (goals)
  * Link job in a flow
    * define multiple jobs
    * define order between them (seq,paralel)
  * Allow input file feed
    * use cli flags to overwrite 
  * Import plugins