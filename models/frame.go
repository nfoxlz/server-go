// system
package models

type Menu struct {
	MenuNo          int64  `db:"menu_no"`
	ParentMenuNo    int64  `db:"parent_menu_no"`
	Sn              int64  `db:"sn"`
	DisplayName     string `db:"display_name"`
	ToolTip         string `db:"tool_tip"`
	PluginSetting   string `db:"plugin_setting"`
	PluginParameter string `db:"plugin_parameter"`
	Authorition     int64  `db:"authorition"`
}

type EnumInfo struct {
	Name        string `db:"name"`
	Value       int16  `db:"value"`
	DisplayName string `db:"display_name"`
	Sn          int64  `db:"sn"`
}
