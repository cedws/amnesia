package amnesia

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/hashicorp/vault/shamir"
)

func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decryptShare(data []byte, key []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

func joinSecret(shares [][]byte, compressed bool) ([]byte, error) {
	joined, err := shamir.Combine(shares)
	if err != nil {
		return nil, err
	}

	if len(joined) < sha256.Size {
		panic("decrypted share too short")
	}

	data := joined[:len(joined)-sha256.Size]
	hash := joined[len(joined)-sha256.Size:]

	actual := sha256.Sum256(data)
	if !bytes.Equal(actual[:], hash) {
		return nil, fmt.Errorf("failed to decrypt secret (wrong answers or not enough correct?)")
	}

	if compressed {
		var err error
		data, err = decompressData(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func Unseal(input []byte, answers map[string]string) ([]byte, error) {
	sealed, err := Decode(input)
	if err != nil {
		return nil, err
	}

	switch sealed.Version {
	case "1":
		return unsealV1(sealed, answers)
	default:
		return nil, fmt.Errorf("unknown seal version: %s", sealed.Version)
	}
}

func unsealV1(sealFile *SealedSecret, answers map[string]string) ([]byte, error) {
	var shares [][]byte

	for _, encryptedShare := range sealFile.Shares {
		answer, ok := answers[encryptedShare.Question]
		if !ok {
			// Missing answer, skip decrypting this share
			continue
		}
		if answer == "" {
			// Blank answer, user doesn't know it so skip decrypting this share
			continue
		}

		salt, err := encoding.DecodeString(encryptedShare.Salt)
		if err != nil {
			return nil, err
		}

		ciphertext, err := encoding.DecodeString(encryptedShare.Share)
		if err != nil {
			return nil, err
		}

		key := kdf([]byte(answer), salt)
		decryptedShare, err := decryptShare(ciphertext, key)
		if err != nil {
			return nil, err
		}

		shares = append(shares, decryptedShare)
	}

	buf, err := joinSecret(shares, sealFile.Compressed)
	if err != nil {
		return nil, fmt.Errorf("error joining shares: %w", err)
	}

	return buf, nil
}
