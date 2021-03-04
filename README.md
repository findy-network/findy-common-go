# findy-grpc


Main purpose of the package is to provide helpers for JWT and gRPC handling that sill is under construction.

#### Minimum Dependencies

The `findy-grpc` package is first client API implementation for `findy-agent` DID Agency which don't use Hyperledger Indy SDK for anything. It has minimal dependencies. When `findy-agent` won't have indy-based legacy API anymore all client tools can be run with Docker if wanted.

#### Usage


These helper packages are made to help use gRPC and JWT together. It also helps with TLS keys.

Both client and server use configuration structs to init them. The most important information is location of TLS files: `LoadPKI(tlsPath)`. `ServerCfg` also has `TestLis` which allows use of `bufconn.Listener` to make inmemory unit tests possible.

#### Todo
- [ ] github Actions for CI/CD
- [ ] fsm package, protocol QA implementation and TypeID check
- [ ] fsm package, timers
- [ ] fsm package, error handling
- [x] Unit tests as far as can be done (going to study: below)
- [x] integrate `google.golang.org/grpc/test/bufconn`
- [x] add a client and server main programs for manual testing
- [x] fix Go 1.15 tls certificate format problems
- [x] Simplify cert format and generation process (see `cert/`)
- [x] check if client TLS certificate could be used as well. **We now have mutual TLS authentication and 1.2 version in use** 

## Publishing new version

Release script will tag the current version and push the tag to remote. Release script assumes it is triggered from dev branch. It takes one parameter, the next working version. E.g. if current working version is 0.1.0, following will release version 0.1.0 and update working version to 0.2.0.

```bash
git checkout dev
./release 0.2.0
```
