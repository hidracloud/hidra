package main

import (
	"github.com/hidracloud/hidra/v3/cmd"
	_ "github.com/hidracloud/hidra/v3/plugins/all"
)

func main() {
	cmd.Execute()
}
