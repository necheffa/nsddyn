#!/usr/bin/env bash

set -eu -o pipefail

NSDDYN_SRC="$(dirname $(dirname $(dirname $(realpath $0))))"

UM="$NSDDYN_SRC/bin/nsddynum"

if ! [ -f "$UM" ]; then
    echo "Did you build nsddyn?"
    echo "Can't run tests without the build, exiting."
    exit 1
fi

WORKDIR=$(mktemp -d)

function cleanup {
    rm -rf $WORKDIR
}
trap cleanup EXIT

FAILED_TEST="no"

# we don't want to modify files in version control.
# work with a copy instead.
cp ./* $WORKDIR/

# Abuse GNU script to get around a problem where the Go terminal.ReadPassword() function
# tries to use tty control to disable echoing back the "typed" password but since this
# is a script not an interactive ptty, we end up with "bad ioctl" errors.
# Instead of writing a "typescript" file, we write to /dev/null.
# We need sed to help cleanup some ugly - for whatever reason GNU script uses DOS line terminators instead of Unix...
RET1=$(printf "password\n" | script -q -c "$UM adduser -p $WORKDIR/badperms -u alex host1 host2" /dev/null)
CMP1=$(echo "$RET1" | tail -1 | awk '{$1=$2=""; print $0}' | sed -e 's/^[ \t]*//' | sed -e 's/\r$//')
VALID1="nsddynum: permissions on nsddynpasswd too permissive, recommend chmod 0600."

if [ "$VALID1" != "$CMP1" ]; then
    echo "Failed permissive nsddynpasswd file perms test on adduser command."
    FAILED_TEST="yes"
fi

RET2=$(script -q -c "$UM deluser -p $WORKDIR/badperms -u alex" /dev/null)
CMP2=$(echo "$RET2" | tail -1 | awk '{$1=$2=""; print $0}' | sed -e 's/^[ \t]*//' | sed -e 's/\r$//')
VALID2="nsddynum: permissions on nsddynpasswd too permissive, recommend chmod 0600."

if [ "$VALID2" != "$CMP2" ]; then
    echo "Failed permissive nsddynpasswd file perms test on deluser command."
    FAILED_TEST="yes"
fi

RET3=$(printf "n\nn\n" | script -q -c "$UM moduser -p $WORKDIR/badperms -u alex" /dev/null)
CMP3=$(echo "$RET3" | sed -n 3p | awk '{$1=$2=""; print $0}' | sed -e 's/^[ \t]*//' | sed -e 's/\r$//')
VALID3="nsddynum: permissions on nsddynpasswd too permissive, recommend chmod 0600."

if [ "$VALID3" != "$CMP3" ]; then
    echo "Failed permissive nsddynpasswd file perms test on moduser command."
    FAILED_TEST="yes"
fi

RET4=$(printf "password\n" | script -q -c "$UM adduser -p $WORKDIR/nonexistent -u alex host1 host2" /dev/null)
CMP4=$(echo "$RET4" | tail -1 | awk '{$1=$2=""; print $0}' | sed -e 's/^[ \t]*//' | sed -e 's/\r$//')
VALID4="nsddynum: creating nsddynpasswd file at: $WORKDIR/nonexistent"

if [ "$VALID4" != "$CMP4" ]; then
    echo "Failed non-existent nsddynpasswd file test on adduser command: warning message."
    FAILED_TEST="yes"
fi

if [ ! -f "$WORKDIR/nonexistent" ]; then
    echo "Failed non-existent nsddynpasswd file test on adduser command: creating the file."
    FAILED_TEST="yes"
fi

if [ "$FAILED_TEST" == "no" ]; then
    echo "All tests passed."
fi

