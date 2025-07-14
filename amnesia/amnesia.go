package amnesia

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"

	"golang.org/x/crypto/argon2"
)

var encoding = base64.StdEncoding

type Share struct {
	Question string `json:"question"`
	Salt     string `json:"salt"`
	Share    string `json:"share"`
}

type SealedSecret struct {
	Version         string  `json:"version"`
	SealedTimestamp string  `json:"sealed_timestamp"`
	Shares          []Share `json:"shares"`
	Encrypted       []byte  `json:"encrypted"`
}

func kdf(password, salt []byte) []byte {
	const (
		time    = uint32(5)
		memory  = uint32(64 * 1024) // 64MiB (unit is KiB)
		threads = uint8(4)
		keyLen  = uint32(32)
	)

	return argon2.IDKey(password, salt, time, memory, threads, keyLen)
}

func random(length int) []byte {
	salt := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}

	return salt
}

func Decode(buf []byte) (*SealedSecret, error) {
	var sf SealedSecret

	if err := json.Unmarshal(buf, &sf); err != nil {
		return nil, err
	}

	return &sf, nil
}
