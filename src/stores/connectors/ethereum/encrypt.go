package eth

import (
	"context"

	"github.com/consensys/quorum-key-manager/src/auth/types"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

func (c Connector) Encrypt(ctx context.Context, addr ethcommon.Address, data []byte) ([]byte, error) {
	logger := c.logger.With("address", addr.Hex())

	err := c.authorizator.CheckPermission(&types.Operation{Action: types.ActionEncrypt, Resource: types.ResourceEthAccount})
	if err != nil {
		return nil, err
	}

	acc, err := c.db.Get(ctx, addr.Hex())
	if err != nil {
		return nil, err
	}

	result, err := c.store.Encrypt(ctx, acc.KeyID, data)
	if err != nil {
		return nil, err
	}

	logger.Debug("data encrypted successfully")
	return result, nil
}
