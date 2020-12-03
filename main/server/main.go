package main

import (
	"math/rand"
	"time"

	"github.com/hashicorp/go-checkpoint"
)

// runs differential privacy server. `go run main/server/main.go`
func main() {
	rand.Seed(time.Now().UnixNano())
	checkpoint.ServeDiffPriv()
}
