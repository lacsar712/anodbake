package app

import (
	"fmt"

	"github.com/lacsar712/anodbake/internal/model"
)

func (a *App) CheckAnodepitLevel(snap model.PlantSnapshot) error {
	if snap.Anodepit.LevelPercent < model.MinAnodepitLevelPercent {
		return fmt.Errorf("%w", model.ErrAnodepitLevelLow)
	}
	return nil
}
