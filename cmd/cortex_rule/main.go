package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"ansible-cortex-rules/ansible"
)

func main() {
	if len(os.Args) < 2 {
		panic("expected at least 2 arguments")
	}

	ansibleInputFile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer ansibleInputFile.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	module, err := ansible.ModuleSetup(ctx, ansibleInputFile)
	logger := module.Logger()
	if err != nil {
		fmt.Println(module.RenderResponse())
		logger.Print(err)
		return
	}

	err = module.Run(logger)
	if err != nil {
		logger.Print(err)
		return
	}
	fmt.Println(module.RenderResponse())
}
