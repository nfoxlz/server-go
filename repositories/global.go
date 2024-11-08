// global
package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"server/components"
	"server/config"
	"server/models"
	"server/util"

	"github.com/jmoiron/sqlx"
)

var defaultConfig config.DbConfig

var configCache components.Cache[string, config.DbConfig]
var sqlConfigCache components.Cache[string, config.SqlConfig]
var sqlCache components.Cache[string, string]
var paramCache components.Cache[string, map[string]string]
var sysParamCache components.Cache[string, map[string]string]
var relatedParamCache components.Cache[string, map[string]string]
var sequenceInfoMap map[int64]models.SequenceInfo

func init() {
	InitCache()

	// defaultConfig = &config.DbConfig{}
	// err := getConfig("default", defaultConfig)
	// if nil != err {
	// 	return
	// }
	defaultConfig, _ = getConfig("default")

	var sequenceInfos []models.SequenceInfo
	err := util.ReadJsonFile(config.GetPath("sequenceInfo.json"), &sequenceInfos)
	if nil != err {
		util.LogError(err)
	}
	sequenceInfoMap = make(map[int64]models.SequenceInfo, len(sequenceInfos))
	for _, info := range sequenceInfos {
		sequenceInfoMap[info.No] = info
	}
}

func InitCache() {
	configCache.Initialize()
	sqlConfigCache.Initialize()
	sqlCache.Initialize()
	paramCache.Initialize()
	sysParamCache.Initialize()
	relatedParamCache.Initialize()
}

func toTenDays(days int) int {
	result := days / 10
	if result == 3 {
		return 2
	} else {
		return result
	}
}

func getConfig(name string) (config.DbConfig, error) {
	return configCache.TryGet(name, func(_ *string) (config.DbConfig, error) {
		result := config.DbConfig{}

		err := util.ReadJsonFile(fmt.Sprintf("database_%s.json", name), &result)
		if nil != err {
			util.LogError(err)
			return result, err
		}

		return result, nil
	})
}

func getSqlConfig(path, name string) (config.SqlConfig, error) {
	fileName := fmt.Sprintf("%s/%s/%s.json", config.PluginsPath, path, name)
	return sqlConfigCache.TryGet(fileName, func(_ *string) (config.SqlConfig, error) {
		result := config.SqlConfig{UseTransaction: false}

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			return result, nil
		}

		err = util.ReadJsonFile(fileName, &result)
		if nil != err {
			util.LogError(err)
			return result, err
		}

		return result, nil
	})
}

func getDriverName(name string) (string, error) {
	dbConfig, err := getConfig(name)
	if nil != err {
		return "", err
	}

	return dbConfig.DriverName, nil
}

func getSqlName(path, name, driverName string) string {
	return fmt.Sprintf("%s/%s/%s/%s.sql", config.PluginsPath, path, driverName, name)
}

func getSql(path, name, driverName string) (string, error) {
	return sqlCache.TryGet(getSqlName(path, name, driverName), func(key *string) (string, error) {
		buf, err := os.ReadFile(*key)
		if nil != err {
			util.LogError(path)
			util.LogError(driverName)
			util.LogError(name)
			util.LogError(err)
			return "", err
		}

		return string(buf), nil
	})
}

func getSqlByServerName(path, name, serverName string) (string, error) {
	dbConfig, err := getConfig(serverName)
	if nil != err {
		return "", err
	}

	return getSql(path, name, dbConfig.DriverName)
}

func getDefaultSql(path, name string) (string, error) {
	return getSql(path, name, defaultConfig.DriverName)
}

func isFileExist(path, name, serverName, ext string) bool {
	dbConfig, err := getConfig(serverName)
	if nil != err {
		return false
	}
	return util.IsFileExist(fmt.Sprintf("%s/%s/%s/%s.%s", config.PluginsPath, path, dbConfig.DriverName, name, ext))
}

func isSqlFileExist(path, name, serverName string) bool {
	return isFileExist(path, name, serverName, "sql")
}

func isParamFileExist(path, name, serverName string) bool {
	return isFileExist(path, name, serverName, "param")
}

func isSystemParamFileExist(path, name, serverName string) bool {
	return isFileExist(path, name, serverName, "sp")
}

func isRelatedParamFileExist(path, name, serverName string) bool {
	return isFileExist(path, name, serverName, "rp")
}

func getMapFile(key *string) (map[string]string, error) {
	if !util.IsFileExist(*key) {
		return nil, nil
	}

	buf, err := os.ReadFile(*key)
	if nil != err {
		util.LogError(err)
		return nil, err
	}

	result := make(map[string]string, 0)
	lines := util.Split(string(buf))
	for _, line := range lines {
		index := strings.Index(line, "=")
		if index >= 0 {
			result[line[0:index]] = line[index+1:]
		} else {
			result[line] = line
		}
	}

	return result, nil
}

func getParamSettings(path, name, driverName string) (map[string]string, error) {
	settings, err := paramCache.TryGet(fmt.Sprintf("%s/%s/%s/%s.param", config.PluginsPath, path, driverName, name), getMapFile)
	if nil != err {
		return nil, err
	}

	if nil == settings {
		ext := filepath.Ext(name)
		return paramCache.TryGet(fmt.Sprintf("%s/%s/%s/%s.param", config.PluginsPath, path, driverName, name[:len(name)-len(ext)]), getMapFile)
	}

	return settings, nil
}

func getParamSettingsByServerName(path, name, serverName string) (map[string]string, error) {
	dbConfig, err := getConfig(serverName)
	if nil != err {
		return nil, err
	}

	return getParamSettings(path, name, dbConfig.DriverName)
}

func getDefaultParamSettings(path, name string) (map[string]string, error) {
	return getParamSettings(path, name, defaultConfig.DriverName)
}

func getSystemParamSettings(path, name, driverName string) (map[string]string, error) {
	return sysParamCache.TryGet(fmt.Sprintf("%s/%s/%s/%s.sp", config.PluginsPath, path, driverName, name), getMapFile)
}

func getSystemParamSettingsByServerName(path, name, serverName string) (map[string]string, error) {
	dbConfig, err := getConfig(serverName)
	if nil != err {
		return nil, err
	}

	return getSystemParamSettings(path, name, dbConfig.DriverName)
}

func getSystemDefaultParamSettings(path, name string) (map[string]string, error) {
	return getSystemParamSettings(path, name, defaultConfig.DriverName)
}

func getRelatedParamSettings(path, name, serverName string) (map[string]string, error) {
	dbConfig, err := getConfig(serverName)
	if nil != err {
		return nil, err
	}

	return relatedParamCache.TryGet(fmt.Sprintf("%s/%s/%s/%s.rp", config.PluginsPath, path, dbConfig.DriverName, name), getMapFile)
}

func createDbByName(name string) (*sqlx.DB, error) {
	dbConfig, err := getConfig(name)
	if nil != err {
		return nil, err
	}
	return sqlx.Open(dbConfig.DriverName, dbConfig.DataSourceName)
}

func StructScan(rows *sqlx.Rows, columns []string, dest any) error {
	destVal := reflect.ValueOf(dest)
	for destVal.Kind() != reflect.Struct {
		destVal = destVal.Elem()
	}

	row, err := rows.SliceScan()
	if nil != err {
		util.LogError(err)
		return err
	}

	destType := reflect.TypeOf(dest)
	for destType.Kind() != reflect.Struct {
		destType = destType.Elem()
	}

	var field reflect.StructField
	count := destVal.NumField()
	for i := 0; i < count; i++ {
		field = destType.Field(i)
		tag := field.Tag.Get("db")
		if "" == tag {
			tag := ""
			for _, c := range field.Name {
				if unicode.IsUpper(c) {
					tag += "_" + string(unicode.ToLower(c))
				} else {
					tag += string(c)
				}
			}

			if '_' == tag[0] {
				tag = tag[1:]
			}
		}

		destVal.Kind()

		for index, name := range columns {
			if tag == name && nil != row[index] {
				destVal.Field(i).Set(util.ConvertType(reflect.ValueOf(row[index]), destVal.Field(i).Kind()))
			}
		}
	}

	return nil
}

func tablesScan(rows *sqlx.Rows, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error {
	var index, rowIndex int64
	index = 0
	for true {
		var err error
		if nil != beforHandler {
			err = beforHandler(index, rows)
			if nil != err {
				util.LogError(err)
				return err
			}
		}

		rowIndex = 0

		for rows.Next() {
			err = handler(index, rowIndex, rows)
			if nil != err {
				util.LogError(err)
				return err
			}

			rowIndex++
		}

		if nil != afterHandler {
			err = afterHandler(index, rows)
			if nil != err {
				util.LogError(err)
				return err
			}
		}

		if !rows.NextResultSet() {
			break
		}
		index++
	}

	return nil
}

func amend(data []any) []any {
	for index, item := range data {
		if val, ok := item.([]byte); ok {

			str := strings.Replace(strings.Trim(string(val), " "), ",", "", -1)

			if 1 < len(str) {
				if "$" == string(str[0]) {
					str = str[1:]
				} else if "-$" == str[:2] {
					str = "-" + str[2:]
				}
			}

			num, err := strconv.ParseFloat(str, 64)
			if nil != err {
				continue
			}

			data[index] = num
		}
	}

	return data
}
