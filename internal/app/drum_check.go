package app

import (
	"fmt"

	"github.com/lacsar712/paperweb/internal/model"
)

func (a *App) CheckCouchLevel(snap model.PlantSnapshot) error {
	if snap.Couch.LevelPercent < model.MinCouchLevelPercent {
		return fmt.Errorf("%w", model.ErrCouchLevelLow)
	}
	return nil
}
