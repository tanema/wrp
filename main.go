package main

import (
	"fmt"
	"github.com/tanema/wrp/src/config"
	"os"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	args := os.Args
	if len(args) == 1 {
		if err := cfg.FetchAllDependencies(); err != nil {
			fmt.Println(err)
			return
		}
	} else if len(args) >= 2 && args[1] == "add" {
		if len(args) < 3 {
			fmt.Println("please specify a dependency")
			return
		}
		pick := []string{}
		if len(args) > 3 {
			pick = args[3:]
		}
		if err := cfg.Add(args[2], pick); err != nil {
			fmt.Println(err)
			return
		}
	} else if len(args) >= 2 && args[1] == "rm" {
		if len(args) < 3 {
			fmt.Println("please specify a dependency")
			return
		}
		if err := cfg.Remove(args[2]); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("Unknown command", args[1], len(args))
		return
	}

	if err := cfg.Save(); err != nil {
		fmt.Println(err)
	}
}
