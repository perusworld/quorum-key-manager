package keys

import (
	"context"

	"github.com/consensys/quorum-key-manager/src/auth/types"

	"github.com/consensys/quorum-key-manager/src/stores/entities"
)

func (c Connector) Create(ctx context.Context, id string, alg *entities.Algorithm, attr *entities.Attributes) (*entities.Key, error) {
	logger := c.logger.With("id", id, "algorithm", alg.Type, "curve", alg.EllipticCurve)
	logger.Debug("creating key")

	err := c.authorizator.CheckPermission(&types.Operation{Action: types.ActionWrite, Resource: types.ResourceKey})
	if err != nil {
		return nil, err
	}

	key, err := c.store.Create(ctx, id, alg, attr)
	if err != nil {
		return nil, err
	}

	key, err = c.db.Add(ctx, key)
	if err != nil {
		// @TODO Ensure key is destroyed if we fail to insert in DB
		return nil, err
	}

	logger.Info("key created successfully")
	return key, nil
}