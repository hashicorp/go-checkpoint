package checkpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/differential-privacy/go/dpagg"
	"github.com/google/differential-privacy/go/noise"
	"github.com/wcharczuk/go-chart/v2"
)

const epsilon = 2.0

// store stores differentially private data from users
type store struct {
	lock sync.Mutex

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
	configCount map[string]*dpagg.Count
}

// ServeDiffPriv serves the differential privacy server
func ServeDiffPriv() error {
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

	configCount := make(map[string]*dpagg.Count)

	store := &store{
		agentSum:         dpagg.NewBoundedSumInt64(opts),
		agentCount:       agentsCount,
		agentCountActual: make(map[string]int),
		configCount:      configCount,
	}

	mux := http.NewServeMux()
	mux.Handle("/submit", &submitHandler{store: store})
	mux.Handle("/agent/sum", &agentSumHandler{store: store})
	mux.Handle("/agent/count", &agentCountHandler{store: store})
	mux.Handle("/config/count", &configCountHandler{store: store})

	srv := &http.Server{
		// Addr:         fmt.Sprintf("%s:%d", address, port),
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mux,
	}

	fmt.Printf("Starting diferential privacy server at %d\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting api server at '%d': '%s'", port, err)
	}

	return nil
}

type submitHandler struct {
	store *store
}

type submitBody struct {
	AgentsSum    []byte            `json:"agents_sum"`
	AgentsCount  map[string][]byte `json:"agents_count"`
	AgentsActual int64             `json:"agents_actual"`
	ConfigCount  map[string][]byte `json:"config_count"`
}

func (h *submitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// decode
	decoder := json.NewDecoder(r.Body)
	var body submitBody
	if err := decoder.Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{
			"err": err.Error(),
		})
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
		jsonResponse(w, http.StatusBadRequest, map[string]string{
			"err": err.Error(),
		})
	}
	h.store.lock.Lock()
	defer h.store.lock.Unlock()
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
			jsonResponse(w, http.StatusBadRequest, map[string]string{
				"err": err.Error(),
			})
		}
		h.store.agentCount[k].Merge(serverAgents2)
	}

	h.store.bucketAgentsCountActual(body.AgentsActual)

	// config count
	for k, configCount := range body.ConfigCount {
		configCountOpts := &dpagg.CountOptions{
			Epsilon:                  epsilon,
			MaxPartitionsContributed: 2,
			Noise:                    noise.Laplace(),
		}
		config := dpagg.NewCount(configCountOpts)
		if err := config.GobDecode(configCount); err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{
				"err": err.Error(),
			})
		}
		// If we don't have a count yet, initialize one
		if h.store.configCount[k] == nil {
			h.store.configCount[k] = dpagg.NewCount(configCountOpts)
		}
		h.store.configCount[k].Merge(config)
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
		"agent_sum":        sum,
		"agent_sum_actual": h.store.agentSumActual,
	})
}

type agentCountHandler struct {
	store *store
}

func (h *agentCountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Bound contributions clientside w/ dpagg.PreAggSelectPartition

	p0 := h.store.agentCount["0-10"].Result()
	p11 := h.store.agentCount["11-100"].Result()
	p101 := h.store.agentCount["101-1000"].Result()
	p1001 := h.store.agentCount["1001-10000"].Result()
	p10000 := h.store.agentCount["10000+"].Result()

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
			{Value: float64(p0), Label: fmt.Sprintf("0-10 (Count: %d)", p0)},
			{Value: float64(p11), Label: fmt.Sprintf("11-100 (Count: %d)", p11)},
			{Value: float64(p101), Label: fmt.Sprintf("101-1000 (Count: %d)", p101)},
			{Value: float64(p1001), Label: fmt.Sprintf("1001-10000 (Count: %d)", p1001)},
			{Value: float64(p10000), Label: fmt.Sprintf("10000+ (Count: %d)", p10000)},
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
			{Value: float64(h.store.agentCountActual["0-10"]), Label: fmt.Sprintf("0-10 (Count: %d)", h.store.agentCountActual["0-10"])},
			{Value: float64(h.store.agentCountActual["11-100"]), Label: fmt.Sprintf("11-100 (Count: %d)", h.store.agentCountActual["11-100"])},
			{Value: float64(h.store.agentCountActual["101-1000"]), Label: fmt.Sprintf("101-1000 (Count: %d)", h.store.agentCountActual["101-1000"])},
			{Value: float64(h.store.agentCountActual["1001-10000"]), Label: fmt.Sprintf("1001-10000 (Count: %d)", h.store.agentCountActual["1001-10000"])},
			{Value: float64(h.store.agentCountActual["10000+"]), Label: fmt.Sprintf("10000+ (Count: %d)", h.store.agentCountActual["10000+"])},
		},
	}

	fa, _ := os.Create("agent_count_histogram_actual.png")
	defer fa.Close()
	actualGraph.Render(chart.PNG, fa)

	jsonResponse(w, http.StatusOK, map[string]string{
		"msg": "see generated histogram charts",
	})
}

func jsonResponse(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("","    ")
	if err := encoder.Encode(response); err != nil {
		// at this point, we've tried to return the error. log it out for now
		// and still send the status code
		fmt.Printf("error encoding status '%d' with response '%s': %s",
			code, response, err)
	}
}

type configCountHandler struct {
	store *store
}

// TODO: Bound contributions clientside w/ dpagg.PreAggSelectPartition
func (h *configCountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Store our results
	results := make(map[string]int64)
	for k, v := range h.store.configCount {
		results[k] = v.Result()
	}
	// Sort by value
	pl := make(PairList, len(results))
	i := 0
	for k, v := range results {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	jsonResponse(w, http.StatusOK, pl)
}

type Pair struct {
	Key   string
	Value int64
}
type PairList []Pair

func (p PairList) Len() int {
	return len(p)
}

func (p PairList) Less(i, j int) bool {
	return p[i].Value < p[j].Value
}

func (p PairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
