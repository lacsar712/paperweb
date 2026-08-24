package api

import (
	"errors"

	"github.com/lacsar712/paperweb/internal/model"
)

func classifyCouchError(err error) (string, bool) {
	if errors.Is(err, model.ErrCouchLevelLow) {
		return "couch_level_low", true
	}
	return "", false
}
