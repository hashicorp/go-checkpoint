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

	go serveDiffPriv()
	runs := 10000
	for i := 0; i < runs; i++ {
		main()
	}

	// Retrieve results
	Get("/agent/sum")
	Get("/agent/count")
}

func Get(path string) {
	u := fmt.Sprintf("http://localhost:%d%s", port, path)
	resp, err := http.Get(u)
	if err != nil {
		panic(err) // FIXME
	}
	defer resp.Body.Close()
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
