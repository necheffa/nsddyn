#
# Set up a Docker container for dynupd as part of integration testing.
#
# I could eventually spinup multiple containers and use a shared volume to
# provide the illusion that NSD and dynupd are running on the same system.
# But using a shell script is the quick and dirty band aid I need to press on
# early in development.
#

# NOTE: I tried using alpine linux but ran into trouble because it links against muslc not glibc
FROM debian:10-slim
USER root

# bash is a little heavy but it will make scripting easier for me
#RUN /sbin/apk add --no-cache nsd bash
RUN apt-get update && apt-get -y upgrade && apt-get install -y nsd
RUN useradd nsddyn
RUN mkdir -p /nsddyn/bin
RUN mkdir -p /nsddyn/etc
COPY test/nsddynpasswd /nsddyn/etc/
RUN chown -R nsddyn /nsddyn

#
# Prep the NSD install
#
# installing the nsd package creates this user for me
#RUN adduser -D nsd
RUN mkdir -p /etc/nsd/zones
COPY test/example.com.zone /etc/nsd/zones/
COPY test/nsd.conf /etc/nsd/

#
# This is the stuff I'd need to do if I only ran dynupd from this container.
#
#USER nsddyn
#WORKDIR /nsddyn
ENV NSDDYN_HOME="/nsddyn"
ENV PATH="$NSDDYN_HOME/bin:$PATH"
COPY src/bin/dynupd /nsddyn/bin/
COPY src/bin/nsddynum /nsddyn/bin/
#CMD ["dynupd", "--debug", "--addr", "0.0.0.0:8080"]

COPY test/dynupd_test_deploy /
CMD ./dynupd_test_deploy
