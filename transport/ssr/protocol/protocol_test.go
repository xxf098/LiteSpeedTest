package protocol

import (
	"fmt"
	"testing"
)

func TestGetDataLength(t *testing.T) {
	b := []byte{1, 2, 3, 3, 4, 5, 6, 5, 7, 7, 7, 7, 7, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	// b := []byte{1, 2, 3, 3, 4, 5, 6}
	l := getDataLength(b)
	fmt.Println(l)
}

func TestTrapezoidRandom(t *testing.T) {
	max := 1339
	d := -0.3
	r := trapezoidRandom(max, d)
	fmt.Println(r)
}
