#known_hosts

A simple package to manage `~/.ssh/known_hosts` verification.

##Full Example

The [example](https://github.com/emptyinterface/known_hosts/tree/master/example) folder contains a working example that runs a single shell command remotely via ssh, and verifies the host using this package.

##LICENSE
Apache 2.0

_This package was derived from [github.com/coreos/fleet/ssh](https://github.com/coreos/fleet/tree/master/ssh) and updated to import [golang.org/x/crypto/ssh](https://golang.org/x/crypto/ssh) directly instead of the locally vendored verison.  The local vendor namespacing prohibited use of the library outside of the Fleet project.  There is an open [issue](https://github.com/coreos/fleet/issues/1477) regarding this vendoring strategy. The proxy, tunnel, remote execution functionality, and external dependencies on fleet packages have been removed._

