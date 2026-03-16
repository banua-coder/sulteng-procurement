package main

import (
	"fmt"

	"github.com/ryanaidilp/sulteng-procurement/backend/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("API server starting on port %s\n", cfg.Port)
}
