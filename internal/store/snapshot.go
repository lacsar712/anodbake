package store

import "github.com/lacsar712/anodbake/internal/model"

type AnodepitSnapshotView struct {
	UnitID   string
	Anodepit     model.AnodepitReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneAnodepitSnapshot(s model.PlantSnapshot) AnodepitSnapshotView {
	out := AnodepitSnapshotView{
		UnitID:   s.UnitID,
		Anodepit:     s.Anodepit,
		Revision: s.Revision,
	}
	out.Alarms = make([]model.AlarmEvent, len(s.Alarms))
	copy(out.Alarms, s.Alarms)
	return out
}
