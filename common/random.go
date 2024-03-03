package common

import "math/rand"

// RandRange returns a random number between min and max (max is not included)
func RandRange(min, max int) int {
	return min + rand.Intn(max-min)
}
