// tenant
package repositories

import (
	"errors"
	"server/models"
	"server/util"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type TenantRepository struct {
}

func (r *TenantRepository) GetTenant(id int64) (models.Tenant, error) {
	var result models.Tenant
	db, err := sqlx.Open(defaultConfig.DriverName, defaultConfig.DataSourceName)
	if nil != err {
		util.LogError(err)
		return result, err
	}
	defer db.Close()

	sql, err := getSql("system/common", "getTenant", defaultConfig.DriverName)
	if nil != err {
		util.LogError(err)
		return result, err
	}
	// row := db.QueryRow(sql, id)
	// err = row.Scan(&result.Id, &result.Code, &result.Name, &result.DbServerName, &result.ReadOnlyDbServerName)
	// if nil != err {
	// 	util.LogError(err)
	// }
	rows, err := db.NamedQuery(sql, map[string]any{"id": id})
	if nil != err {
		util.LogError(err)
		return result, err
	}
	if rows.Next() {
		err = rows.Scan(&result.Id, &result.Code, &result.Name, &result.DbServerName, &result.ReadOnlyDbServerName)
		if nil != err {
			util.LogError(err)
		}
	} else {
		err = errors.New("no rows in result set")
	}

	return result, err
}

func (r *TenantRepository) GetTenantByCode(code string) (models.Tenant, error) {
	var result models.Tenant
	db, err := sqlx.Open(defaultConfig.DriverName, defaultConfig.DataSourceName)
	if nil != err {
		util.LogError(err)
		return result, err
	}
	defer db.Close()

	sql, err := getSql("system/common", "getTenantByCode", defaultConfig.DriverName)
	if nil != err {
		util.LogError(err)
		return result, err
	}
	// row := db.QueryRow(sql, code)
	// err = row.Scan(&result.Id, &result.Code, &result.Name, &result.DbServerName, &result.ReadOnlyDbServerName)
	// if nil != err {
	// 	util.LogError(err)
	// }
	rows, err := db.NamedQuery(sql, map[string]any{"code": code})
	if nil != err {
		util.LogError(err)
		return result, err
	}
	if rows.Next() {
		err = rows.Scan(&result.Id, &result.Code, &result.Name, &result.DbServerName, &result.ReadOnlyDbServerName)
		if nil != err {
			util.LogError(err)
		}
	} else {
		err = errors.New("no rows in result set")
	}

	return result, err
}

func (r *TenantRepository) GetUser(tenant models.Tenant, code string) (models.User, error) {
	var result models.User

	dbConfig, err := getConfig(tenant.ReadOnlyDbServerName)
	if nil != err {
		return result, err
	}

	db, err := sqlx.Open(dbConfig.DriverName, dbConfig.DataSourceName)
	if nil != err {
		util.LogError(err)
		return result, err
	}
	defer db.Close()

	sql, err := getSql("system/common", "getUserByCode", defaultConfig.DriverName)
	if nil != err {
		util.LogError(err)
		return result, err
	}
	// row := db.QueryRow(sql, code)
	// err = row.Scan(&result.Id, &result.Code, &result.Name, &result.Role.Id, &result.UserPassword)
	// if nil != err {
	// 	util.LogError(err)
	// }
	rows, err := db.NamedQuery(sql, map[string]any{"tenant_Id": tenant.Id, "code": code})
	if nil != err {
		util.LogError(err)
		return result, err
	}
	if rows.Next() {
		err = rows.Scan(&result.Id, &result.Code, &result.Name, &result.Role.Id, &result.UserPassword)
		if nil != err {
			util.LogError(err)
		}
		result.Tenant = tenant
	} else {
		err = errors.New("no rows in result set")
	}

	return result, err
}
