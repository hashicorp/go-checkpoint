package checkpoint

import (
  "github.com/google/differential-privacy/go/dpagg"
)

func differentialPrivacyPOC(epsilon float64) (int64, error){
	limit := 10000
	lowerBounds := 0
	upperBounds := 1
	contributions := make([][]byte, limit)

	// Serial encode
	for i := range contributions {
		clientSum := dpagg.NewBoundedSumInt64(&dpagg.BoundedSumInt64Options{
			Epsilon: epsilon,
			Lower:   int64(lowerBounds),
			Upper:   int64(upperBounds),
		})
		clientSum.Add(1)
		encode, _ := clientSum.GobEncode()
		contributions[i] = encode
	}

	// Serial decode and aggregate
	// note(kit): this would probably be faster with a chunking parallel sum, but a naive impl. with 1 goroutine per
	//  contribution was slower
	var aggSum int64 // Atomic access *only*
	for _, cont := range contributions {
		serverSum := dpagg.NewBoundedSumInt64(&dpagg.BoundedSumInt64Options{
			Epsilon: epsilon,
			Lower:   int64(lowerBounds),
			Upper:   int64(upperBounds),
		})
		serverSum.GobDecode(cont)
		aggSum = aggSum + serverSum.Result()
	}

	return aggSum, nil
}
