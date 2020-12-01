package checkpoint

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/google/differential-privacy/go/dpagg"
	"github.com/google/differential-privacy/go/noise"
)

const address = "127.0.0.1"
const port = 8080

// DPClient stores our measurements locally
type DPClient struct {
	Agents *dpagg.BoundedSumInt64

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
	c.Agents.Add(val)
}

// Submit sends the DPClient's data to the aggregating server.
func (c *DPClient) Submit() {
	// TODO: Bound contributions clientside w/ dpagg.PreAggSelectPartition

	// address := fmt.Sprintf("%s:%d", address, port)
	agents, err := c.Agents.GobEncode()
	if err != nil {
		panic(err)
	}

	s := struct{
		Agents []byte `json:"agents"`
	}{
		Agents: agents,
	}

	body, err := json.Marshal(s)
	if err != nil {
		// FIXME
		panic(err)
	}

	v := struct{
		Agents []byte `json:"agents"`
	}{
	}
	err = json.Unmarshal(body, &v)
	if err != nil {
		panic(err)
	}

	opts := &dpagg.BoundedSumInt64Options{
		Epsilon:                  math.Log(3),
		MaxPartitionsContributed: 100000,
		Lower:                    0,
		Upper:                    100000,
		Noise:                    noise.Laplace(),
	}
	serverAgents := dpagg.NewBoundedSumInt64(opts)
	err = serverAgents.GobDecode(v.Agents)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d\n", serverAgents.Result())
	return

	// FIXME: POST to server eventually
	// req, err := http.NewRequest("POST", address, bytes.NewBuffer(body))
	// if err != nil {
	// 	panic(err)
	// }
	// req.Header.Set("Content-Type", "application/json")
	// client := http.DefaultClient
	// resp, err := client.Do(req)
	// if err != nil {
	// 	// FIXME
	// 	panic(err)
	// }
	// defer resp.Body.Close()

	// TODO: Do something with the response? Maybe store in an errors collection on the DPClient?

	// TODO: Flush
}

func main() {
	agentsOpts := &dpagg.BoundedSumInt64Options{
		Epsilon:                  math.Log(3),
		MaxPartitionsContributed: 100000,
		Lower:                    0,
		Upper:                    100000,
		Noise:                    noise.Laplace(),
	}
	agents := dpagg.NewBoundedSumInt64(agentsOpts)
	client := DPClient{
		Agents: agents,
	}
	// pasp := dpagg.PreAggSelectPartition{}

	simulatedClusterSize := rand.Int63n(100000)
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
