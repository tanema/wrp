# wrp - git based project dependency management
This is not a good solution, this is a quick & dirty solution. I wrote this mainly
to quickly manage lua libraries for Love2d where most libraries want you to download
the repo and copy and paste a single file into your project.

I hated doing this so I wrote wrp instead.

# CLI

Run `wrp --help` for command help

# What this does
- Will download any git repo by url
- Check out a tag/branch/revision
- Pick out the files that you want
- Write them to disk
- Update wrp.yaml with pinned versions

# What this does not do (not sure if I care to implement them)
- Dependency tree of any sort. it will not solve your version issues
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
