package checkpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/google/differential-privacy/go/dpagg"
	"github.com/google/differential-privacy/go/noise"
)

const address = "127.0.0.1"
const port = 8080

// DPClient stores our measurements locally
type DPClient struct {
	AgentsSum   *dpagg.BoundedSumInt64
	AgentsCount map[string]*dpagg.Count

	AgentsActual int64
	// Config     map[string]dpagg.Count `json:"config"`

	// TODO: Ideas for storage of dynamic fields
	// Data map[string]interface{}       `json:"payload"`
	// partitions []Partiton
}

// Partition stores a k/v with its PreAgg validation
// type Partition struct {
// 	key string
// 	val interface{}
// 	PreAgg dpagg.PreAggSelectPartition
// }

func (c *DPClient) Write(val int64) {
	c.AgentsSum.Add(val)

	// refactor this later maybe TODO
	switch {
	case 0 <= val && val <= 10:
		c.AgentsCount["0-10"].Increment()
	case 11 <= val && val <= 100:
		c.AgentsCount["11-100"].Increment()
	case 101 <= val && val <= 1000:
		c.AgentsCount["101-1000"].Increment()
	case 1001 <= val && val <= 10000:
		c.AgentsCount["1001-10000"].Increment()
	case val > 10000:
		c.AgentsCount["10000+"].Increment()
	default:
		panic(fmt.Sprintf("value not right!! %d", val))
	}
	c.AgentsActual = val
}

// Submit sends the DPClient's data to the aggregating server.
func (c *DPClient) Submit() {
	// TODO: figure this out
	// address := fmt.Sprintf("%s:%d", address, port)
	address := fmt.Sprintf("http://localhost:%d/submit", port)
	agents, err := c.AgentsSum.GobEncode()
	if err != nil {
		panic(err)
	}

	agentsCount := make(map[string][]byte)
	for k, v := range c.AgentsCount {
		a, err := v.GobEncode()
		if err != nil {
			panic(err)
		}
		agentsCount[k] = a
	}

	s := submitBody{
		AgentsSum:    agents,
		AgentsCount:  agentsCount,
		AgentsActual: c.AgentsActual,
	}

	body, err := json.Marshal(s)
	if err != nil {
		// FIXME
		panic(err)
	}

	req, err := http.NewRequest("POST", address, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		// FIXME
		panic(err)
	}
	defer resp.Body.Close()

	// TODO: Do something with the response? Maybe store in an errors collection on the DPClient?

	// TODO: Flush
}

func main() {
	agentsSumOpts := &dpagg.BoundedSumInt64Options{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 1,
		Lower:                    0,
		Upper:                    100000,
		Noise:                    noise.Laplace(),
	}
	agentsSum := dpagg.NewBoundedSumInt64(agentsSumOpts)

	agentsCount := make(map[string]*dpagg.Count)
	agentsCountOpts := &dpagg.CountOptions{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 2,
		Noise:                    noise.Laplace(),
	}
	agentsCount["0-10"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["11-100"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["101-1000"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["1001-10000"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["10000+"] = dpagg.NewCount(agentsCountOpts)

	client := DPClient{
		AgentsSum:   agentsSum,
		AgentsCount: agentsCount,
	}
	// pasp := dpagg.PreAggSelectPartition{}

	// weighted random number
	min := int64(0)
	max := int64(0)
	bucket := rand.Int63n(12)
	switch bucket {
	case 0, 1, 2, 3, 4:
		min = 0
		max = 10
	case 5, 6, 7, 8:
		min = 11
		max = 100
	case 9, 10:
		min = 101
		max = 1000
	case 11:
		min = 1001
		max = 10000
	default:
		min = 10001
		max = 100000
	}
	simulatedClusterSize := rand.Int63n(max-min+1) + min
	if simulatedClusterSize < 0 {
		simulatedClusterSize *= -1
	}

	client.Write(simulatedClusterSize)
	client.Submit()
	return
}

func exampleWireProtocol() string {
	s := "{" +
		"agents: [0 255 98 ... ]" +
		"}"
	return s
}
