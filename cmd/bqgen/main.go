package main

import (
	"log"
	"os"

	"github.com/aereal/bqgen/cli"
)

func main() {
	app := cli.NewApp(nil, nil)
	if err := app.Run(os.Args); err != nil {
		log.Printf("! %+v", err)
		os.Exit(1)
	}
}
