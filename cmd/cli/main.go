package main

import (
	"fmt"
	"github.com/jame-developer/aeontrac/internal/cli"
	"os"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
