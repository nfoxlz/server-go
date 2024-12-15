// database
package config

type DbConfig struct {
	DriverName     string
	DataSourceName string
}

type SqlConfig struct {
	UseTransaction bool `json:"useTransaction"`
}

type SaveConfig struct {
	CommonTable string `json:"commonTable"`
}
