package status

import (
	"fmt"
	"strings"
)

var clear = fmt.Sprintf("%c[%dA%c[2K", ESC, 1, ESC)

func (w *Writer) clearLines() {
	_, _ = fmt.Fprint(w.Out, strings.Repeat(clear, w.lineCount))
}
