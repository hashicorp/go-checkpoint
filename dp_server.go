package checkpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/differential-privacy/go/dpagg"
	"github.com/google/differential-privacy/go/noise"
	"github.com/wcharczuk/go-chart/v2"
)

const epsilon = 2.0

// store stores differentially private data from users
type store struct {
	agentSum       *dpagg.BoundedSumInt64
	agentSumActual int64

	// interval ranges: 0-10, 11-100, 101-1000, 1001-10000, 10000+
	agentCount       map[string]*dpagg.Count
	agentCountActual map[string]int

	// TODO(Kit): Hey Lorna! I simulated a map of {config key: bool} for _every_ config flag in Consul. There's a 33%
	//  chance that we flip the bool to "true", simulating that a user is running the Consul agent with that config set.
	//  TestSimulateConfig should show you what this looks like before we encode it!
	//  Then I translate that bool into a Count with a single increment on the client - we should be able to merge them
	//  into the store on the server from there!
	configCount      map[string]*dpagg.Count
}

// Serve serves
func serveDiffPriv() error {
	opts := &dpagg.BoundedSumInt64Options{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 1,
		Lower:                    0,
		Upper:                    100000,
		Noise:                    noise.Laplace(),
	}

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

	store := &store{
		agentSum:         dpagg.NewBoundedSumInt64(opts),
		agentCount:       agentsCount,
		agentCountActual: make(map[string]int),
		//configCount: FIXME(kit),
	}

	mux := http.NewServeMux()
	mux.Handle("/submit", &submitHandler{store: store})
	mux.Handle("/agent/sum", &agentSumHandler{store: store})
	mux.Handle("/agent/count", &agentCountHandler{store: store})

	srv := &http.Server{
		// Addr:         fmt.Sprintf("%s:%d", address, port),
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mux,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting api server at '%d': '%s'", port, err)
	}

	return nil
}

type submitHandler struct {
	store *store
}

type submitBody struct {
	AgentsSum   []byte            `json:"agents_sum"`
	AgentsCount map[string][]byte `json:"agents_count"`
	AgentsActual int64            `json:"agents_actual"`
	ConfigCount map[string][]byte `json:"config_count"`
}

func (h *submitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// decode
	decoder := json.NewDecoder(r.Body)
	var body submitBody
	if err := decoder.Decode(&body); err != nil {
		panic(err) // FIXME:
	}

	// agents sum
	opts := &dpagg.BoundedSumInt64Options{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 1,
		Lower:                    0,
		Upper:                    100000,
		Noise:                    noise.Laplace(),
	}
	serverAgents := dpagg.NewBoundedSumInt64(opts)
	if err := serverAgents.GobDecode(body.AgentsSum); err != nil {
		panic(err)
	}
	h.store.agentSum.Merge(serverAgents)

	h.store.agentSumActual += body.AgentsActual

	// agents count
	for k, agentCount := range body.AgentsCount {
		agentsCountOpts := &dpagg.CountOptions{
			Epsilon:                  epsilon,
			MaxPartitionsContributed: 2,
			Noise:                    noise.Laplace(),
		}
		serverAgents2 := dpagg.NewCount(agentsCountOpts)
		if err := serverAgents2.GobDecode(agentCount); err != nil {
			panic(err)
		}
		h.store.agentCount[k].Merge(serverAgents2)
	}

	switch {
	case 0 <= body.AgentsActual && body.AgentsActual <= 10:
		h.store.agentCountActual["0-10"]++
	case 11 <= body.AgentsActual && body.AgentsActual <= 100:
		h.store.agentCountActual["11-100"]++
	case 101 <= body.AgentsActual && body.AgentsActual <= 1000:
		h.store.agentCountActual["101-1000"]++
	case 1001 <= body.AgentsActual && body.AgentsActual <= 10000:
		h.store.agentCountActual["1001-10000"]++
	case body.AgentsActual > 10000:
		h.store.agentCountActual["10000+"]++
	default:
		panic(fmt.Sprintf("value not right!! %v", body.AgentsActual))
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type agentSumHandler struct {
	store *store
}

func (h *agentSumHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sum := h.store.agentSum.Result()

	// // this can only be used once so copy it over???
	// // TODO: have an interval where this data starts over (epoch)
	// opts := &dpagg.BoundedSumInt64Options{
	// 	Epsilon:                  epsilon,
	// 	MaxPartitionsContributed: 1,
	// 	Lower:                    0,
	// 	Upper:                    100000,
	// 	Noise:                    noise.Laplace(),
	// }
	// h.store.agentSum = dpagg.NewBoundedSumInt64(opts)
	// h.store.agentSum.Add(sum)

	fmt.Println("agent sum (diff priv): ", sum)
	fmt.Println("agent sum (actual): ", h.store.agentSumActual)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{
		"agent_sum": sum,
	})
}

type agentCountHandler struct {
	store *store
}

func (h *agentCountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Bound contributions clientside w/ dpagg.PreAggSelectPartition

	// fmt.Println(h.store.agentCountActual["0-10"], "0-10 (Actual)")
	// fmt.Println(h.store.agentCount["0-10"].Result(), "0-10 (Private)")
	// fmt.Println(h.store.agentCountActual["11-100"], "11-100 (Actual)")
	// fmt.Println(h.store.agentCount["11-100"].Result(), "11-100 (Private)")
	// fmt.Println(h.store.agentCountActual["101-1000"], "101-1000 (Actual)")
	// fmt.Println(h.store.agentCount["101-1000"].Result(), "101-1000 (Private)")
	// fmt.Println(h.store.agentCountActual["1001-10000"], "1001-10000 (Actual)")
	// fmt.Println(h.store.agentCount["1001-10000"].Result(), "1001-10000 (Private)")
	// fmt.Println(h.store.agentCountActual["10000+"], "10000+ (Actual)")
	// fmt.Println(h.store.agentCount["10000+"].Result(), "10000+ (Private)")

	privateGraph := chart.BarChart{
		Title: "Agent Count Histogram (Private)",
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		Bars: []chart.Value{
			{Value: float64(h.store.agentCount["0-10"].Result()), Label: "0-10"},
			{Value: float64(h.store.agentCount["11-100"].Result()), Label: "11-100"},
			{Value: float64(h.store.agentCount["101-1000"].Result()), Label: "101-1000"},
			{Value: float64(h.store.agentCount["1001-10000"].Result()), Label: "1001-10000"},
			{Value: float64(h.store.agentCount["10000+"].Result()), Label: "10000+"},
		},
	}

	fp, _ := os.Create("agent_count_histogram_private.png")
	defer fp.Close()
	privateGraph.Render(chart.PNG, fp)

	actualGraph := chart.BarChart{
		Title: "Agent Count Histogram (Actual)",
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		Bars: []chart.Value{
			{Value: float64(h.store.agentCountActual["0-10"]), Label: "0-10"},
			{Value: float64(h.store.agentCountActual["11-100"]), Label: "11-100"},
			{Value: float64(h.store.agentCountActual["101-1000"]), Label: "101-1000"},
			{Value: float64(h.store.agentCountActual["1001-10000"]), Label: "1001-10000"},
			{Value: float64(h.store.agentCountActual["10000+"]), Label: "10000+"},
		},
	}

	fa, _ := os.Create("agent_count_histogram_actual.png")
	defer fa.Close()
	actualGraph.Render(chart.PNG, fa)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{
		// "agent_count": sum,
	})
}
