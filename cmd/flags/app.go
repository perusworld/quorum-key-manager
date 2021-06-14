package flags

import (
	app "github.com/consensysquorum/quorum-key-manager/src"
	"github.com/spf13/viper"
)

func NewAppConfig(vipr *viper.Viper) *app.Config {
	return &app.Config{
		Logger:    newLoggerConfig(vipr),
		HTTP:      newHTTPConfig(vipr),
		Manifests: newManifestsConfig(vipr),
	}
}