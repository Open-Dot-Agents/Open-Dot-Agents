package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/adapterserver"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/pkg/adapterprotocol"
)

var version = "dev"

func main() {
	if err := adapterprotocol.Serve(context.Background(), os.Stdin, os.Stdout, adapterserver.New("copilot", version)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
