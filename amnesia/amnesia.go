package amnesia

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	MinQuestions = 2
	MaxQuestions = 255
)

var encoding = base64.StdEncoding

var (
	ErrTooFewQuestions  = fmt.Errorf("too few questions, minimum is %d", MinQuestions)
	ErrTooManyQuestions = fmt.Errorf("too many questions, maximum is %d", MaxQuestions)
)

var (
	ErrTooFewAnswers  = fmt.Errorf("too few answers, minimum is %d", MinQuestions)
	ErrTooManyAnswers = fmt.Errorf("too many answers, maximum is %d", MaxQuestions)
)

type Share struct {
	ID       int    `json:"id"`
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

type Question struct {
	Question string
	Answer   string
}

type Questions map[int]Question

func NewQuestions() Questions {
	return make(Questions)
}

func (q Questions) Validate() error {
	if len(q) < MinQuestions {
		return ErrTooFewQuestions
	}
	if len(q) > MaxQuestions {
		return ErrTooManyQuestions
	}
	return nil
}

func (q Questions) Set(id int, question Question) {
	q[id] = question
}

func (q Questions) Contains(s string) bool {
	for _, question := range q {
		if question.Question == s {
			return true
		}
	}

	return false
}

type Answers map[int]string

func NewAnswers() Answers {
	return make(Answers)
}

func (a Answers) Validate() error {
	if len(a) < MinQuestions {
		return ErrTooFewAnswers
	}
	if len(a) > MaxQuestions {
		return ErrTooManyAnswers
	}
	return nil
}

func (a Answers) Set(id int, answer string) {
	a[id] = answer
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
