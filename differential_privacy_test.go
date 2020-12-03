package checkpoint

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestTheFnCalledMain(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	go ServeDiffPriv()
	runs := 10000
	for i := 0; i < runs; i++ {
		SimulateClientDonations()
	}

	// Retrieve results
	if err := Get("/agent/sum"); err != nil {
		fmt.Println(err)
	}
	if err := Get("/agent/count"); err != nil {
		fmt.Println(err)
	}
}

func TestSimulateConfig(t *testing.T) {
	fmt.Printf("%v\n", simulateConfig())
}

func Get(path string) error {
	u := fmt.Sprintf("http://localhost:%d%s", port, path)
	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("error requesting '%s': %s", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error requesting '%s' - returned status code %d ",
			u, resp.StatusCode)
	}

	return nil
}

// func TestDifferentialPrivacy(t *testing.T) {
// 	runs := 10
// 	results := make([]string, runs)
// 	wg := &sync.WaitGroup{}
// 	wg.Add(runs)
// 	for i := 0; i < runs; i++ {
// 		go func(i int) {
// 			defer wg.Done()
// 			res := main()
// 			results[i] = res
// 		}(i)
// 	}
// 	wg.Wait()
//
// 	t.Logf("%+v", results)
// }
