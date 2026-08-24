package store

import "github.com/lacsar712/paperweb/internal/model"

type CouchSnapshotView struct {
	UnitID   string
	Couch     model.CouchReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneCouchSnapshot(s model.PlantSnapshot) CouchSnapshotView {
	out := CouchSnapshotView{
		UnitID:   s.UnitID,
		Couch:     s.Couch,
		Revision: s.Revision,
	}
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
