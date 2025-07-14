package amnesia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = []byte("zed > vim")

func TestQuestions(t *testing.T) {
	t.Run("Minimum", func(t *testing.T) {
		q := NewQuestions()
		q.Set(0, Question{
			Question: "What's your favourite animal?",
			Answer:   "cat",
		})
		err := q.Validate()
		assert.ErrorIs(t, err, ErrTooFewQuestions)
	})

	t.Run("Maximum", func(t *testing.T) {
		q := NewQuestions()

		for i := range MaxQuestions * 2 {
			q.Set(i, Question{
				Question: "What's your favourite animal?",
				Answer:   "cat",
			})
		}

		err := q.Validate()
		assert.ErrorIs(t, err, ErrTooManyQuestions)
	})

	t.Run("NoError", func(t *testing.T) {
		q := NewQuestions()

		for i := range 8 {
			q.Set(i, Question{
				Question: "What's your favourite animal?",
				Answer:   "cat",
			})
		}

		err := q.Validate()
		assert.NoError(t, err)
	})
}

func TestSeal(t *testing.T) {
	q := NewQuestions()
	q.Set(0, Question{
		Question: "What's your favourite animal?",
		Answer:   "cat",
	})
	q.Set(1, Question{
		Question: "What's your favourite food?",
		Answer:   "pizza",
	})

	sealed, err := Seal(testData, q, 2)
	assert.NoError(t, err)
	assert.NotEmpty(t, sealed)
}

func TestUnseal(t *testing.T) {
	q := NewQuestions()
	q.Set(0, Question{
		Question: "What's your favourite animal?",
		Answer:   "cat",
	})
	q.Set(1, Question{
		Question: "What's your favourite food?",
		Answer:   "pizza",
	})

	sealed, err := Seal(testData, q, 2)
	assert.NoError(t, err)
	assert.NotEmpty(t, sealed)

	a := NewAnswers()
	a.Set(0, "cat")
	a.Set(1, "pizza")

	unsealed, err := Unseal(sealed, a)
	assert.NoError(t, err)
	assert.Equal(t, testData, unsealed)
}
