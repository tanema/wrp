# wrp - git based project dependency management
This is a quick & dirty solution. I wrote this mainly to quickly manage lua
libraries for Love2d where most libraries require you to download the repo,
then copy and paste a single file into your project.

I hated doing this so I wrote `wrp` instead.

# CLI

Run `wrp` to install all of your dependencies, but otherwise:

```
Usage:
  wrp [flags]
  wrp [command]
Available Commands:
  add         add a new dependency to project
  help        Help about any command
  init        init config
  rm          rm a dependency from project
  update      update a dependency in project
  version     Print the version number of wrp
Flags:
  -h, --help   help for wrp
```

# What this does
- Will download any git repo by url
- Check out a tag/branch/revision
- Pick out the files that you want
- Write them to a chosen destination folder
- Update `wrp.yaml` with pinned versions

# What this does not do (not sure if I care to implement them)
- Dependency tree of any sort. It will not solve your version issues
- Nested manifests

# How does it work

Add a `wrp.yaml` file to your project that looks like the following:

```yaml
destination: vnd # what folder you will put your dependencies in
dependencies:
  github.com/kikito/anim8: # github repo to fetch at a tag
    tag: v2.3.1
    pick: [anim8.lua] # files to save from that repo
  github.com/tanema/Moan.lua: # commit hash
    hash: 404923c672f76b82ec9dfead7077fa9f289be9bd
    pick: [Moan] # save a folder
  github.com/kikito/beholder.lua: # branch
    branch: demo
```

# Installation

```
> git clone git@github.com:tanema/wrp.git
> cd wrp
> go install .
```
