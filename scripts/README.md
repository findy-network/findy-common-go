# Tools

## Maintenance

### API updates and code generation

1. Install [prerequisities](https://grpc.io/docs/languages/go/quickstart/#prerequisites)

1. Clone and checkout desired version of [findy-agent-api](https://github.com/findy-network/findy-agent-api) to one level up on folder hierarchy.

1. Run `make protoc`

### Publishing new version

Release script will tag the current version and push the tag to remote. Release script assumes it is triggered from dev branch. It takes one parameter, the next working version. E.g. if current working version is 0.1.0, following will release version 0.1.0 and update working version to 0.2.0.

```bash
git checkout dev
./scripts/release 0.2.0
```
