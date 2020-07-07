package amo

import (
	"bytes"
	"fmt"
)

const Migration = "ProtocolMigration"

func (app *AMOApp) migrateTo(protocolVersion uint64, changes []string, operation func() error) {
	// condition for migration
	if app.state.Height != app.config.UpgradeProtocolHeight {
		return
	}

	app.logger.Info(Migration+" - BEGIN", "ProtocolVersion", protocolVersion)
	for i, change := range changes {
		app.logger.Info(Migration, fmt.Sprintf("change %d", i+1), change)
	}

	err := operation()
	if err != nil {
		panic(err)
	}

	app.logger.Info(Migration+" - DONE", "ProtocolVersion", protocolVersion)
}

func (app *AMOApp) MigrateToX() {
	protocolVersion := uint64(0)
	changes := []string{
		"sample changes description",
		"describe changes happening on this migration",
	}

	app.migrateTo(protocolVersion, changes, func() error { return nil })
}

func (app *AMOApp) MigrateTo4() {
	protocolVersion := uint64(0x4)
	changes := []string{
		"shorten store key 'balance:' -> 'bal:'",
		"set key-value with new prefix",
		"remove exisiting key-value with old prefix",
	}

	app.migrateTo(protocolVersion, changes, func() error {
		// ignore when merkle tree doesn't have available versions
		merkleVersion := app.store.GetMerkleVersion()
		if merkleVersion == int64(0) {
			return nil
		}

		mt := app.store.GetMerkleTree()

		imt, err := mt.GetImmutable(merkleVersion)
		if err != nil {
			return err
		}

		beforeKeyPrefix := []byte("balance:")
		afterKeyPrefix := []byte("bal:")

		imt.IterateRangeInclusive(beforeKeyPrefix, nil, true, func(key, value []byte, version int64) bool {
			if !bytes.HasPrefix(key, beforeKeyPrefix) {
				return false
			}

			beforeKey := key
			afterKey := append(afterKeyPrefix, key[len(beforeKeyPrefix):]...)

			app.logger.Debug(Migration, "store:set", fmt.Sprintf("%x", afterKey))
			mt.Set(afterKey, value)
			app.logger.Debug(Migration, "store:remove", fmt.Sprintf("%x", beforeKey))
			mt.Remove(beforeKey)

			return false
		})

		return nil
	})
}

/* sample code for migration

func (app *AMOApp) MigrateTo5() {
	protocolVersion := uint64(5)
	changes := []string{
		"shorten store key 'balance:' -> 'bal:'",
		"set key-value with new prefix",
		"remove exisiting key-value with old prefix",
	}

	app.migrateTo(protocolVersion, changes, func() error {
		// ignore when merkle tree doesn't have available versions
		_, err := app.store.GetLatestVersion()
		if err != nil {
			return nil
		}

		imt, err := app.store.GetImmutableTree(true)
		if err != nil {
			return err
		}

		beforeKeyPrefix := []byte("balance:")
		afterKeyPrefix := []byte("bal:")

		imt.IterateRangeInclusive(beforeKeyPrefix, nil, true, func(key, value []byte, version int64) bool {
			if !bytes.HasPrefix(key, beforeKeyPrefix) {
				return false
			}

			beforeKey := key
			afterKey := append(afterKeyPrefix, key[len(beforeKeyPrefix):]...)

			app.logger.Debug(Migration, "store:set", fmt.Sprintf("%x", afterKey))
			app.store.Set(afterKey, value)
			app.logger.Debug(Migration, "store:remove", fmt.Sprintf("%x", beforeKey))
			app.store.Remove(beforeKey)

			return false
		})

		return nil
	})
}

*/
