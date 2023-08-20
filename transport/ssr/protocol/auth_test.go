package protocol

import (
	"crypto/rc4"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"testing"

	"github.com/xxf098/lite-proxy/transport/ssr/tools"
)

func TestAuthAes128Sha1Encode(t *testing.T) {
	// base := &Base{
	// 	Key:      []byte{253, 53, 153, 216, 16, 234, 129, 68, 247, 183, 217, 78, 176, 51, 247, 223, 201, 115, 166, 180, 163, 135, 75, 165, 113, 87, 160, 116, 106, 70, 47, 122},
	// 	Overhead: 14,
	// 	Param:    "29220:k5A4Ni",
	// }
	// auth := newAuthAES128SHA1(base)

	// data := []byte{3, 19, 99, 108, 105, 101, 110, 116, 115, 51, 46, 103, 111, 111, 103, 108, 101, 46, 99, 111, 109, 0, 80}
	// buf := tools.BufPool.Get().(*bytes.Buffer)
	// auth.Encode(buf, data)
	// fmt.Println(buf.Bytes())
}

func TestAuthAes128Sha1Decode(t *testing.T) {
}

func TestCrc32(t *testing.T) {
	buf := []byte{3, 19, 99, 108, 105, 101, 110, 116, 115, 51, 46, 103, 111, 111, 103, 108, 101, 46, 99, 111, 109, 0, 80}
	h := 0xffffffff - crc32.ChecksumIEEE(buf)
	fmt.Println(h)
}

func TestAdler32(t *testing.T) {
	buf := []byte{3, 19, 99, 108, 105, 101, 110, 116, 115, 51, 46, 103, 111, 111, 103, 108, 101, 46, 99, 111, 109, 0, 80}
	h := adler32.Checksum(buf)
	fmt.Println(h)
}

func TestGetRandLength(t *testing.T) {
	lastHash := []byte{36, 253, 199, 121, 90, 151, 29, 181, 227, 115, 216, 51, 91, 188, 55, 64}
	random := tools.XorShift128Plus{}
	a := authChainA{}
	length := a.getRandLength(23, lastHash, &random)
	fmt.Println(length)
}

func TestRc4(t *testing.T) {
	key := []byte{107, 41, 179, 39, 181, 176, 111, 92, 103, 186, 17, 249, 183, 131, 181, 124}
	c, _ := rc4.NewCipher(key)
	src := []byte{3, 19, 99, 108, 105, 101, 110, 116, 115, 51, 46, 103, 111, 111, 103, 108, 101, 46, 99, 111, 109, 0, 80}
	// dst := []byte{95, 186, 186, 30, 52, 206, 58, 189, 157, 89, 32, 71, 153, 164, 56, 109, 229, 110, 149, 218, 97, 172, 179}
	c.XORKeyStream(src, src)
	fmt.Println(src)
}

func TestAuthAES128(t *testing.T) {
	item := []byte("Q2dECE")
	usekey := tools.SHA1Sum(item)
	fmt.Printf("%x", usekey)
}
