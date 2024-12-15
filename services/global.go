// global.go
package services

import (
	"fmt"
	"os"

	"server/components"
	"server/config"
	"server/util"
)

var saveConfigCache components.Cache[string, config.SaveConfig]

func init() {
	saveConfigCache.Initialize()
}

func getSaveConfig(path, name string) (config.SaveConfig, error) {
	fileName := fmt.Sprintf("%s/%s/%s.json", config.PluginsPath, path, name)
	return saveConfigCache.TryGet(fileName, func(_ *string) (config.SaveConfig, error) {
		result := config.SaveConfig{CommonTable: ""}

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			return result, err
		}

		err = util.ReadJsonFile(fileName, &result)
		if nil != err {
			util.LogError(err)
			return result, err
		}

		return result, nil
	})
}
