# Pipeviz

Pipeviz is a system for visualizing what you need to know as a software-driven organization: your development workflow, the state of your infrastructure, and more. It's a database, and visualization webapp rolled into one, informed by an ecosystem of agents and other message producers.

## Installation

Pipeviz requires that you either have a working Go environment, or that you use the provided Dockerfiles to do it for you. There are three basic cases:

* Hacking on pipeviz (frontend or backend): local Go environment.
* Hacking on message producers: you can get away with Docker.
* Actually running pipeviz: Docker, or whatever your hosting environment dictates.

If you're going the Docker route, the Dockerfiles have their own instructions. For a local Go environment, setting it up oughtn't be too difficult. [These instructions](http://www.golangbootcamp.com/book/get_setup) should get you there. Make sure you set your `$GOPATH` and update your `$PATH` with the `$GOPATH/bin` directory, as it recommends!

Once you have a working environment, you'll start by getting the package. Because pipeviz is currently a private repository, this can be little tricky. If you have a Github authentication token (for HTTPS cloning) already set up, then this will work:

```
go get github.com/tag1consulting/pipeviz
```

This is the standard method for fetching Go packages, and will work once the pipeviz repository is public. But if you don't have an HTTPS token set up, do this instead:

```
mkdir -p $GOPATH/src/github.com/tag1consulting
git clone git@github.com:tag1consulting/pipeviz $GOPATH/src/github.com/tag1consulting/pipeviz
```

This puts the repository in the same place as where `go get` would have dropped it. Whichever way you use, the next step is to `cd` into that directory and install:

```
cd $GOPATH/src/github.com/tag1consulting/pipeviz
make install
```

This will install several binaries into your `$GOPATH/bin` directory: `pvc`, `pvutil`, and `pipeviz` itself.

You're done! You can just run `pipeviz` with no arguments, and it will bind (on loopback only) to port 8008 for the webapp, and 2309 for message ingestion.
