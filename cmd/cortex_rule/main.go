package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"ansible-cortex-rules/ansible"
)

func main() {
	logFile, err := os.OpenFile("/tmp/cortex-ansible.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	logger := log.New(logFile, "", log.LstdFlags)

	if len(os.Args) < 2 {
		panic("expected at least 2 arguments")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	module, err := ansible.ModuleSetup(ctx, f)
	if err != nil {
		fmt.Println(module.RenderResponse())
		logger.Print(err)
		return
	}

	state := module.CompareState(logger)
	if state.GroupFailed() {
		fmt.Println(module.RenderResponse())
		logger.Print(module.ResponseMessage())
		return
	}

	if state.GroupNeedToChange() || state.GroupNotFound() {
		err := module.ApplyChange(&state)
		if err != nil {
			logger.Print(module.ResponseMessage())
		}
	}
	fmt.Println(module.RenderResponse())
}
