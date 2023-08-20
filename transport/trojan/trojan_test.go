package trojan

import (
	"fmt"
	"testing"
)

func TestHexSha224(t *testing.T) {
	h := hexSha224([]byte("123"))
	fmt.Println(h)
}
