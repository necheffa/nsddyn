#
# dynupd.Dockerfile sets up dynupd for integration testing.
# A Docker volume should be mounted to /nsddyn/ in order to provide
# access to the dynupd binary.
# A Docker volume should be mounted to /nsd/ in order to provide access
# to the zonefile.
#

FROM debian:11-slim

RUN mkdir -p /nsddyn
RUN mkdir -p /nsd

ENV NSDDYN_HOME="/nsddyn"
ENV PATH="$NSDDYN_HOME/bin:$PATH"

COPY test/dynupd_deploy /
CMD ./dynupd_deploy

