package download

import "fmt"

type Range struct {
	Url       string
	RangeFrom int64
	RangeTo   int64
}

func (rng Range) toHeader(contentLength int64) string {
	var ranges string
	if rng.RangeTo != int64(contentLength) {
		ranges = fmt.Sprintf("bytes=%d-%d", rng.RangeFrom, rng.RangeTo)
	} else {
		ranges = fmt.Sprintf("bytes=%d-", rng.RangeFrom) //get all
	}
	return ranges
}

func calcRange(part int64, len int64, url string) []Range {
	ret := []Range{}
	for j := int64(0); j < part; j++ {
		from := (len / part) * j
		var to int64
		if j < part-1 {
			to = (len/part)*(j+1) - 1
		} else {
			to = len
		}

		ret = append(ret, Range{Url: url, RangeFrom: from, RangeTo: to})
	}
	return ret
}
