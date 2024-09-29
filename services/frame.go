// frame
package services

import (
	"time"

	"server/models"
	"server/repositories"

	"github.com/jmoiron/sqlx"
)

type FrameService struct {
	BusinessService
}

// func exec() error {
// 	return nil
// }

func (s *FrameService) GetMenus() ([]models.Menu, error) {
	s.repository.SetComponent(s.BusinessComponent)

	parameters := make(map[string]any)
	parameters["application"] = 0
	parameters["client_Side"] = 0
	// parameters["user_Id"] = s.CurrentUserId

	var columns []string
	menus := make([]models.Menu, 0)
	err := s.repository.Query("system/frame", "getMenus", parameters, func(_ int64, rows *sqlx.Rows) error {
		var result error
		columns, result = rows.Columns()
		return result
	}, nil, func(_, _ int64, rows *sqlx.Rows) error {
		menu := models.Menu{}
		result := repositories.StructScan(rows, columns, &menu)
		if nil != result {
			return result
		}

		menus = append(menus, menu)
		return nil
	})
	if nil != err {
		return nil, err
	}

	return menus, nil
}

func (s *FrameService) GetEnums() ([]models.EnumInfo, error) {
	s.repository.SetComponent(s.BusinessComponent)

	var columns []string
	enums := make([]models.EnumInfo, 0)
	err := s.repository.Query("system/frame", "getEnums", nil, func(_ int64, rows *sqlx.Rows) error {
		var result error
		columns, result = rows.Columns()
		return result
	}, nil, func(_, _ int64, rows *sqlx.Rows) error {
		enum := models.EnumInfo{}
		result := repositories.StructScan(rows, columns, &enum)
		if nil != result {
			return result
		}

		enums = append(enums, enum)
		return nil
	})
	if nil != err {
		return nil, err
	}

	return enums, nil
}

func (s *FrameService) GetConfigurations() (map[string]string, error) {
	s.repository.SetComponent(s.BusinessComponent)

	configurations := make(map[string]string)
	err := s.repository.Query("system/frame", "getConfigurations", nil, nil, nil, func(_, _ int64, rows *sqlx.Rows) error {
		row, err := rows.SliceScan()
		if nil != err {
			return err
		}
		configurations[row[0].(string)] = row[1].(string)

		return nil
	})
	if nil != err {
		return nil, err
	}

	return configurations, nil
}

func (s *FrameService) GetServerDateTime() (time.Time, error) {
	s.repository.SetComponent(s.BusinessComponent)

	// result, err := s.repository.QueryScalar("system/frame", "getServerDateTime", nil)
	// if nil != err {
	// 	return time.Time{}, err
	// }

	return s.repository.GetServerDateTime()
}

func (s *FrameService) GetAccountingDate() (time.Time, error) {
	s.repository.SetComponent(s.BusinessComponent)

	return s.repository.GetAccountingDate()
}
