package config

import (
	"github.com/consensysquorum/quorum-key-manager/tests/acceptance/docker/config/hashicorp"
	"github.com/consensysquorum/quorum-key-manager/tests/acceptance/docker/config/localstack"
	"github.com/consensysquorum/quorum-key-manager/tests/acceptance/utils"
)

type Composition struct {
	Containers map[string]*Container
}

type Container struct {
	HashicorpVault  *hashicorp.Config
	LocalstackVault *localstack.Config
}

func (c *Container) Field() (interface{}, error) {
	return utils.ExtractField(c)
}