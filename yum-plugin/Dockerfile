# Dockerfile for testing yum plugin.
# With a pipeserv container already running
#
# Build with
# docker build -t halstead/yum-plugin .
#
# Run with
# docker run --rm -ti --name yumtest -h yumtest.pipeviz.org --link pipeserv:pipeserv halstead/yum-plugin /bin/bash
#
# Then yum install something to send messages.


FROM centos:latest

COPY pipeviz.conf /etc/yum/pluginconf.d/pipeviz.conf
COPY pipeviz.py /usr/lib/yum-plugins/pipeviz.py

CMD /bin/bash
