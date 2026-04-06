package migration

import "github.com/rawbytedev/aether/index"

type Migration struct {
	migrationEv chan index.MigrationEvent
}

func NewMigrationService(migrationEvent <-chan index.MigrationEvent) {
	<-migrationEvent
}
