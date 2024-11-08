// data
package viewmodels

import (
	"server/models"
)

type DataParameter struct {
	Path string
	Name string
}

type QueryParameter struct {
	DataParameter
	Parameters map[string]any
}

type PagingQueryParameter struct {
	QueryParameter
	CurrentPageNo uint64
	PageSize      uint16
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
	ActionId []byte
}

type SaveParameter struct {
	ActionDataParameter
	Data       []models.SimpleData
	TableNames []string
}

type SaveData struct {
	AddedTable            models.SimpleData
	DeletedTable          models.SimpleData
	ModifiedTable         models.SimpleData
	ModifiedOriginalTable models.SimpleData
}

type DifferentiatedSaveParameter struct {
	ActionDataParameter
	Data map[string]SaveData
}
