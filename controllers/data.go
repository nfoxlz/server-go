// data
package controllers

import (
	"server/models"
	"server/services"
	"server/viewmodels"

	"github.com/kataras/iris/v12"
)

type DataController struct {
	Ctx     iris.Context
	service services.DataService
}

func (c *DataController) PostQuery(parameter viewmodels.QueryParameter) (map[string]models.SimpleData, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	return c.service.QueryByParameter(parameter)
}

func (c *DataController) PostPagingQuery(parameter viewmodels.PagingQueryParameter) (viewmodels.PagingQueryResult, error) {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	return c.service.PagingQueryByParameter(parameter)
}

func (c *DataController) PostSave(parameter viewmodels.SaveParameter) viewmodels.Result {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	no, err := c.service.SaveByParameter(parameter)
	if nil == err {
		return viewmodels.Result{
			ErrorNo: no,
			Message: "",
		}
	} else {
		return viewmodels.Result{
			ErrorNo: no,
			Message: err.Error(),
		}
	}
}

func (c *DataController) PostDifferentiatedSave(parameter viewmodels.DifferentiatedSaveParameter) viewmodels.Result {
	defer errorExit()

	c.service.SetContext(c.Ctx)
	no, err := c.service.DifferentiatedSaveByParameter(parameter)
	if nil == err {
		return viewmodels.Result{
			ErrorNo: no,
			Message: "",
		}
	} else {
		return viewmodels.Result{
			ErrorNo: no,
			Message: err.Error(),
		}
	}
}
