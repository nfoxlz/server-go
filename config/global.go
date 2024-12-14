// global
package config

import (
	"fmt"
	"log"

	"server/util"
)

type GlobalConfig struct {
	SettingsPath              string
	PublicPath                string
	DbBusinessExceptionPrefix string
}

var global *GlobalConfig

func init() {
	global = &GlobalConfig{}
	err := util.ReadJsonFile("global.json", global)
	if nil != err {
		log.Println(err)
		return
	}
	if "" == global.DbBusinessExceptionPrefix {
		global.DbBusinessExceptionPrefix = "pq: "
	}

	util.DbBusinessExceptionPrefix = global.DbBusinessExceptionPrefix
	PluginsPath = fmt.Sprintf("%s/plugins", global.SettingsPath)
	PublicPath = global.PublicPath
}

var PluginsPath string
var PublicPath string

func GetPath(path string) string {
	return fmt.Sprintf("%s/%s", global.SettingsPath, path)
}

func GetDbBusinessExceptionPrefix() string {
	return global.DbBusinessExceptionPrefix
}
