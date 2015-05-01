# Pipeviz

## Installation

Pipeviz is, for now, structured as a standard executable Go program where, given a functioning local Go environment, simply running

```
go get github.com/sdboyer/pipeviz
```

will download and build an executable binary, placing it at `$GOPATH/bin/pipeviz`. However, because the repository is currently private, and because `go get` operates over http, it will require that you provide your Github credentials for http cloning. This can be a bit of a pain, so there’s an alternative:

```
$ cd $GOPATH/src
$ mkdir -p github.com/sdboyer
$ git clone git@github.com:sdboyer/pipeviz github.com/sdboyer/pipeviz
$ go get -f github.com/sdboyer/pipeviz
```

The `-f` flag will make `go get` ignore the fact that the repo is set to clone over ssh instead of http, and just update it that way instead. At which point, it should build you a pipeviz binary.

However, the main binary isn’t actually that useful yet. At this stage, what’s more useful is the `pvutil` binary that it can also build, which can then be used to create dot files representing the results of interpreting message fixtures, which can then be transformed by graphviz (dot, specifically) into visualizations.

To build pvutil, `cd` into the root of the pipeviz repository, then run `make; make install`. The `pvutil` binary should then be installed to `$GOPATH/bin`, at which point you can run `pvutil dotdump` to create the dot output file, or simply pipe it to dot: `pvutil dotdump -a | dot -Tpng > output.png`.

### Git log generation for messages

command used to generate the formatted git commit logs:

```
git log \
    --pretty=format:'"%H": {"parents":["%P"],"author": "%an <%ae>","date": "%ad","message": "%s"},' \
    $@ | \
    perl -pe 'BEGIN{print "{"}; END{print "[]}\n"}' | \
    perl -pe 's/},\[\]}/}}/'
```

...almost. note that this is broken because it doesn’t correctly break out multiple parent commits (a merge) into separate array elements - that’s gotta be done manually.
