# Pipeviz

## Prototype installation

silly things with react-commonjs seem to require particular placement of bower files

```
npm install
bower install
ln -s bower_components src/bower_components
```

command used to generate the formatted git commit logs:

```
git log \
    --pretty=format:'"%H": {"parents":["%P"],"author": "%an <%ae>","date": "%ad","message": "%s"},' \
    $@ | \
    perl -pe 'BEGIN{print "{"}; END{print "[]}\n"}' | \
    perl -pe 's/},\[\]}/}}/'
```

...almost. note that this is broken because it doesn’t correctly break out multiple parent commits (a merge) into separate array elements - that’s gotta be done manually.
