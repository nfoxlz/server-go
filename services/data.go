// data
package services

import (
	"errors"
	"fmt"
	"server/components"
	"server/models"
	"server/util"
	"server/viewmodels"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type DataService struct {
	BusinessService
}

func (s DataService) Query(path, name string, parameters map[string]any) (map[string]models.SimpleData, error) {
	s.repository.SetComponent(s.BusinessComponent)
	return s.repository.QueryTables(path, name, parameters)
}

func (s DataService) QueryByParameter(parameter viewmodels.QueryParameter) (map[string]models.SimpleData, error) {
	return s.Query(parameter.Path, parameter.Name, parameter.Parameters)
}

func (s DataService) PagingQuery(path, name string, parameters map[string]any, currentPageNo uint64, pageSize uint16) (viewmodels.PagingQueryResult, error) {
	s.repository.SetComponent(s.BusinessComponent)

	var result viewmodels.PagingQueryResult

	rowCount, err := s.repository.QueryScalar(path, name+".count", parameters)
	if nil != err {
		return result, err
	}

	count := uint64(rowCount.(int64))
	size := uint64(pageSize)

	beginNo := (currentPageNo - 1) * size

	if nil == parameters {
		parameters = make(map[string]any, 2)
	}

	if size >= count {
		parameters["begin_No"] = 0
		parameters["page_Size"] = count

		result.PageNo = 1
	} else if beginNo < count {
		parameters["begin_No"] = beginNo
		parameters["page_Size"] = count

		result.PageNo = currentPageNo
	} else {
		parameters["begin_No"] = beginNo

		pageNo := count / size
		parameters["page_Size"] = pageNo * size

		result.PageNo = pageNo
	}

	tables, err := s.repository.QueryTables(path, name, parameters)
	if nil != err {
		return result, err
	}

	result.Data = tables
	result.Count = count

	return result, nil
}

func (s DataService) PagingQueryByParameter(parameter viewmodels.PagingQueryParameter) (viewmodels.PagingQueryResult, error) {
	return s.PagingQuery(parameter.Path, parameter.Name, parameter.Parameters, parameter.CurrentPageNo, parameter.PageSize)
}

func (s DataService) saveActionId(tx *sqlx.Tx, actionId []byte) bool {
	var result bool
	defer func() {
		if p := recover(); nil != p {
			result = true
		}
	}()

	affected, err := s.repository.Update(tx, "system/frame", "insertAction", map[string]interface{}{"id": actionId})
	result = 0 >= affected || nil != err

	return result
}

func getParamMap(table models.SimpleData, index int) map[string]any {
	columns := table.Columns
	row := table.Rows[index]
	count := util.Min(len(columns), len(row))

	result := make(map[string]any, count)
	for i := 0; i < count; i++ {
		result[columns[i]] = row[i]
	}

	if _, ok := result["Sn"]; !ok {
		result["Sn"] = index
	}

	return result
}

func (s DataService) verifyTable(tx *sqlx.Tx, path, name string, table models.SimpleData, data map[string]models.SimpleData) (int64, error) {
	rowLen := len(table.Rows)
	fileIndex := 0
	for s.repository.IsSqlFileExist(path, name) {
		for index := 0; index < rowLen; index++ {
			relatedParam, err := s.repository.GetRelatedParam(path, name, data)
			if nil != err {
				return 0, err
			}
			rowLen = len(table.Rows)
			for i := 0; i < rowLen; i++ {
				param := util.MergeMaps[string, any](relatedParam, getParamMap(table, i))
				m, err := s.repository.QueryScalarForUpdate(tx, path, name, param)
				if nil != err {
					return -1, err
				}
				no := m.(string)
				if "" != no {
					errNo, err := strconv.ParseInt(no, 10, 64)
					if nil != err {
						return -1, err
					}
					return errNo, errors.New(components.GetMessage(path, errNo, param))
				}
			}
		}
		fileIndex++
		name = fmt.Sprintf("%s_%d", name, fileIndex)
	}

	return 0, nil
}

func (s DataService) verify(tx *sqlx.Tx, path, name string, data map[string]models.SimpleData) (int64, error) {
	for k, v := range data {
		sqlName := fmt.Sprintf("%s_%s.verify", name, k)
		errNo, err := s.verifyTable(tx, path, sqlName, v, data)
		if 0 != errNo || nil != err {
			return errNo, err
		}
	}
	return 0, nil
}

func (s DataService) Save(path, name string, data map[string]models.SimpleData, actionId []byte) (int64, error) {
	s.repository.SetComponent(s.BusinessComponent)

	return s.repository.DoInTransaction(func(tx *sqlx.Tx) (int64, error) {
		if nil != actionId && s.saveActionId(tx, actionId) {
			return 0, nil
		}

		errorNo, err := s.verify(tx, path, name, data)
		if 0 != errorNo {
			return -1, err
		}
		// var firstData models.SimpleData
		var count int64 = 0
		for k, v := range data {
			// if nil == firstData {
			// 	firstData = v
			// }

			sqlName := fmt.Sprintf("%s_%s", name, k)
			rowLen := len(v.Rows)
			relatedParam, err := s.repository.GetRelatedParam(path, sqlName, data)
			if nil != err {
				return -1, err
			}
			var sqlIndex int64 = 0
			for i := 0; i < rowLen; i++ {
				param := util.MergeMaps[string, any](relatedParam, getParamMap(v, i))
				for s.repository.IsSqlFileExist(path, sqlName) {
					rowAffected, err := s.repository.Update(tx, path, sqlName, param)
					if nil != err {
						return count, err
					}
					if 0 >= rowAffected {
						return -1, errors.New("并发冲突，数据没有保存，请稍后再试。")
					}
					count += rowAffected
					sqlIndex++
					sqlName = fmt.Sprintf("%s_%s_%d", name, k, sqlIndex)
				}
			}
		}

		sqlName := fmt.Sprintf("%s.after", name)
		if s.repository.IsSqlFileExist(path, sqlName) {
			rowAffected, err := s.repository.Update(tx, path, sqlName, nil)
			if nil != err {
				util.LogError(err)
				return -1, err
			}
			count += rowAffected
		}

		return count, nil
	})
}

func (s DataService) SaveByParameter(parameter viewmodels.SaveParameter) (int64, error) {
	return s.Save(parameter.Path, parameter.Name, parameter.Data, parameter.ActionId)
}

func (s DataService) saveTableData(tx *sqlx.Tx, path, name string, data models.SimpleData) (int64, error) {
	if !s.repository.IsSqlFileExist(path, name) {
		return 0, nil
	}

	var count int64 = 0
	rowsLen := len(data.Rows)
	for i := 0; i < rowsLen; i++ {
		affected, err := s.repository.Update(tx, path, name, getParamMap(data, i))
		if nil != err {
			return count, err
		}
		count += affected
	}

	return count, nil
}

func (s DataService) DifferentiatedSave(path, name string, data map[string]viewmodels.SaveData, actionId []byte) (int64, error) {
	s.repository.SetComponent(s.BusinessComponent)

	affected, err := s.repository.DoInTransaction(func(tx *sqlx.Tx) (int64, error) {
		if nil != actionId && s.saveActionId(tx, actionId) {
			return 0, nil
		}

		var count int64 = 0
		for k, v := range data {
			sqlName := fmt.Sprintf("%s_%s", name, k)

			if 0 < len(v.AddedData.Rows) {
				no, er := s.verifyTable(tx, path, fmt.Sprintf("%s.verify", sqlName), v.AddedData, nil)
				if nil != er {
					return no, er
				} else if 0 != no {
					return no, errors.New("Unknown error.")
				}

				aff, er := s.saveTableData(tx, path, fmt.Sprintf("%s.add", sqlName), v.AddedData)
				if nil != er {
					return count, er
				}
				count += aff
			}

			if 0 < len(v.DeletedData.Rows) {
				aff, er := s.saveTableData(tx, path, fmt.Sprintf("%s.delete", sqlName), v.AddedData)
				if nil != er {
					return count, er
				}
				count += aff
			}

			rowsLen := util.Min(len(v.ModifiedData.Rows), len(v.ModifiedOriginalData.Rows))
			sqlName = fmt.Sprintf("%s.modify", sqlName)
			for i := 0; i < rowsLen; i++ {
				param := getParamMap(v.ModifiedData, i)
				originalParam := getParamMap(v.ModifiedOriginalData, i)
				for pk, pv := range originalParam {
					param[fmt.Sprintf("Original_%s", pk)] = pv
				}
				aff, er := s.repository.Update(tx, path, sqlName, param)
				if nil != er {
					return count, er
				}
				count += aff
			}
		}

		return count, nil
	})
	if nil != err {
		return affected, err
	}

	return affected, nil
}

func (s DataService) DifferentiatedSaveByParameter(parameter viewmodels.DifferentiatedSaveParameter) (int64, error) {
	return s.DifferentiatedSave(parameter.Path, parameter.Name, parameter.Data, parameter.ActionId)
}
