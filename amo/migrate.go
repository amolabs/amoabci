package amo

func (app *AMOApp) migrateTo(protocolVersion uint64, changes []string, operation func() error) {
	app.logger.Debug("Migrating to %s", protocolVersion)
	for i, change := range changes {
		app.logger.Debug("%d. %s", i+1, change)
	}

	err := operation()
	if err != nil {
		panic(err)
	}

	app.logger.Debug("Successfully migrated to %s", protocolVersion)
}

func (app *AMOApp) MigrateToX(protocolVersion uint64) {
	changes := []string{
		"sample changes description",
		"describe changes happening on this migration",
	}

	app.migrateTo(protocolVersion, changes, func() error { return nil })
}
