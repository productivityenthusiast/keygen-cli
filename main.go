package main

import (
	"github.com/productivityenthusiast/keygen-cli/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
