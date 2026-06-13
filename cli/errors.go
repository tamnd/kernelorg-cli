package cli

import (
	"errors"

	"github.com/tamnd/kernelorg-cli/kernelorg"
)

func isNotFound(err error) bool {
	return errors.Is(err, kernelorg.ErrNotFound)
}
