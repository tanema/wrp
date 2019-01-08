package main

import (
	"fmt"
	"github.com/tanema/wrp/cmd"
)

func main() {
	if err := cmd.WrpCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
