package main

import (
	"context"
	"fmt"
	"noctua/api"
	"noctua/kernel"
	_ "noctua/pkg"
	"noctua/types"
)

func main() {
	app := kernel.NewKernel(context.Background(), "dev")
	app.AddRuntimeHandler(func(data types.RuntimeData) {
		fmt.Println(data.EventData)
	})
	app.ProcessRuntime()

	server := api.NewServer(app)
	err := server.Run(app.Ctx)
	if err != nil {
		return
	}
}
