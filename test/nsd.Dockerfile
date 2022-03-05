#
# nsd.Dockerfile sets up an NSD and dynupd-broker process for integration testing.
# A Docker volume should be mounted to /nsddyn/ in order to provide access to
# the dynupd-broker binary.
# A Docker volume should be mounted to /nsd/ in order to provide access to the
# zonefile.
#

FROM debian:11-slim

RUN apt-get update && apt-get -y upgrade && apt-get install -y nsd procps
RUN nsd-control-setup
RUN mkdir -p /nsddyn

RUN mkdir -p /nsd
COPY test/nsd.conf /etc/nsd/

ENV NSDDYN_HOME="/nsddyn"
ENV PATH="$NSDDYN_HOME/bin:$PATH"

COPY test/nsd_deploy /
CMD ./nsd_deploy

