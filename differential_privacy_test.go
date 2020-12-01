package checkpoint

import (
	"testing"
)

func TestTheFnCalledMain(t *testing.T) {
	runs := 10
	for i := 0; i < runs; i++ {
		main()
	}
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
