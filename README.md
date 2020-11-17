# findy-grpc

Includes WebAuthn sample server. That will work as a reference implementation how to allocate `findy-agent` cloud agents from fido2 compatible web wallets.

Main purpose of the package is to provide helpers for JWT and gRPC handling that sill is under construction.

#### Usage

A current version of the WebAuthn server can be started from package root:
```shell script
$ go run .
```

These helper packages are made to help use gRPC and JWT together. It also helps with TLS keys.

Both client and server use configuration structs to init them. The most important information is location of TLS files: `LoadPKI(tlsPath)`. `ServerCfg` also has `TestLis` which allows use of `bufconn.Listener` to make inmemory unit tests possible.

#### Todo
- [x] Unit tests as far as can be done (going to study: below)
- [x] integrate `google.golang.org/grpc/test/bufconn`
- [x] add a client and server main programs for manual testing
- [x] fix Go 1.15 tls certificate format problems
- [x] Simplify cert format and generation process (see `cert/`)
- [x] check if client TLS certificate could be used as well. **We now have mutual TLS authentication and 1.3 version in use** 
