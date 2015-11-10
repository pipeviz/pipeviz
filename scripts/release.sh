#!/usr/bin/env bash

TAG=$(git describe --tags --exact-match) || { echo "no tag exists at current commit"; exit 1; }

github-release release -u pipeviz -r pipeviz -t $TAG -p || { echo "failed to create release"; exit 1; }

# always must run this from the pipeviz repo root
make build-all

for TGT in `ls dist`; do
    zip -j pipeviz-${TGT}.zip dist/${TGT}/*
    github-release upload -u pipeviz -r pipeviz -t $TAG -n pipeviz-${TGT}.zip -f pipeviz-${TGT}.zip || echo "upload of pipeviz-$TGT failed"
done
