package utils

import "github.com/valyala/fastrand"

// RandomInt returns a random integer between 0 and max.
// It intentionally uses maphash to avoid thread-locking of
// math/rand which might not scale well under high concurrency.
// func RandomInt(max int) int {
// 	outUint64 := new(maphash.Hash).Sum64()
// 	out := int(outUint64)
// 	if out < 0 {
// 		out = -out
// 	}
// 	return out % max
// }

// RandomInt returns a random integer between 0 and max.
func RandomInt(max int) uint32 {
	return fastrand.Uint32n(uint32(max))
}
