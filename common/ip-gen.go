package common

import (
	"fmt"
	"math/rand"
)

func GenerateIP() string {
	// Generate a random number between 20 and 240
	segment2 := rand.Intn(221) + 20
	segment3 := rand.Intn(256)
	segment4 := rand.Intn(256)

	ipAddress := fmt.Sprintf("104.%d.%d.%d", segment2, segment3, segment4)
	return ipAddress
}
