package checkpoint

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestTheFnCalledMain(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	go ServeDiffPriv()
	runs := 10000
	groups := 20
	wg := sync.WaitGroup{}
	wg.Add(groups)
	for i := 0; i < groups; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < runs / groups; j++ {
				if err := SimulateClientDonations(); err != nil {
					fmt.Printf("Simulation failed!? :O")
				}
			}
		}()
	}
	wg.Wait()

	// Retrieve results
	if err := Get("/agent/sum"); err != nil {
		fmt.Println(err)
	}
	if err := Get("/agent/count"); err != nil {
		fmt.Println(err)
	}
	if err := Get("/config/count"); err != nil {
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

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}
