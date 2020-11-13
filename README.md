# findy-grpc

General helpers for JWT and gRPC handling that sill is heavily under construction.

#### Usage

These helper packages are made to help use gRPC and JWT together. It also helps with TLS keys.

Both client and server use configuration structs to init them. The most important information is location of TLS files: `LoadPKI(tlsPath)`. `ServerCfg` also has `TestLis` which allows use of `bufconn.Listener` to make inmemory unit tests possible.

#### Todo
- [x] Unit tests as far as can be done (going to study: below)
- [x] integrate `google.golang.org/grpc/test/bufconn`
- [x] add a client and server main programs for manual testing
- [x] fix Go 1.15 tls certificate format problems
- [x] Simplify cert format and generation process (see `cert/`)
- [x] check if client TLS certificate could be used as well. **We now have mutual TLS authentication and 1.3 version in use** 