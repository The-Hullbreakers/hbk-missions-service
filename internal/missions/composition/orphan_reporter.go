package composition

import (
	"context"
	"log"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// orphanLogReporter surfaces adjudication locks held past the reconcile
// threshold through the service log; operators reconcile them manually.
type orphanLogReporter struct{}

// Report logs each orphaned lock once per sweep pass.
func (orphanLogReporter) Report(_ context.Context, orphans []application.LockedSeason) error {
	for _, orphan := range orphans {
		log.Printf("adjudication lock orphaned: season=%s attempt=%s cutoff=%s — operator reconcile required",
			orphan.SeasonID, orphan.AttemptID, orphan.Cutoff.Format("2006-01-02T15:04:05Z07:00"))
	}
	return nil
}
