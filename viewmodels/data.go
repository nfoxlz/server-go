// data
package viewmodels

import (
	"server/models"
)

type DataParameter struct {
	ViewModelBase
	Path string `json:"path"`
	Name string `json:"name"`
}

type QueryParameter struct {
	DataParameter
	Parameters map[string]any `json:"parameters"`
}

type PagingQueryParameter struct {
	QueryParameter
	CurrentPageNo   uint64 `json:"currentPageNo"`
	PageSize        uint16 `json:"pageSize"`
	SortDescription string `json:"sortDescription"`
}

type Result struct {
	ErrorNo int64
	Message string
}

type QueryResult struct {
	Data []models.SimpleData
	Result
}

type PagingQueryResult struct {
	QueryResult
	Count  uint64
	PageNo uint64
}

type ActionDataParameter struct {
	DataParameter
	ActionId []byte `json:"actionId"`
}

type SaveParameter struct {
	ActionDataParameter
	Data       []models.SimpleData `json:"data"`
	TableNames []string            `json:"tableNames"`
}

type SaveData struct {
	AddedTable            models.SimpleData `json:"addedTable"`
	DeletedTable          models.SimpleData `json:"deletedTable"`
	ModifiedTable         models.SimpleData `json:"modifiedTable"`
	ModifiedOriginalTable models.SimpleData `json:"modifiedOriginalTable"`
}

type DifferentiatedSaveParameter struct {
	ActionDataParameter
	Data map[string]SaveData `json:"data"`
}
