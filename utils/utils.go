package utils

import (
	"context"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"unsafe"

	"github.com/xxf098/lite-proxy/common"
	"github.com/xxf098/lite-proxy/log"
)

func CheckLink(link string) ([]string, error) {
	r := regexp.MustCompile("(?i)^(vmess|trojan|ss|ssr)://.+")
	matches := r.FindStringSubmatch(link)
	if len(matches) < 2 {
		return nil, common.NewError("Not Suported Link")
	}
	return matches, nil
}

func getMaxProcs() int {
	if runtime.GOOS != "linux" {
		return 1
	}
	return runtime.NumCPU()
}

func GetListens(ctx context.Context, network, address string) ([]net.Listener, error) {
	maxProcs := getMaxProcs() / 2
	if maxProcs < 1 {
		maxProcs = 1
	}
	listens := make([]net.Listener, maxProcs)
	for i := 0; i < maxProcs; i++ {
		listen, err := Listen(ctx, network, address)
		if err != nil {
			return nil, err
		}
		log.D("server", i, "pid", os.Getpid(), "serving on", listen.Addr())
		listens[i] = listen
	}
	return listens, nil
}

// nolint
func B2s(b []byte) string { return *(*string)(unsafe.Pointer(&b)) } // tricks

// fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32

func U16toa(i uint16) string {
	return strconv.FormatUint(uint64(i), 10)
}
