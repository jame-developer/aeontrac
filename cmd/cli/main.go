package main

import (
	"fmt"
	"os"
	"github.com/jame-developer/aeontrac/internal/cli"
)


func main() {
	if err := cli.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
