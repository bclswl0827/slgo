package main

import "math"

const (
	sineFrequency = 1.0
	amplitude     = 1000
)

// generateSineWave returns one second of a 1 Hz sine wave.
func generateSineWave(sampleRate int) []int32 {
	samples := make([]int32, sampleRate)
	for i := range samples {
		phase := 2 * math.Pi * sineFrequency * float64(i) / float64(sampleRate)
		samples[i] = int32(math.Round(float64(amplitude) * math.Sin(phase)))
	}
	return samples
}
