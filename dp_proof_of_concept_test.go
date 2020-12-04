package checkpoint

import (
	"sync"
	"testing"
)

func TestDifferentialPrivacy(t *testing.T) {
	runs := 10
	epsilon := 8.0
	results := make([]int64, runs)
	wg := &sync.WaitGroup{}
	wg.Add(runs)
	for i := 0; i < runs; i++ {
		go func(i int) {
			defer wg.Done()
			res, err := differentialPrivacyPOC(epsilon)
			if err != nil {
				t.Fatalf("%v", err)
			}
			results[i] = res
		}(i)
	}
	wg.Wait()

	t.Logf("%+v", results)
	// Calculate min, max, delta
}
