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
	"context"
	"os"

	"github.com/findy-network/findy-common-go/agency/client"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// Note: this may panic on incorrect configuration/authentication/etc.
func TryCreateInvitation(ctx context.Context, jwtToken, label string) (*agency.Invitation, error) {
	conf := client.BuildClientConnBase(
		"/path/to/findy-common-go/cert",
		"localhost",
		50051,
		[]grpc.DialOption{},
	)
	conn := client.TryAuthOpen(jwtToken, conf)

	sc := agency.NewAgentServiceClient(conn)
	id := uuid.New().String()

	return sc.CreateInvitation(
		ctx,
		&agency.InvitationBase{Label: label, ID: id},
	)
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