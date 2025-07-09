package interactive

import (
	"context"
	"fmt"

	"github.com/cedws/amnesia/amnesia"
	"github.com/charmbracelet/huh"
)

type options struct {
	testQuestions bool
}

type Option func(*options)

func WithTestQuestions() Option {
	return func(o *options) {
		o.testQuestions = true
	}
}

func Seal(ctx context.Context, secret []byte, opts ...Option) ([]byte, error) {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	questions, err := promptForQuestions(ctx)
	if err != nil {
		return nil, err
	}

	if options.testQuestions {
		if err := promptForTestQuestions(ctx, questions); err != nil {
			return nil, err
		}
	}

	threshold, err := promptThreshold(ctx, len(questions))
	if err != nil {
		return nil, err
	}

	return amnesia.Seal(secret, questions, threshold)
}

func Unseal(ctx context.Context, sealed []byte, _ ...Option) ([]byte, error) {
	decoded, err := amnesia.Decode(sealed)
	if err != nil {
		return nil, err
	}

	answers := make(map[string]string)

	for _, share := range decoded.Shares {
		answer, err := promptForAnswer(ctx, share.Question)
		if err != nil {
			return nil, err
		}

		answers[share.Question] = answer
	}

	return amnesia.Unseal(sealed, answers)
}

func promptForQuestions(ctx context.Context) (map[string]string, error) {
	questions := make(map[string]string)
	cont := true

	newGroup := func(question, answer *string) *huh.Group {
		return huh.NewGroup(
			huh.NewInput().
				Title("Enter a question").
				Description("This question will be asked when unsealing the secret").
				Value(question).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("string cannot be empty")
					}
					if _, ok := questions[*question]; ok {
						return fmt.Errorf("question already set")
					}
					return nil
				}),
			huh.NewInput().
				Title("Enter an answer").
				Description("This answer will be required to unseal the secret").
				EchoMode(huh.EchoModePassword).
				Value(answer).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("string cannot be empty")
					}
					return nil
				}),
			huh.NewConfirm().
				Title("Enter another question?").
				Value(&cont).
				Validate(func(b bool) error {
					// < 1 because question hasn't been added yet
					if !b && len(questions) < 1 {
						return fmt.Errorf("at least two questions are required")
					}
					return nil
				}),
		)
	}

	for cont {
		var (
			question string
			answer   string
		)

		form := huh.NewForm(newGroup(&question, &answer))
		if err := form.RunWithContext(ctx); err != nil {
			return nil, err
		}

		questions[question] = answer

		if len(questions) >= 256 {
			return nil, fmt.Errorf("too many questions, must be fewer than 256")
		}
	}

	return questions, nil
}

func promptForTestQuestions(ctx context.Context, answers map[string]string) error {
	var fields []huh.Field

	for question, answer := range answers {
		fields = append(fields, huh.NewInput().
			Title(fmt.Sprintf("Test question: %s", question)).
			Description("Enter the answer to the test question").
			EchoMode(huh.EchoModePassword).
			Validate(func(s string) error {
				if s != answer {
					return fmt.Errorf("incorrect answer")
				}
				return nil
			}))
	}

	form := huh.NewForm(
		huh.NewGroup(
			fields...,
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		return err
	}

	return nil
}

func promptThreshold(ctx context.Context, numQuestions int) (int, error) {
	var options []int
	var threshold int

	for i := 2; i <= numQuestions; i++ {
		options = append(options, i)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Options(huh.NewOptions(options...)...).
				Title("Select threshold").
				Description("This is the number of correct answers required to unseal the secret").
				Value(&threshold),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		return 0, err
	}

	return threshold, nil
}

func promptForAnswer(ctx context.Context, question string) (string, error) {
	var answer string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(question).
				Description("If you don't know the answer, leave it blank").
				EchoMode(huh.EchoModePassword).
				Value(&answer),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		return "", err
	}

	return answer, nil
}
