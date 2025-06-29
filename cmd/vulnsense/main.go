package main

import (
	"fmt"
	"log"

	"github.com/0xmanhnv/vulnsense/internal/config"
)

func main() {
	cfg, err := config.LoadConfig("configs/app.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println(cfg)
}
