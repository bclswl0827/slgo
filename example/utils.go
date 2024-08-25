package main

import (
	"time"

	"math/rand"
)

func generateRandomArray(length int, min, max int32) []int32 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	arr := make([]int32, length)

	for i := 0; i < length; i++ {
		arr[i] = min + rand.Int31n(max-min+1)
	}

	return arr
}
