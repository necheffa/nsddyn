# nsddyn

nsddyn provides a secure method for achieving Dynamic DNS when using NSD as an authoritative DNS server.

## Abstract

The NSD authoritative name server by NLlabs does not support RFC 2136 or RFC 3007 as of this writing. \
(https://nlnetlabs.nl/projects/nsd/rfc-compliance/)

Several other third-party scripts can be found on the web for providing Dynamic DNS using NSD but \
I found them all to lack security. \
nsddyn aims to provide a secure alternative.

## Installing and Getting Started

nsddyn is currently a source only distribution. A Unix-like system is required to perform compilation. In particular, GNU Make and Google Go are required.

Starting with v0.3.3 nsddyn uses Go modules and vendoring to handle dependencies and the minimum Go version.
Earlier versions depended on the $GOPATH and should not be used.

### Compilation

Compiling nsddyn is fairly simple, just follow these steps:

1. `git clone https://gitlab.com/necheffa/nsddyn.git`
2. `cd nsddyn/`
3. `make`
4. `make package`

Once the package is generated, it will need to be copied to your primary NSD server.

The client component of nsddyn can be found in the client/ directory. nsddyncc is the cURL client and is designed to be run from a typical GNU/Linux distro.
A RouterOS client can be found in nsddynrosc. It is not required to use both, choose which one best fits your environment and manually copy it to the client.

### Configuring the Install

Most components of nsddyn are able to use the $NSDDYN\_HOME environment variable to locate resource files. Although other mechanisms exist, where possible,
this environment variable should be the preferred way to specify file locations.

#### Server

1. Create an unpriviledged user and group for nsddyn. On Debian the recommanded command to do this is: `adduser --disabled-password --group --system --no-create-home --home /opt/nsddyn nsddyn`
2. Unpack the generated tarball (see compilation above). Make sure by default root owns everything.
3. Configure NSD with a new sub-domain, for example, dyn.example.com. Place the zone file (from here on: dyn.zone) in $NSDDYN\_HOME/etc/ and a symlink to it in /etc/nsd/. This is to allow for stronger sandboxing of dynupd.
4. Execute `chown nsddyn:nsd dyn.zone && chmod 0644 dyn.zone` so that only dynupd has write access.
5. Edit `$NSDDYN\_HOME/etc/nsddyn` to match the environment. At a minimum, ensure $NSDDYN\_HOME reflects the installation directory and that the -z and -n options on dynupd are set.
6. Execute `touch $NSDDYN\_HOME/etc/nsddynpasswd && chmod 0640 $NSDDYN\_HOME/etc/nsddynpasswd && chown root:nsddyn $NSDDYN\_HOME/etc/nsddynpasswd`.
7. Configure a reverse proxy if dynupd will be listening on localhost.
8. Install, enable, and start the dynupd systemd unit file.
9. Edit `$NSDDYN\_HOME/etc/dynupd-broker.json` to reflect the environment, minimally the ZoneName will need changed to match your new forward lookup zone for dynamic hosts.
10. Install, enable, and start the dynupd-broker systemd unit file.

It is recommended to run dynupd behind a reverse proxy like nginx or Apache. As a result, dynupd will listen for clients on localhost:8080 by default.
But this can be changed with the --addr option at start up.
Currently, dynupd relies on a reverse proxy configuration to provide TLS tunneling to protect authentication from prying eyes.

It is also recommended to use a specific sub-domain for dynamic hosts rather than try to force dynupd to manage your primary forward lookup zones.
This is to avoid problems with permissions that inevitably lead to poor security choices.

Currently dynupd does not support reverse-lookup zonefile updates and likely never will as in most cases where ISPs issue dynamic IP addresses, the reverse-lookup zones
are never delegated, so it would be meaningless to attempt to manage them with dynupd.

By default, dynupd-broker listens on localhost:8081, this can be changed from the dynupd-broker.json file which is read at startup by the broker.
The broker runs as root because by default NSD's nsd-control certificate and key are only readable by root. Whatever user/group is permitted to
execute nsd-control in your environment should be the user/group that the broker runs as; this can be configured in the systemd unit file.
Currently communication with the broker is not authenticated so it is recommended to run the broker on the same host as dynupd and NSD to
limit connections to localhost. Otherwise, it may be possible for a public facing broker to be abused and at least hammer reloads of your dynamic zone.

Both nsddynum and dynupd will look for the nsddynpasswd in the following locations in the following order: \
* Path specified by the --passwd-file option
* $NSDDYN\_HOME/etc/nsddynpasswd
* /usr/local/etc/nsddynpasswd

At this time, dynupd does not write to a specific log file, instead messages are logged to STDERR which are picked up by `journalctl`.

#### Client - nsddyncc

nsddyncc is the cURL client and is meant for installation on Unix workstations and servers. It expects a typical Unix userspace and depends on cURL and jq.
The --help option can be used to find argument details and configuration options.
Copy the client/nsddyncc script to a sensible location, such as /usr/local/sbin/, and create a cronjob to execute the client as an unprivileged user.

nsddyncc relies on a JSON formatted configuration file and will search in the following locations in order until nsddynccrc is found:
* Path specified by the --file option
* $NSDDYN\_HOME/etc/nsddynccrc
* /usr/local/etc/nsddynccrc

nsddynccrc should take the following form:
```
{
"username": "yourusername",
"password": "secret",
"hostnames": [ "host1", "host2" ],
"server": "dynupdhost",
"port": "8080"
}
```
Note that "hostnames" is always given as a list, even if only a single hostname will be updated.
nsddynccrc should be owned by the unprivileged user it will run as in cron and be chmod 0400 so that group and world are unable to read it.

#### Client - nsddynrosc

The RouterOS client is built on the /tool fetch utility and has been tested on RouterOS 6.46, but any version of RouterOS supporting /tool fetch should be compatible.
To install, copy the client/nsddynrosc script, edit the SERVER, PASSWD, USERNAME, and HOSTS fields. Optionally, the default polling interval of 4 hours may be changed.
Use sftp to copy the client script up to the RouterOS device and then use the /import file-name command to import the copied script.

## Contributing and Developer Information

### Development Model

nsddyn uses a somewhat continuous strategy. Ideally, `master` should be kept clean with commit squashes and always buildable.
The `devel` branch serves as an integration branch. Tags are used to track releases and points of interest.

### Design

nsddyn is comprised of multiple components:
* dynupd - A web app that provides an HTTP API for accessing the name server. Upon successful authentication, dynupd updates the zonefile requests the dynupd-broker reload the zone.
* dynupd-broker - A web app that reloads the configured zone. This allows for a seporation of priviledge between NSD and dynupd.
* A web client. While official clients will be provided, anyone can create their own. \
        A minimal client might take the form of a shell script wrapped around cURL. \
        More interesting might be a RouterOS script wrapped around the `/tool fetch` client.
* nsddynum - A command-line, swiss army knife style tool for managing the nsddynpasswd file.

### Protocol Description

Clients will initiate an update by sending an HTTP POST with the content type set to `application/json` to `https://www.example.com/api/dynupd`.
Note that the URI (e.g. /api/dynupd) may be overridden with the --uri flag, /api/dynupd is just the default if --uri is not specified.
The data sent will be of a JSON object taking the following form:
```
{
    "version": "0.1.0",
    "username": "clientusername",
    "password": "plaintextpassword",
    "ipaddr": "desiredipaddress",
    "hosts": [ "hostname1", "hostname2" ]
}
```

Note that `hosts` may simply be an array containing a single element but will always be an array and not a scalar. This provides maximum flexibility while
limiting edge cases to be handled.
One might find it odd to explicitly specify `ipaddr` as well as one could infer this from the HTTP session data.
However, this limits client flexibility, one might wish to use a proxy for updating for some odd reason.
The version of the nsddyn client is included in the request so that the protocol may be versioned.

Once dynupd receives the request it will attempt to authenticate the connected client.
If successful, dynupd will update the zonefile and reload the zone.
This is accomplished by calling `nsd-control`; finally, dynupd returns a status code to the client.

nsddyn will always return a status as a JSON object with the following form:
```
{
    "version": "0.1.0",
    "code": "codenumber"
}
```

The following status codes may be returned:
* 200 - Success. The request was authenticated and applied.
* 400 - The request was malformed, dynupd rejected it.
* 403 - Authentication failed, nsddynd didn't agree with your provided username or password.
* 405 - A bad method was used to connect to the server, we only accept HTTP POSTs.
* 418 - Account authentication succeeded, but permitted hosts authentication failed.
* 500 - Account and permitted hosts authentication succeeded, but something failed when updating the zone. Probably not the client's fault.

Notice that the nsddyn server will always return its version number in the response, this is so the protocol may be versioned.

### Running Tests

Before tests can be run for the first time `scripts/test/integrate_setup` needs to be executed to create Docker images.
Afterwards, all tests can be executed with `make test` and unit test coverage can be viewed with `make testcoverage`.

Unit tests can be executed individually with `make unittest` while integration tests can be executed individually with
`make integrationtest`. The `make test` target simply executes both suites.

Most people don't want to install and configure a name server on their laptop, integration testing uses Docker
to host an NSD instance alongside the development build of nsddyn.
Since integration testing relies on Docker images, it is a best practice to update images when migrating between major/minor versions of nsddyn.

Where possible, contributions should come with enhancements to the test suite to cover new features and fixes.

## Licensing

nsddyn is released under the terms of the GPLv3 license, a copy of the GPL is provided in the COPYING file located in the root of this repo.
