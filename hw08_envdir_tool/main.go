package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: go-envdir /path/to/env/dir command arg1 arg2")
		log.Fatalf("too few arguments")
	}

	env, err := ReadDir(os.Args[1])
	if err != nil {
		log.Fatalf("read environements error - %v", err.Error())
	}

	resultCode := RunCmd(os.Args[2:], env)
	os.Exit(resultCode)
}
