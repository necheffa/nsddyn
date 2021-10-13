# nsddyn

nsddyn provides a secure method for achieving Dynamic DNS when using NSD as an authoritative DNS server.

## Abstract

The NSD authoritative name server by NLlabs does not support RFC 2136 or RFC 3007 as of this writing. \
(https://nlnetlabs.nl/projects/nsd/rfc-compliance/)

Several other third-party scripts can be found on the web for providing Dynamic DNS using NSD but \
I found them all to lack security. \
nsddyn aims to provide a secure alternative.

## Project Status

After much development, the project has finally entered its first beta. Although not ready for a v1.0.0 release, \
all the major functionality is in place, including RouterOS and GNU/Linux clients. Major milestones targeted before the \
first stable release include:

* v0.4.0 - Improve developer documentation, improve user documentation, transition from `$GOPATH` to use Go modules, and add vendoring for all Go dependencies.
* v0.5.0 - Improve the end user experience by providing installation scripts, buttoning up installation documentation, and addressing some outstanding issues 
related to usability.
* v0.6.0 - Improve robustness of test suite by increasing the coverage of unit and integration tests.

| Feature | Status |
| ------- | ------ |
| User Management | Complete |
| Server | Complete |
| Client | Complete |
| Protocol | Complete |

* The nsddyncc client works but assumes you are running a GNU userspace with cURL and jq installed and in your `$PATH`. \
A future release may include a client written purely in Go.

## Installing and Getting Started

nsddyn is currently a source only distribution. A Unix-like system is required to perform compilation. In particular, GNU Make and Google Go are required.

Starting with v0.3.3 nsddyn uses Go modules and vendoring to handle dependencies and the minimum Go version.
Earlier versions depended on the $GOPATH and should not be used.

### Compilation

Compiling nsddyn is fairly simple, just follow these steps:

1. `git clone https://gitlab.com/necheffa/nsddyn.git`
2. `cd nsddyn/`
3. `make`

This will result in a bin/ directory being created, located here will be the dynupd and nsddynum binaries.
Until the Makefile can be updated with an install target, these will need to be manually copied to your primary NSD server.

The client component of nsddyn can be found in the client/ directory. nsddyncc is the cURL client and is designed to be run from a typical GNU/Linux distro.
A RouterOS client can be found in nsddynrosc. It is not required to use both, choose which one best fits your environment and manually copy it to the client.

### Configuring the Install

Most components of nsddyn are able to use the $NSDDYN\_HOME environment variable to locate resource files. Although other mechanisms exist, where possible,
this environment variable should be the preferred way to specify file locations.

#### Server

An unprivileged user and group should be created, the installation scripts assume both are named `nsddyn` but motivated admins may manually change this.
On Debian the recommended command to do this is: `adduser --disabled-password --group --system --no-create-home --home /opt/nsddyn nsddyn`

Ensure nsddyn is compiled, see Compilation above. Then use the install target on make from the root of the cloned repo:

`make install`

This will create a directory hierarchy at `/opt/nsddyn/` along with symlinks to the dynupd.service unit and nsddynum binary into this hierarchy in system locations.
The admin will need to manually enable and start the dynupd service with `systemctl`.

The `$DYNUPD_ARGS` environment variable in /opt/nsddyn/etc/dynupd should at a minimum be modified to specify the zonefile and zone name to be managed by dynupd.

By default, dynupd will listen on localhost:8080 for client requests. The --addr option is provided to override the listen address and port: `dynupd --addr 192.0.2.2:1337`.
While any valid address:port combination may be specified, it is highly recommended to listen on a local loopback address and use a proxy to route public traffic to dynupd.
Currently, dynupd relies on a proxy such as nginx to provide TLS tunneling and rate limiting functionality.
When starting dynupd, only the --zone-file argument is required to specify the location of the forward-lookup zonefile dynupd should manage.
Currently dynupd does not support reverse-lookup zonefile updates and likely never will as in most cases where ISPs issue dynamic IP addresses, the reverse-lookup zones
are never delegated, so it would be meaningless to attempt to manage them with dynupd.

Because dynupd requires Unix filesystem permissions for reading and writing to the zonefile, it is recommended to create a subdomain to segregate dynamic records from
static records. The zonefile should be owned by the nsddyn user and the group nsd is running as, both the user and group should have read-write permissions to the file.

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
"username": "yourusername"
"password": "secret"
"hostnames": [ "host1", "host2" ]
"server": "dynupdhost"
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

nsddyn is comprised of 3 components:
* dynupd - A web app that provides an HTTP API for accessing the name server. Upon successful authentication, dynupd updates the zonefile and reloads the zone.
* A web client. While official clients will be provided, anyone can create their own. \
        A minimal client might take the form of a shell script wrapped around cURL. \
        More interesting might be a RouterOS script wrapped around the `/tool fetch` client.
* nsddynum - A command-line, swiss army knife style tool for managing the nsddynpasswd file.

nsddyn is a secure protocol for the following reasons:
* It is just HTTP and so may be tunneled over TLS for confidentiality.
* Clients are authenticated with a username and password, not just anyone can initiate a zone update. \
        Further, nsddynd limits what A records a client is allowed to update. \
        Passwords are stored as salted hashes to buy more time in the event hashes are leaked.
* Its about as simple as I could make it - less attack surface.

An astute reader will notice that a number of features are missing like rate limiting and permitted client IP ranges. \
nsddyn is intended to be run on the localloop interface while a battle tested server like Apache or Nginx acts as a proxy.

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

The Makefile contains `test` and `testresults` targets for executing the unit tests and reviewing the results.
Where possible, contributions should come with unit tests.

Integration testing is somewhat more difficult since most people don't want to install and configure a name server on their laptop.
To execute the integration tests, follow these steps:

* Compile nsddyn
* `cd` to the root of the distribution
* Execute `docker build -t nsddyntest -f test/dynupd.Dockerfile .` to build an image containing a preconfigured NSD daemon and dynupd listener
* Execute `docker run -d -t --name mytest nsddyntest` to instantiate a container based on the nsddyntest image
* Then, on the host system, execute `src/scripts/test/automated_integration.sh` to execute the tests against the running container
* Use `docker stop mytest` to shutdown the container, if changes are made to the dynupd binary, the image will need to be rebuilt

The `automated_integration.sh` script assumes you are running with a standard GNU userspace and have both `cURL` and `jq` in your `$PATH`.

The `test/nsddynum/run.sh` script assumes your current working directory is `test/nsddynum/` and performs behavioral tests on the nsddynum binary.

## Licensing

nsddyn is released under the terms of the GPLv3 license, a copy of the GPL is provided in the COPYING file located in the root of this repo.
