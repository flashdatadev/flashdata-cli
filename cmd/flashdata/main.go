package main

import (
	"context"
	"github.com/flashdatadev/flashdata-cli/internal/app"
	"os"
	"os/signal"
)

var version = "1.1.2"
var commit = "local"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	a := app.New(app.IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}, version, commit)
	os.Exit(a.Execute(ctx, os.Args[1:]))
}
