# nsddyn

nsddyn provides a secure method for achieving Dynamic DNS when using NSD as an authoritative DNS server.

## Abstract

The NSD authoritative name server by NLlabs does not support RFC 2136 or RFC 3007 as of this writing. \
(https://nlnetlabs.nl/projects/nsd/rfc-compliance/)

Several other third-party scripts can be found on the web for providing Dynamic DNS using NSD but \
I found them all to lack security. \
nsdyn aims to provide a secure alternative.

## Project Status

Currently the project is under heavy development. As the project evolves, look here for a more detailed list \
of what is and is not working.

## Development Model

nsddyn uses a relaxed Gitflow strategy. That is, `master` should always be buildable, stable, and only contain tagged releases.
`devel` acts as an integration branch and serves as a parent to any number of feature branches.
Given the size of the project, release branches are overkill.

## Design

nsddyn is comprised of 3 components:
* dynupd - A Flask webapp that provides an HTTP API for accessing the name server.
* nsddynd - A Python daemon used to perform forward zone updates. While dynupd could \
        handle this itself, a concious design decision was made to seporate these tasks so \
        that the HTTP API has limited control over zone updates.
* A web client. While official clients will be provided, anyone can create their own. \
        A minimal client might take the form of a shell script wrapped around curl. \
        More interesting might be a RouterOS script wrapped around the `/tool fetch` client.

nsddyn is a secure protocol for the following reasons:
* It is just HTTP and so may be tunneled over TLS for confidentiality.
* Clients are authenticated with a username and password, not just anyone can initiate a zone update. \
        Further, nsddynd limits what A records a client is allowed to update. \
        Passwords are stored as salted hashes to buy more time in the event hashes are leaked.
* The seporation of roles between dynupd and nsddynd makes it harder for abuse of the public facing API to \
        manipulate the zone of your domain since an attacker cannot simply exploit a buffer overflow to get shell access \
        and arbitrarily write to the zone files. Instead they must craft malicous messages to pass to nsddynd which has \
        authority to edit a zone. This, I hope, is much harder to do.
* Its about as simple as I could make it - less attack surface.

### Protocol Description

Clients will initiate an update by sending an HTTP POST with the content type set to `application/json` to `https://www.example.com/api/dynupd`.
The data sent will be of a JSON object taking the following form:
```
{
    "username": "clientusername",
    "password": "plaintextpassword",
    "ipaddr": "desiredipaddress",
    "hosts": [ "hostname1", "hostname2" ]
}
```

Note that `hosts` may simply be an array containing a single element but will always be an array and not a scalar. This provides maxium flexability while
limiting edge cases to be handled.
One might find it odd to explicitly specify `ipaddr` as well as one could infer this from the HTTP session data.
However, this limits client flexability, one might wish to use a proxy for updating for some bizzare reason.

Once dynupd receaves the request it will perform some preliminary validation, ensuring the request is in the proper format. \
With the data somewhat validated, a message is passed in a to-be-determined format to nsddynd which first authenticates both 
the user account and permitted hosts. Once successfully authenticated, nsdynd uses `nsd-control` to update the zone if it already 
exists and reload the zones. Finally, dynupd returns a status code and message to the client.

nsddyn will always return a status as a JSON object with the following form:
```
{
    "code": "codenumber"
}
```

The following status codes may be returned:
* XXX - Success. The request was authenticated and applied.
* XXX - The request was malformed, dynupd rejected it.
* XXX - Authentication failed, nsddynd didn't agree with your provided username or password.
* XXX - Account authentication succeeded, but permitted hosts authentication failed.
* XXX - Account and permitted hosts authentication succeeded, but something failed when updating the zone. Probably not the client's fault.

### Further Design Discussion

An astute reader will notice that a number of features are missing like rate limiting and permitted client IP ranges. \
nsddyn is intended to be run on the localloop interface while a battle tested server like Apache or Nginx acts as a proxy.

No tools are provided to manage the password store because nsddyn is not intended for large scale or enterprise installations. \
Helper scripts may be provided which require shell access to the server nsddyn is running on. \
However, the author may be open to providing these facilities in the future should available time and need arise.

## Installation Instructions

TODO

## Licensing and Copyright

nsddyn is released under the terms of the GPLv3 license, a copy of the GPL is provided in the COPYING file located in the root of this repo.
