// database
package repositories

import (
	"errors"
	"fmt"
	"log"
	"os"
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

func (r *DataRepository) GetAccountingDate() (time.Time, error) {

	number, err := r.QueryScalar("system/frame", "getAccountingDate", nil)
	if nil == err {
		if nil == number {
			return time.Time{}, nil
		}

		numberStr, ok := number.([]uint8)
		if ok {
			date, err := strconv.Atoi(string(numberStr))
			if nil == err {
				year := date / 10000
				monthNum := date/100 - year*100
				return time.Date(year, time.Month(monthNum), date-year*10000-monthNum*100, 0, 0, 0, 0, time.UTC), nil
			} else {
				util.LogError(numberStr)
			}
		} else {
			err = errors.New("Cannot be converted to integer type.")
		}
		util.LogError(number)
	}

	util.LogError(err)
	return time.Time{}, err
}

func (r *DataRepository) GetServerDateTime() (time.Time, error) {
	result, err := r.QueryScalar("system/frame", "getServerDateTime", nil)
	if nil != err {
		return time.Time{}, err
	}

	return result.(time.Time), nil
}

func (r *DataRepository) getSerialNoByTime(no int64, accountingDate time.Time, tx *sqlx.Tx) string {
	if sequenceInfo, ok := sequenceInfoMap[no]; ok {
		var period int
		switch sequenceInfo.PeriodType {
		case models.None:
			period = 0
		case models.Year:
			period = accountingDate.Year()
		case models.Quarter:
			period = accountingDate.Year()*10 + (int(accountingDate.Month())-1)/3
		case models.Month:
			period = accountingDate.Year()*100 + int(accountingDate.Month())
		case models.TenDays:
			period = accountingDate.Year()*1000 + int(accountingDate.Month())*10 + toTenDays(accountingDate.Day())
		case models.Day:
			period = accountingDate.Year()*10000 + int(accountingDate.Month())*100 + accountingDate.Day()
		case models.Hour:
			period = accountingDate.Year()*1000000 + int(accountingDate.Month())*10000 + accountingDate.Day()*100 + accountingDate.Hour()
		case models.QuarterHour:
			period = accountingDate.Year()*1000000 + accountingDate.YearDay()*1000 + accountingDate.Hour()*10 + accountingDate.Minute()/15
		case models.Minute:
			period = accountingDate.Year()*1000000 + (accountingDate.YearDay()-1)*1440 + accountingDate.Hour()*60 + accountingDate.Minute()
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
			number = int64(1)
		}

		if "" == sequenceInfo.Format {
			return strconv.FormatInt(number.(int64), 10)
		} else {
			return fmt.Sprintf(sequenceInfo.Format, period, number)
		}
	}

	return ""
}

func (r *DataRepository) getSerialNo(no int64, tx *sqlx.Tx) string {
	serverDateTime, _ := r.GetServerDateTime()
	return r.getSerialNoByTime(no, serverDateTime, tx)
}

func (r *DataRepository) getSqlByParam(path, name, driverName string, parameters map[string]any, sortDescription string, tx *sqlx.Tx) (string, map[string]any, error) {
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
					return "", parameters, err
				}
			case "currentTimeMillis":
				parameters[k] = time.Now()
			case "currentDate":
				r.GetServerDateTime()
				year, month, day := time.Now().Date()
				parameters[k] = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			case "accountingDate":
				accountingDate, err := r.GetAccountingDate()
				if nil == err {
					parameters[k] = accountingDate
				} else {
					util.LogError(err)
					return "", parameters, err
				}
			default:
				psp := strings.Split(v, ":")
				switch psp[0] {
				case "serialNo":
					no, err := strconv.ParseInt(psp[1], 10, 64)
					if nil == err {
						parameters[k] = r.getSerialNo(int64(no), tx)
					} else {
						util.LogError(err)
						return "", parameters, err
					}
				case "accountingSerialNo":
					accountingDate, err := r.GetAccountingDate()
					if nil != err {
						util.LogError(err)
						return "", parameters, err
					}

					no, err := strconv.ParseInt(psp[1], 10, 64)
					if nil == err {
						parameters[k] = r.getSerialNoByTime(int64(no), accountingDate, tx)
					} else {
						util.LogError(err)
						return "", parameters, err
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

	if "" == sortDescription {
		sql = strings.ReplaceAll(sql, "{order_By}", "")
	} else {
		if strings.Contains(strings.ToUpper(sql), "ORDER BY") {
			sql = strings.ReplaceAll(sql, "{order_By}", " "+sortDescription+", ")
		} else {
			sql = strings.ReplaceAll(sql, "{order_By}", " ORDER BY "+sortDescription+" ")
		}
	}

	return sql, parameters, nil
}

func (r *DataRepository) getSqlName(path, name string) (string, error) {
	driverName, err := r.getDriverName()
	if nil != err {
		return "", err
	}

	return getSqlName(path, name, driverName), nil
}

func (r *DataRepository) getSql(path, name string, parameters map[string]any, sortDescription string, tx *sqlx.Tx) (string, map[string]any, error) {
	driverName, err := r.getDriverName()
	if nil != err {
		return "", nil, err
	}
	return r.getSqlByParam(path, name, driverName, parameters, sortDescription, tx)
}

func (r *DataRepository) getReadOnlySql(path, name string, parameters map[string]any, sortDescription string) (string, map[string]any, error) {
	driverName, err := r.getReadOnlyDriverName()
	if nil != err {
		return "", parameters, err
	}
	return r.getSqlByParam(path, name, driverName, parameters, sortDescription, nil)
}

type QueryRowHandler = func(int64, int64, *sqlx.Rows) error

type QueryHandler = func(int64, *sqlx.Rows) error

func (r *DataRepository) Query(path, name string, parameters map[string]any, sortDescription string, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error {
	db, err := r.createReadOnlyDb()
	if nil != err {
		util.LogError(err)
		return err
	}
	defer db.Close()

	sql, parameters, err := r.getReadOnlySql(path, name, parameters, sortDescription)
	if nil != err {
		return err
	}

	util.LogDebug("SQL: ", sql)
	util.LogDebug("parameters: ", parameters)

	rows, err := db.NamedQuery(sql, parameters)
	if nil != err {
		util.LogError(path, name, err)
		return err
	}
	defer rows.Close()

	return tablesScan(rows, beforHandler, afterHandler, handler)
}

func (r *DataRepository) QueryUseTransaction(path, name string, parameters map[string]any, sortDescription string, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error {

	_, err := r.DoInTransaction(func(tx *sqlx.Tx) (int64, error) {
		return 0, r.QueryForUpdate(tx, path, name, parameters, sortDescription, beforHandler, afterHandler, handler)
	})

	return err
}

func (r *DataRepository) QueryForUpdate(tx *sqlx.Tx, path, name string, parameters map[string]any, sortDescription string, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error {
	sql, parameters, err := r.getSql(path, name, parameters, sortDescription, tx)
	if nil != err {
		return err
	}

	util.LogDebug("SQL: ", sql)
	util.LogDebug("parameters: ", parameters)

	rows, err := tx.NamedQuery(sql, parameters)
	if nil != err {
		util.Log.LogError(path, name, err)
		return err
	}
	defer rows.Close()

	return tablesScan(rows, beforHandler, afterHandler, handler)
}

type queryExecuteHandler = func(path, name string, parameters map[string]any, sortDescription string, beforHandler, afterHandler QueryHandler, handler QueryRowHandler) error

func (r *DataRepository) QueryTables(path, name string, parameters map[string]any, sortDescription string) ([]models.SimpleData, error) {

	sqlName := name
	fileName, err := r.getSqlName(path, sqlName)
	if nil != err {
		return nil, err
	}

	config, err := getSqlConfig(path, sqlName)
	if nil != err {
		return nil, err
	}

	var row []any
	var table models.SimpleData
	result := make([]models.SimpleData, 0)
	i := 0
	_, filErr := os.Stat(fileName)

	var handler queryExecuteHandler
	if config.UseTransaction {
		handler = r.QueryUseTransaction
	} else {
		handler = r.Query
	}

	for filErr == nil {
		err = handler(path, sqlName, parameters, sortDescription, func(index int64, rows *sqlx.Rows) error {
			columns, err := rows.Columns()
			if nil != err {
				return err
			}
			table = models.SimpleData{}
			if 0 != index {
				sqlName = fmt.Sprintf("%s_%d", sqlName, index)
			}
			table.TableName = sqlName
			table.Columns = columns
			table.Rows = make([][]any, 0)
			return nil
		}, func(_ int64, rows *sqlx.Rows) error {
			// result[sqlName] = table
			result = append(result, table)
			return nil
		}, func(_, _ int64, rows *sqlx.Rows) error {
			row, err = rows.SliceScan()
			if nil != err {
				return err
			}
			table.Rows = append(table.Rows, amend(row))
			return nil
		})

		if nil != err {
			return result, err
		}

		i++
		sqlName = fmt.Sprintf("%s_%d", name, i)
		fileName, err = r.getSqlName(path, sqlName)
		if nil != err {
			return result, err
		}
		_, filErr = os.Stat(fileName)
	}

	return result, err
}

func (r *DataRepository) QueryScalar(path, name string, parameters map[string]any) (any, error) {
	var result any
	var err error
	err = r.Query(path, name, parameters, "", nil, nil, func(_, _ int64, rows *sqlx.Rows) error {
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
	err = r.QueryForUpdate(tx, path, name, parameters, "", nil, nil, func(_, _ int64, rows *sqlx.Rows) error {
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
			// util.LogError(err)
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
			// util.LogError(txErr)
		}
	}()

	affected, txErr = handler(tx)

	return affected, txErr
}

func (r *DataRepository) Update(tx *sqlx.Tx, path, name string, parameters map[string]any) (int64, error) {
	sql, parameters, err := r.getSql(path, name, parameters, "", tx)
	if nil != err {
		return -1, err
	}

	// util.LogDebug("SQL: ", sql)

	var count int64 = 0
	sqls := strings.Split(sql, ";")
	for _, updateSql := range sqls {
		updateSql = strings.TrimSpace(updateSql)
		if "" == updateSql {
			continue
		}

		util.LogDebug("SQL: ", updateSql)
		util.LogDebug("parameters: ", parameters)

		result, err := tx.NamedExec(updateSql, parameters)
		if nil != err {
			util.LogError(path, name, err)
			return -1, err
		}
		affected, err := result.RowsAffected()
		if nil != err {
			util.LogError(path, name, err)
			return -1, err
		}
		count += affected
	}
	return count, nil
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

func (r *DataRepository) GetRelatedParam(path, name string, data []models.SimpleData) (map[string]any, error) {
	if !r.isRelatedParamFileExist(path, name) || nil == data || 0 == len(data) {
		return nil, nil
	}

	settings, err := r.getRelatedParamSettings(path, name)
	if nil != err {
		return nil, err
	}

	result := make(map[string]any)
	for k, v := range settings {
		for _, table := range data {
			if table.TableName == v && 0 < len(table.Rows) {
				rows := table.Rows
				columns := table.Columns
				columnLen := len(columns)
				for i := 0; i < columnLen; i++ {
					if k == columns[i] {
						result[k] = rows[0][i]
						break
					}
				}
				break
			}
		}
	}

	return result, nil
}
