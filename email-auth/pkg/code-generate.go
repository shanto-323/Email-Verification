package pkg

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateSixDigitCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := 100000 + r.Intn(900000)
	return strconv.Itoa(code)
}
