## Realistic fixtures
#### JSON that tells a story

This directory contains valid pipeviz JSON messages. The intention is that, when sent to a pipeviz daemon in filename lexical order, they simulate a realistic series of messages that a real pipeviz instance might receive, in the case of someone setting up pipeviz around some existing software/infrastructure.

The files are named with a 3-digit leading pattern, followed by a brief description. The three-digit space ensures there’s copious space in which to interleave new messages later without needing to rename anything.

Being that JSON is rather averse to inline documentation, this file serves as a location to document the role each message plays in this story, in order.

In particular, the list also indicates the imagined client/source that would logically have access to all the data contained in the message. This client-speculation, along with the logical ordering of the messages themselves, are the two ways in which this set of fixtures aims to be “realistic.”

* **000-commits** (scan-repo) - since we’re imagining an existing git repository for the app that is our focal point, this is the most logical place to begin: a dump of all commits in the repository. (pipeviz’ own commit history is used here)
* **004-labels** (scan-repo) - follows pretty directly on the heels of 000-commits; this identifies the labels (branches and tags) that are present in the repository (at least, in the origin repo - branches still need to be namespaced). These are extracted just by scanning a repo.
* **007-test-stats** (scan-github) - scanner that slurps data from github to reports test pass/failure status as recorded by, say, travis-ci.
* **010-prod** (pvc) - manual reporting about the existence of all three of the servers involved in constituting “prod” - `prod-web01`, `prod-web02`, and `prod-db01.`
* **015-prod1-pkg-ls** (scan-yum) - a yum-based scanner, reports apache httpd binary and libphp5.so library on prod-web01.
