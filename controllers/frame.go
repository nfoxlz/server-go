// frame
package controllers

import (
	"errors"
	"time"

	"server/models"
	"server/repositories"
	"server/services"
	"server/util"
	"server/viewmodels"

	"github.com/kataras/iris/v12"
)

type FrameController struct {
	Ctx     iris.Context
	service services.FrameService
}

func (c *FrameController) PostMenus() ([]models.Menu, error) {
	defer errorExit()

	util.LogDebug("PostMenus")
	c.service.SetContext(c.Ctx)
	return c.service.GetMenus()
}

func (c *FrameController) PostEnums() ([]models.EnumInfo, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	return c.service.GetEnums()
}

func (c *FrameController) PostServerDateTime() (time.Time, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	return c.service.GetServerDateTime()
}

func (c *FrameController) PostAccountingDate() (time.Time, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	return c.service.GetAccountingDate()
}

func (c *FrameController) PostSettings() (map[string]string, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	return c.service.GetSettings()
}

func (c *FrameController) PostModifyPassword(parameter viewmodels.ModifyPasswordParameter) (string, error) {
	defer errorExit()

	if !c.service.CheckSign(&parameter, c.Ctx) {
		return "false", errors.New("签名验证失败")
	}
	// c.service.SetContext(c.Ctx)
	isSuccessful, err := c.service.ModifyPassword(parameter.OriginalPassword, parameter.NewPassword)
	var result string
	if isSuccessful {
		result = "true"
	} else {
		result = "false"
	}

	return result, err
}

func (c *FrameController) PostIsFinanceClosed() (string, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	isClose, err := c.service.IsFinanceClosed()
	var result string
	if isClose {
		result = "true"
	} else {
		result = "false"
	}

	return result, err
}

func (c *FrameController) PostIsFinanceClosedByDate(parameter viewmodels.PeriodYearMonthParameter) (string, error) {
	defer errorExit()

	if !c.service.CheckSign(&parameter, c.Ctx) {
		return "false", errors.New("签名验证失败")
	}
	// c.service.SetContext(c.Ctx)
	isClose, err := c.service.IsFinanceClosedByDate(parameter.PeriodYearMonth)
	var result string
	if isClose {
		result = "true"
	} else {
		result = "false"
	}

	return result, err
}

func (c *FrameController) PostClearCache() {
	defer errorExit()

	repositories.InitCache()
}
