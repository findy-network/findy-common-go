# findy-common-go


Main purpose of this package is to provide helpers and utility functionality for connecting to [findy-agent core](https://github.com/findy-network/findy-agent) through [findy-agent-api](https://github.com/findy-network/findy-agent-api) GRPC interface.

## Main features

* [Code](./grpc) generated from API IDL.
* Helpers for opening secure GRPC connection to core agency.
* Helpers for API protocol starters and event listeners.
* Test [TLS certificates](./cert) for local development setup.

## Example

```go
import (
	...

	"github.com/findy-network/findy-common-go/agency/client"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	didexchange "github.com/findy-network/findy-common-go/std/didexchange/invitation"
	...
)

func CreateInvitation(ctx context.Context, jwtToken, label string) (string, error) {
	conf := client.BuildClientConnBase(
		"/path/to/findy-common-go/cert",
		"localhost",
		50051,
		[]grpc.DialOption{},
	)
	conn := client.TryAuthOpen(jwtToken, conf)

	sc := agency.NewAgentServiceClient(conn)
	id = uuid.New().String()

	res, err := sc.CreateInvitation(
		ctx,
		&agency.InvitationBase{Label: label, ID: id},
	)

	if err != nil {
		return res.JSON, nil
	}

	return nil, err
}
```

## Reference implementations

* [findy-agent-cli](https://github.com/findy-network/findy-agent-cli): *GRPC client*, agent CLI tool providing most API functionality through a handy command-line-interface.

* [findy-agent](https://github.com/findy-network/findy-agent): *GRPC server* (agency internal). Implements core agency services.
* [findy-agent-vault](https://github.com/findy-network/findy-agent-vault): *GRPC client* (agency internal). Provides agency data storage service.
* [findy-agent-auth](https://github.com/findy-network/findy-agent-auth): *GRPC client* (agency internal). Provides agency authentication service

## Development

**Unit testing**: `make test`

**Linting**: Install [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) and run `make lint`.

[See more](./scripts/README.md)