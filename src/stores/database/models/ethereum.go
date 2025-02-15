package models

import (
	"time"

	"github.com/consensys/quorum-key-manager/src/stores/entities"
	"github.com/ethereum/go-ethereum/common"
)

type ETHAccount struct {
	tableName struct{} `pg:"eth_accounts"` // nolint:unused,structcheck // reason

	Address             string `pg:",pk"`
	StoreID             string `pg:",pk"`
	KeyID               string
	PublicKey           []byte
	CompressedPublicKey []byte
	Tags                map[string]string
	Disabled            bool
	CreatedAt           time.Time `pg:"default:now()"`
	UpdatedAt           time.Time `pg:"default:now()"`
	DeletedAt           time.Time `pg:",soft_delete"`
}

func NewETHAccount(account *entities.ETHAccount) *ETHAccount {
	return &ETHAccount{
		Address:             account.Address.Hex(),
		KeyID:               account.KeyID,
		PublicKey:           account.PublicKey,
		CompressedPublicKey: account.CompressedPublicKey,
		Tags:                account.Tags,
		Disabled:            account.Metadata.Disabled,
		CreatedAt:           account.Metadata.CreatedAt,
		UpdatedAt:           account.Metadata.UpdatedAt,
		DeletedAt:           account.Metadata.DeletedAt,
	}
}

func (eth *ETHAccount) ToEntity() *entities.ETHAccount {
	return &entities.ETHAccount{
		Address:             common.HexToAddress(eth.Address),
		KeyID:               eth.KeyID,
		PublicKey:           eth.PublicKey,
		CompressedPublicKey: eth.CompressedPublicKey,
		Metadata: &entities.Metadata{
			Disabled:  eth.Disabled,
			CreatedAt: eth.CreatedAt,
			UpdatedAt: eth.UpdatedAt,
			DeletedAt: eth.DeletedAt,
		},
		Tags: eth.Tags,
	}
}
