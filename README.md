# findy-common-go


Main purpose of this package is to provide helpers and utility functionality for connecting to [findy-agent core](https://github.com/findy-network/findy-agent) through [findy-agent-api](https://github.com/findy-network/findy-agent-api) GRPC interface.

## Features

## Example

## Reference implementations

* [findy-agent-cli](https://github.com/findy-network/findy-agent-cli): *GRPC client*, agent CLI tool providing most API functionality through a handy command-line-interface.

* [findy-agent](https://github.com/findy-network/findy-agent): *GRPC server* (agency internal). Implements core agency services.
* [findy-agent-vault](https://github.com/findy-network/findy-agent-vault): *GRPC client* (agency internal). Provides agency data storage service.
* [findy-agent-auth](https://github.com/findy-network/findy-agent-auth): *GRPC client* (agency internal). Provides agency authentication service

## Development

**Unit testing**: `make test`

**Linting**: Install [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) and run `make lint`.

[See more](./scripts/README.md)