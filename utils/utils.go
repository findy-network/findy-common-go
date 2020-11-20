package utils

import (
	"flag"
	"os"
	"strings"

	"github.com/google/uuid"
)

func ParseLoggingArgs(s string) {
	args := make([]string, 1, 12)
	args[0] = os.Args[0]
	args = append(args, strings.Split(s, " ")...)
	orgArgs := os.Args
	os.Args = args
	flag.Parse()
	os.Args = orgArgs
}

func UUID() string {
	return uuid.New().String()
}
