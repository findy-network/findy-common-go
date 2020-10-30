# findy-grpc
General helpers for JWT and gRPC handling

#### Usage

These helper packages are made to help use gRPC and JWT together. It also helps with TLS keys.

Be noted that this version doesn't yet support Go modules. Be patient, it will..

#### Todo
- [ ] Unit tests as far as can be done (going to study: below)
- [ ] integrate `google.golang.org/grpc/test/bufconn`
- [x] add a client and server main programs for manual testing
- [x] fix Go 1.15 tls certificate format problems
- [x] Simplify cert format and generation process (see `cert/`)
- [x] check if client TLS certificate could be used as well. **We have mutual TLS authentication and 1.3 version in use** 