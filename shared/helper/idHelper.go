package helper

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)
func GetUniqueID() string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	id, _ := ulid.New(ulid.Now(), entropy)
	return id.String()
}