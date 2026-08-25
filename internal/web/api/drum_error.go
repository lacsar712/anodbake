package api

import (
	"errors"

	"github.com/lacsar712/anodbake/internal/model"
)

func classifyAnodepitError(err error) (string, bool) {
	if errors.Is(err, model.ErrAnodepitLevelLow) {
		return "anodepit_level_low", true
	}
	return "", false
}
