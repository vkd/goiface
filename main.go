package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Print("Wrong args: goiface impl [receiver] [interface]\n")
		os.Exit(1)
	}
	err := Impl(args[1], args[2], os.Stdout)
	if err != nil {
		fmt.Printf("error on impl: %s\n", err.Error())
	}
}
