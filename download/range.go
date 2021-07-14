package download

type Range struct {
	Url       string
	RangeFrom int64
	RangeTo   int64
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
