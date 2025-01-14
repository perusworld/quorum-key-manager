package eth

import (
	"context"

	"github.com/consensys/quorum-key-manager/src/auth/types"

	"github.com/consensys/quorum-key-manager/pkg/errors"
	"github.com/consensys/quorum-key-manager/src/stores/database"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func (c Connector) Restore(ctx context.Context, addr ethcommon.Address) error {
	logger := c.logger.With("address", addr.Hex())
	logger.Debug("restoring ethereum account")

	err := c.authorizator.CheckPermission(&types.Operation{Action: types.ActionDelete, Resource: types.ResourceEthAccount})
	if err != nil {
		return err
	}

	_, err = c.Get(ctx, addr)
	if err == nil {
		return nil
	}

	acc, err := c.db.GetDeleted(ctx, addr.Hex())
	if err != nil {
		return err
	}

	err = c.db.RunInTransaction(ctx, func(dbtx database.ETHAccounts) error {
		err = dbtx.Restore(ctx, addr.Hex())
		if err != nil {
			return err
		}

		err = c.store.Restore(ctx, acc.KeyID)
		if err != nil && !errors.IsNotSupportedError(err) { // If the underlying store does not support restoring, we only restore in DB
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.Info("ethereum account restored successfully")
	return nil
}
