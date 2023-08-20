package vmess

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

func TestNewID(t *testing.T) {
	uid, err := uuid.FromString("803277e5-64ef-48f9-8fd5-de82481ee789")
	if err != nil {
		t.Error(err)
	}
	newid := newID(&uid)
	fmt.Println(newid.CmdKey)
	ids := newAlterIDs(newid, 1)
	for _, id := range ids {
		fmt.Println(id.UUID.String())
	}
}

func hashTimestamp1(t uint64) []byte {
	md5hash := md5.New()
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, t)
	md5hash.Write(ts)
	md5hash.Write(ts)
	md5hash.Write(ts)
	md5hash.Write(ts)
	return md5hash.Sum(nil)
}

func TestHashTimestamp(t *testing.T) {
	buf := hashTimestamp1(1633248908)
	fmt.Println(buf)
}

func TestCipher(t *testing.T) {
	uid, err := uuid.FromString("db538344-45db-4713-a317-2f7fd1e4c29d")
	if err != nil {
		t.Error(err)
	}
	id := newID(&uid)
	block, err := aes.NewCipher(id.CmdKey)
	if err != nil {
		t.Error(err)
	}

	// timestamp := time.Now()
	stream := cipher.NewCFBEncrypter(block, hashTimestamp1(1633248908))
	// buf := []byte{1, 1, 1, 1}
	// stream.XORKeyStream(buf, buf)
	// fmt.Println(buf)
	buf := []byte("github.com/xxf098")
	stream.XORKeyStream(buf, buf)
	fmt.Println(buf)
}

func TestKdf(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	// kdfSaltConstAuthIDEncryptionKey
	result := kdf(key, kdfSaltConstAuthIDEncryptionKey)
	// [93 153 120 238 109 19 174 100 222 74 202 189 241 233 7 82 23 89 126 149 192 95 66 176 111 202 233 12 181 188 228 205]
	fmt.Println(result)
}

func TestSha256(t *testing.T) {
	buf := sha256.Sum256([]byte("rust"))
	fmt.Println(buf)
}

func TestHmac(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	h := hmac.New(sha256.New, key)
	h.Write([]byte("rust"))
	buf := h.Sum(nil)
	fmt.Println(buf)
}

func TestAes(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	aesBlock, _ := aes.NewCipher(key)
	var result [16]byte
	buf := []byte("rustrustrustrust")
	aesBlock.Encrypt(result[:], buf)
	fmt.Println(result)
}

func TestCreateAuthID(t *testing.T) {
	ti := time.Now()
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	tu := ti.Unix()
	fmt.Println(tu)
	id := createAuthID(key, tu)
	fmt.Println(id)
}

func TestGCM(t *testing.T) {
	payloadHeaderAEADKey := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	payloadHeaderAEADNonce := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2}
	generatedAuthID := []byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8}

	data := []byte("rust")
	aeadPayloadLengthSerializedByte := make([]byte, 2)
	binary.BigEndian.PutUint16(aeadPayloadLengthSerializedByte, uint16(len(data)))

	payloadHeaderAEADAESBlock, _ := aes.NewCipher(payloadHeaderAEADKey)
	payloadHeaderAEAD, _ := cipher.NewGCM(payloadHeaderAEADAESBlock)
	payloadHeaderAEADEncrypted := payloadHeaderAEAD.Seal(nil, payloadHeaderAEADNonce, data, generatedAuthID[:])
	fmt.Println(payloadHeaderAEADEncrypted)
}

func TestSealVMessAEADHeader(t *testing.T) {
	data := []byte("rust")
	key := [16]byte{111, 135, 225, 177, 159, 141, 162, 93, 247, 21, 214, 85, 174, 154, 23, 100}
	buf := sealVMessAEADHeader(key, data, time.Now())
	fmt.Println(buf)
}

func TestUrl(t *testing.T) {
	urlStr := "ws://zhk3.greenyun.tk:11110/onefall"
	u, err := url.Parse(urlStr)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(u.Host)
}
