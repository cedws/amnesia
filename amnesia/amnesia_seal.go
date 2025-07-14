package amnesia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"time"

	"github.com/hashicorp/vault/shamir"
)

// encryptShare encrypts a share of the DEK with AES-CTR
func encryptShare(data, key []byte) []byte {
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

// encryptData encrypts data with AES-GCM using the DEK
func encryptData(data []byte, key []byte) []byte {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	result := make([]byte, 0, len(nonce)+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result
}

func Seal(
	secret []byte,
	questions Questions,
	threshold int,
) ([]byte, error) {
	if err := questions.Validate(); err != nil {
		return nil, err
	}

	return sealV1(secret, questions, threshold)
}

func sealV1(
	secret []byte,
	questions Questions,
	threshold int,
) ([]byte, error) {
	sealedSecret := SealedSecret{
		Version:         "1",
		SealedTimestamp: time.Now().Format(time.RFC3339),
		Shares:          make([]Share, 0, len(questions)),
	}

	// DEK encryption key for secret
	dekKey := random(32)
	sealedSecret.Encrypted = encryptData(secret, dekKey)

	// Split DEK encryption key into shares
	shares, err := shamir.Split(dekKey, len(questions), threshold)
	if err != nil {
		return nil, err
	}

	for id, question := range questions {
		idx := len(sealedSecret.Shares)

		// Encryption key/salt for KEK share
		kekSalt := random(32)
		kekKey := kdf([]byte(question.Answer), kekSalt)
		encryptedShare := encryptShare(shares[idx], kekKey)

		sealedSecret.Shares = append(sealedSecret.Shares, Share{
			ID:       id,
			Question: question.Question,
			Salt:     encoding.EncodeToString(kekSalt),
			Share:    encoding.EncodeToString(encryptedShare),
		})
	}

	return json.MarshalIndent(sealedSecret, "", "  ")
}
