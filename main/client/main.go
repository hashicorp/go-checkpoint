package main

import (
	"fmt"

	"github.com/hashicorp/go-checkpoint"
)

// simulates clients making donations to the server. `go run main/client/main.go`
func main() {
	runs := 10000
	fmt.Printf("Simulating %d clients donating differentially privatized data\n",
		runs)
	for i := 0; i < runs; i++ {
		checkpoint.SimulateClientDonations()
	}
	fmt.Println("Finished")
}
