package status

import (
	"fmt"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	w := New()
	w.Start()
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "=====> %d\n", i)
		time.Sleep(300 * time.Millisecond)
	}
	defer w.Stop()
}
