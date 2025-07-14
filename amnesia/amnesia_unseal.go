package amnesia

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/hashicorp/vault/shamir"
)

// decryptShare decrypts a share of the DEK with AES-CTR
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

// decryptData decrypts the data with AES-GCM using the DEK
func decryptData(data []byte, key []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < aesgcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := data[:aesgcm.NonceSize()]
	ciphertext := data[aesgcm.NonceSize():]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func Unseal(input []byte, answers Answers) ([]byte, error) {
	if err := answers.Validate(); err != nil {
		return nil, err
	}

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

func unsealV1(sealedSecret *SealedSecret, answers Answers) ([]byte, error) {
	var shares [][]byte

	for _, share := range sealedSecret.Shares {
		answer, ok := answers[share.ID]
		if !ok {
			// Missing answer, skip decrypting this share
			continue
		}
		if answer == "" {
			// Blank answer, user doesn't know it so skip decrypting this share
			continue
		}

		salt, err := encoding.DecodeString(share.Salt)
		if err != nil {
			return nil, err
		}

		ciphertext, err := encoding.DecodeString(share.Share)
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

	dekKey, err := shamir.Combine(shares)
	if err != nil {
		return nil, fmt.Errorf("error joining shares: %w", err)
	}

	secret, err := decryptData(sealedSecret.Encrypted, dekKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data (incorrect or too few answers?)")
	}

	return secret, nil
}
