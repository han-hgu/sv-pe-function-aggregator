# Sanvine API Aggregator

The `sv-api-aggregator` is a reverse proxy for an experimental API in the
policy engine. It sits in front of multiple policy engines and learn about
their existence via multicast. When queried, it turns around and queries
all policy engines that have previously announced themselves, aggregates
their responses in a single response, and returns that to the caller.

When possible, it will keep the upstream connection to the policy engine
open so multiple requests use the same socket and avoid TCP handshakes.

## Building

You need a Go development environment with both `GOROOT` and `GOPATH`
variables set. To build, run:

	go build

If you'd like to cross-compile the daemon for a different platform, such
as the PTS (freebsd/amd64) you can use the `GOOS` and `GOARCH` variables:

	GOOS=freebsd GOARCH=amd64 go build

Then scp the binary over to the PTS and voilà.

## Testing

Test coverage is quite good, try it:

	go test -v

To have detailed coverage run the following:

	go test -cover -coverprofile=c.out
	go tool cover -html=c.out

It should open a page on your browser with the detailed test coverage
against the code.

## TODO

- Add the X-Forwarded-For header
- Support /tables/$name to get and set table rows
