package errs

import "fmt"

func MaybePanic(message string, err error) {
	if err != nil {
		panic(Wrap(message, err))
	}
}

func New(message string) error {
	return Wrap(message, nil)
}

func Wrap(message string, subErr error) error {
	if subErr != nil {
		return fmt.Errorf("%s\n\t| %w", message, subErr)
	}
	return fmt.Errorf("%s\n", message)
}
