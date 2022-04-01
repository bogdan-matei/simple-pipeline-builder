# Feature list

Simple Pipeline Builder (SPB) provides an easy to use CLI for common CI jobs (build/test/etc) for a pool of predefined languages and allows custom builds that follow an interface.
 
* cli for pipeline steps
  * well documentation
* immutable
  * functional programming
* plugin and play
  * easy to extend
  * minimal prerequisites to install
  * easy to use
  * easy to customize
* constraint enforcement
  * plugin can't be imported without tests
  * plugin without documentation can't be imported
* Optional:
  * workflow definition
  * auto scan of project configuration

## How To Use

`<executable> <cli flags/action> <flags> path_to_code_src`

## CLI

Reserved action words (e.g. import to import a plugin from some src)

### Flags for CLI

* help (-h)
* version (-v)
* import
* add
* remove

## Flags

Flags are use to determine the behaviour of the action. They act as input for the function that runs.

Consensus:
- full name of flag, is the flag with `--` prefix
- short name is set custom with `-char`

### Mandatory fields for action 

* image (-i) -> refers to base image without tag
* version (-v) -> version of plugin
* build-version (-b) -> refers to the tag you want to build for
* debug (-d) 
* command (-c)

### Optional fields for action (TBD)