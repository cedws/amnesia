package amnesia

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/shamir"
)

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func encryptShare(data []byte, key []byte) []byte {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		panic(err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)

	result := make([]byte, 0, len(iv)+len(ciphertext))
	result = append(result, iv...)
	result = append(result, ciphertext...)

	return result
}

func splitSecret(
	data []byte,
	parts,
	threshold int,
	compress bool,
) ([][]byte, error) {
	if compress {
		var err error
		data, err = compressData(data)
		if err != nil {
			return nil, err
		}
	}

	secretHash := sha256.Sum256(data)
	data = append(data, secretHash[:]...)

	shares, err := shamir.Split(data, parts, threshold)
	if err != nil {
		panic(err)
	}

	return shares, nil
}

func Seal(
	secret []byte,
	answers map[string]string,
	threshold int,
	opts ...Option,
) ([]byte, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	shares, err := splitSecret(secret, len(answers), threshold, options.compress)
	if err != nil {
		return nil, err
	}

	sealedData, err := sealV1(answers, shares, options.compress)
	if err != nil {
		return nil, err
	}

	return sealedData, nil
}

func sealV1(
	answers map[string]string,
	shares [][]byte,
	compressed bool,
) ([]byte, error) {
	sealFile := SealedSecret{
		Version:    "1",
		Compressed: compressed,
		Shares:     make([]Share, 0, len(answers)),
	}

	for question, answer := range answers {
		idx := len(sealFile.Shares)

		salt := salt(32)
		key := kdf([]byte(answer), salt)
		encryptedShare := encryptShare(shares[idx], key)

		sealFile.Shares = append(sealFile.Shares, Share{
			Question: question,
			Salt:     encoding.EncodeToString(salt),
			Share:    encoding.EncodeToString(encryptedShare),
		})
	}

	if len(sealFile.Shares) != len(answers) {
		panic(fmt.Errorf("expected to produce %d shares, got %d", len(answers), len(sealFile.Shares)))
	}

	return json.MarshalIndent(sealFile, "", "  ")
}
