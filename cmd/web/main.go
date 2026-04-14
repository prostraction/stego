package main

import (
	"fmt"
	"os"
	"stego/internal/config"
	"stego/internal/web_server"
)

func main() {
	config.LoadConfig()

	port := config.GetEnv("PORT", "8080")
	frontendPath := config.GetEnv("FRONTEND_PATH", "./frontend")

	srv := server.NewStegoServer(frontendPath, ":"+port)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}
