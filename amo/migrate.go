package amo

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/amolabs/amoabci/amo/types"
)

const Migration string = "ProtocolMigration"

var (
	// TODO: remove these variables at protocol v5 release
	// configs from 'cmd/amod'
	DataDirPath string = ""
)

func (app *AMOApp) migrateTo(protocolVersion uint64, changes []string, operation func() error) {
	// condition for migration
	if app.config.UpgradeProtocolHeight == types.DefaultUpgradeProtocolHeight ||
		app.state.Height != app.config.UpgradeProtocolHeight {
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
	protocolVersion := uint64(4)
	changes := []string{
		"#245: Update validator set after penalty is applied",
		"#249: Do not wait apply_count when the draft was not passed",
		"#261: Validator hibernation and new lazy validator penalty calculation",
		"#262: Add DID support",
		"#267: Explicit recipient in request tx",
		"#280: Restrict TxRequest on a parcel in a closed storage",
	}

	app.migrateTo(protocolVersion, changes, func() error {
		// stop draft immediately if not passed and waiting apply_count
		app.logger.Debug(Migration, "draft", "check")
		latestDraftIDUint := app.state.NextDraftID - uint32(1)
		draft := app.store.GetDraft(latestDraftIDUint, false)
		if draft != nil &&
			draft.OpenCount == 0 &&
			draft.CloseCount == 0 &&
			draft.ApplyCount > 0 {

			// totalTally = draft.TallyApprove + draft.TallyReject
			totalTally := new(types.Currency).Set(0)
			totalTally.Add(&draft.TallyApprove)
			totalTally.Add(&draft.TallyReject)

			// pass = totalTally * passRate
			ttf := new(big.Float).SetInt(&totalTally.Int)
			prf := new(big.Float).SetFloat64(app.config.DraftPassRate)
			pf := ttf.Mul(ttf, prf)

			pass := new(types.Currency)
			pf.Int(&pass.Int)

			if draft.TallyQuorum.GreaterThan(totalTally) ||
				pass.GreaterThan(&draft.TallyApprove) {
				app.logger.Debug(Migration, "draft", "found")
				draft.ApplyCount = int64(0)
				app.store.SetDraft(latestDraftIDUint, draft)
				app.logger.Debug(Migration, "draft", "dropped")
			}
		}

		// update lazy validator
		app.logger.Debug(Migration, "lazy validator", "check")
		// done in amo/types/config.go
		app.logger.Debug(Migration, "lazy validator", "set")

		// clean up useless stores: group_counter.db, incentive.db
		app.logger.Debug(Migration, "groupCounterDB", "check")
		dbFilePath := filepath.Join(DataDirPath, "group_counter.db")
		err := os.RemoveAll(dbFilePath)
		if err != nil {
			app.logger.Debug(Migration, "groupCounterDB(warning)", err)
		}
		app.logger.Debug(Migration, "groupCounterDB", "removed")

		app.logger.Debug(Migration, "incentiveDB", "check")
		dbFilePath = filepath.Join(DataDirPath, "incentive.db")
		err = os.RemoveAll(dbFilePath)
		if err != nil {
			app.logger.Debug(Migration, "groupCounterDB(warning)", err)
		}
		app.logger.Debug(Migration, "incentiveDB", "removed")

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
