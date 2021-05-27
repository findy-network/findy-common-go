package rpc

import (
	"os"
	"path"
	"strings"
)

// CertFiles is helper struct to keep both needed certification files together.
type CertFiles struct {
	CertFile string
	KeyFile  string
}

// PKI is helper struct to keep need certification files for both S/C.
type PKI struct {
	ServerName string
	Server     CertFiles
	Client     CertFiles
}

func LoadPKI(tlsPath, serverName string) *PKI {
	if tlsPath == "" {
		p := os.Getenv("GOPATH")
		tlsPath = path.Join(p, "src/github.com/findy-network/findy-common-go/cert")
	}
	if serverName == "" {
		serverName = "localhost"
	}
	if strings.Contains(serverName, ":") {
		serverName = strings.Split(serverName, ":")[0]
	}
	return &PKI{
		ServerName: serverName,
		Server: CertFiles{
			CertFile: path.Join(tlsPath, "server/server.crt"),
			KeyFile:  path.Join(tlsPath, "server/server.key"),
		},
		Client: CertFiles{
			CertFile: path.Join(tlsPath, "client/client.crt"),
			KeyFile:  path.Join(tlsPath, "client/client.key"),
		},
	}
}
