// database
package repositories

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"server/components"
	"server/models"
	"server/util"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DataRepository struct {
	components.BusinessComponent
}

func (r *DataRepository) createDb() (*sqlx.DB, error) {
	return createDbByName(r.CurrentTenant.DbServerName)
}

func (r *DataRepository) createReadOnlyDb() (*sqlx.DB, error) {
	return createDbByName(r.CurrentTenant.ReadOnlyDbServerName)
}

func (r *DataRepository) getDriverName() (string, error) {
	return getDriverName(r.CurrentTenant.DbServerName)
}

func (r *DataRepository) getReadOnlyDriverName() (string, error) {
	return getDriverName(r.CurrentTenant.ReadOnlyDbServerName)
}

func (r *DataRepository) getSerialNo(no int64, tx *sqlx.Tx) string {
	if sequenceInfo, ok := sequenceInfoMap[no]; ok {
		var period int
		now := time.Now()
		switch sequenceInfo.PeriodType {
		case models.None:
			period = 0
		case models.Year:
			period = now.Year()
		case models.Quarter:
			period = now.Year()*10 + (int(now.Month())-1)/3
		case models.Month:
			period = now.Year()*100 + int(now.Month()) + toTenDays(now.Day())
		case models.TenDays:
			period = now.Year()*1000 + int(now.Month())*10
		case models.Day:
			period = now.Year()*10000 + int(now.Month())*100 + now.Day()
		case models.Hour:
			period = now.Year()*1000000 + int(now.Month())*10000 + now.Day()*100 + now.Hour()
		case models.QuarterHour:
			period = now.Year()*1000000 + now.YearDay()*1000 + now.Hour()*10 + now.Minute()/15
		case models.Minute:
			period = now.Year()*1000000 + (now.YearDay()-1)*1440 + now.Hour()*60 + now.Minute()
		}

		paramMap := make(map[string]any, 2)
		paramMap["no"] = no
		paramMap["period"] = period

		number, err := r.QueryScalarForUpdate(tx, "system/frame", "getSerialNo", paramMap)
		if nil != err {
			util.LogError(err)
			return ""
		}

		r.Update(tx, "system/frame", "insertSerialNo", paramMap)

		if nil == number {
			number = 1
		}

		if "" == sequenceInfo.Format {
			return strconv.FormatInt(number.(int64), 10)
		} else {
			return fmt.Sprintf(sequenceInfo.Format, period, number)
		}
	}

	return ""
}

func (r *DataRepository) getSqlByParam(path, name, driverName string, parameters map[string]any, tx *sqlx.Tx) (string, map[string]any, error) {
	sql, err := getSql(path, name, driverName)
	if nil != err {
		return "", parameters, err
	}

	sysParams, err := getSystemParamSettings(path, name, driverName)
	if nil != err {
		return "", parameters, err
	}
	if nil != sysParams && 0 < len(sysParams) {
		if nil == parameters {
			parameters = make(map[string]any)
		}
		for k, v := range sysParams {
			switch strings.Trim(v, " \r\n") {
			case "tenant":
				parameters[k] = r.CurrentTenant.Id
			case "user":
				parameters[k] = r.CurrentUserId
			case "uuid":
				id, err := uuid.New().MarshalBinary()
				if nil == err {
					parameters[k] = id
				} else {
					util.LogError(err)
				}
			case "currentTimeMillis", "currentDate":
				parameters[k] = time.Now()
			default:
				psp := strings.Split(v, ":")
				switch psp[0] {
				case "serialNo":
					no, err := strconv.Atoi(psp[1])
					if nil == err {
						parameters[k] = r.getSerialNo(int64(no), tx)
					} else {
						util.LogError(err)
					}
				}
			}
		}
	}

	params, err := getParamSettings(path, name, driverName)
	if nil != err {
		return "", parameters, err
	}

	for k, v := range params {
		if _, ok := parameters[k]; ok {
			sql = strings.ReplaceAll(sql, "{"+k+"}", v)
		} else {
			sql = strings.ReplaceAll(sql, "{"+k+"}", "")
		}
	}

	for k, v := range parameters {
		if reflect.ValueOf(v).Kind() == reflect.Map {
			parameters[k] = nil
		}
		// switch reflect.ValueOf(v).Kind() {
		// case reflect.Map:
		// 	parameters[k] = nil
		// case reflect.String:
		// 	if str, ok := v.(string); ok {
		// 		if strings.HasSuffix(k, "_Date_Time") {
		// 			t, err := time.Parse("2001-01-01 13:01:01", str)
		// 			if err != nil {
		// 				parameters[k] = t
		// 			}
		// 		} else if strings.HasSuffix(k, "_Date") {
		// 			t, err := time.Parse("2001-01-01", str)
		// 			if err != nil {
		// 				parameters[k] = t
		// 			}
		// 		} else if strings.HasSuffix(k, "_Time") {
		// 			t, err := time.Parse("2001-01-01", str)
		// 			if err != nil {
		// 				parameters[k] = t
		// 			}
		// 		}
		// 	}
		// }
	}

	return sql, parameters, nil
}

func (r *DataRepository) getSql(path, name string, parameters map[string]any, tx *sqlx.Tx) (string, map[string]any, error) {
	driverName, err := r.getDriverName()
	if nil != err {
		return "", nil, err
	}
	return r.getSqlByParam(path, name, driverName, parameters, tx)
}

func (r *DataRepository) getReadOnlySql(path, name string, parameters map[string]any) (string, map[string]any, error) {
	driverName, err := r.getReadOnlyDriverName()
	if nil != err {
		return "", parameters, err
	}
	return r.getSqlByParam(path, name, driverName, parameters, nil)
}

type QueryRowHandler = func(int64, int64, *sqlx.Rows) error

type QueryHandler = func(int64, *sqlx.Rows) error

func (r *DataRepository) Query(path, name string, parameters map[string]any, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error {
	db, err := r.createReadOnlyDb()
	if nil != err {
		util.LogError(err)
		return err
	}
	defer db.Close()

	sql, parameters, err := r.getReadOnlySql(path, name, parameters)
	if nil != err {
		return err
	}

	rows, err := db.NamedQuery(sql, parameters)
	if nil != err {
		util.LogError(path)
		util.LogError(name)
		util.LogError(sql)
		util.LogError(err)
		return err
	}
	defer rows.Close()

	return tablesScan(rows, beforHandler, afterHandler, handler)
}

func (r *DataRepository) QueryForUpdate(tx *sqlx.Tx, path, name string, parameters map[string]any, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error {
	sql, parameters, err := r.getSql(path, name, parameters, tx)
	if nil != err {
		return err
	}

	rows, err := tx.NamedQuery(sql, parameters)
	if nil != err {
		util.LogError(sql)
		util.LogError(err)
		return err
	}
	defer rows.Close()

	return tablesScan(rows, beforHandler, afterHandler, handler)
}

func (r *DataRepository) QueryTables(path, name string, parameters map[string]any) (map[string]models.SimpleData, error) {
	var row []any
	var err error
	var table models.SimpleData
	var tableName string
	result := make(map[string]models.SimpleData, 0)
	err = r.Query(path, name, parameters, func(index int64, rows *sqlx.Rows) error {
		columns, err := rows.Columns()
		if nil != err {
			return err
		}
		table = models.SimpleData{}
		if 0 == index {
			tableName = name
		} else {
			tableName = fmt.Sprintf("%s_%d", name, index)
		}
		table.Columns = columns
		table.Rows = make([][]any, 0)
		return nil
	}, func(_ int64, rows *sqlx.Rows) error {
		result[tableName] = table
		// result = append(result, table)
		return nil
	}, func(_, _ int64, rows *sqlx.Rows) error {
		row, err = rows.SliceScan()
		if nil != err {
			return err
		}
		table.Rows = append(table.Rows, amend(row))
		return nil
	})

	return result, nil
}

func (r *DataRepository) QueryScalar(path, name string, parameters map[string]any) (any, error) {
	var result any
	var err error
	err = r.Query(path, name, parameters, nil, nil, func(_, _ int64, rows *sqlx.Rows) error {
		var row []any
		row, err = rows.SliceScan()
		if nil != err {
			return err
		}
		result = row[0]
		return nil
	})

	return result, err
}

func (r *DataRepository) QueryScalarForUpdate(tx *sqlx.Tx, path, name string, parameters map[string]any) (any, error) {
	var result any
	var err error
	err = r.QueryForUpdate(tx, path, name, parameters, nil, nil, func(_, _ int64, rows *sqlx.Rows) error {
		var row []any
		row, err = rows.SliceScan()
		if nil != err {
			return err
		}
		result = row[0]
		return nil
	})

	return result, err
}

type ExecHandler = func(*sqlx.Tx) (int64, error)

func (r *DataRepository) DoInTransaction(handler ExecHandler) (int64, error) {
	db, err := r.createDb()
	if nil != err {
		util.LogError(err)
		return -1, err
	}
	defer db.Close()

	var affected int64 = 0
	tx, err := db.Beginx() // 开启事务
	if err != nil {
		util.LogError(err)
		return -1, err
	}
	var txErr error
	defer func() {
		if p := recover(); nil != p {
			tx.Rollback()
			log.Panic(p)
			//util.LogError(err)
			// panic(p) // re-throw panic after Rollback
		} else if txErr == nil {
			err = tx.Commit() // err is nil; if Commit returns error update err
			if nil != err {
				util.LogError(err)
			}
		} else {
			err = tx.Rollback() // err is non-nil; don't change it
			if nil != err {
				util.LogError(err)
			}
		}
	}()

	affected, txErr = handler(tx)

	return affected, txErr
}

func (r *DataRepository) Update(tx *sqlx.Tx, path, name string, parameters map[string]any) (int64, error) {
	sql, parameters, err := r.getSql(path, name, parameters, tx)
	if nil != err {
		return -1, err
	}

	var count int64 = 0
	sqls := strings.Split(sql, ";")
	for _, updateSql := range sqls {
		updateSql = strings.Trim(updateSql, " ")
		if "" == updateSql {
			continue
		}
		result, err := tx.NamedExec(updateSql, parameters)
		if nil != err {
			util.LogError(err)
			return -1, err
		}
		affected, err := result.RowsAffected()
		if nil != err {
			util.LogError(err)
			return -1, err
		}
		count += affected
	}
	return count, nil

	// result, err := tx.NamedExec(sql, parameters)
	// if nil != err {
	// 	util.LogError(err)
	// 	return -1, err
	// }
	// affected, err := result.RowsAffected()
	// if nil != err {
	// 	util.LogError(err)
	// 	return -1, err
	// }
	// return affected, nil
}

func (r *DataRepository) IsSqlFileExist(path, name string) bool {
	return isSqlFileExist(path, name, r.CurrentTenant.DbServerName)
}

func (r *DataRepository) IsSqlFileExistReadOnly(path, name string) bool {
	return isSqlFileExist(path, name, r.CurrentTenant.ReadOnlyDbServerName)
}

func (r *DataRepository) isRelatedParamFileExist(path, name string) bool {
	return isRelatedParamFileExist(path, name, r.CurrentTenant.DbServerName)
}

func (r *DataRepository) getRelatedParamSettings(path, name string) (map[string]string, error) {
	return getRelatedParamSettings(path, name, r.CurrentTenant.DbServerName)
}

func (r *DataRepository) GetRelatedParam(path, name string, data map[string]models.SimpleData) (map[string]any, error) {
	if !r.isRelatedParamFileExist(path, name) || nil == data || 0 == len(data) {
		return nil, nil
	}

	settings, err := r.getRelatedParamSettings(path, name)
	if nil != err {
		return nil, err
	}

	result := make(map[string]any)
	for k, v := range settings {
		if table, ok := data[v]; ok && 0 < len(table.Rows) {
			rows := table.Rows
			columns := table.Columns
			columnLen := len(columns)
			for i := 0; i < columnLen; i++ {
				if k == columns[i] {
					result[k] = rows[i]
					break
				}
			}
		}
	}

	return result, nil
}
