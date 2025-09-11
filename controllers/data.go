// data
package controllers

import (
	"server/services"
	"server/util"
	"server/viewmodels"

	"github.com/kataras/iris/v12"
)

type DataController struct {
	Ctx     iris.Context
	service services.DataService
}

func (c *DataController) PostQuery(parameter viewmodels.QueryParameter) viewmodels.QueryResult {
	defer errorExit()

	if !c.service.CheckSign(&parameter, c.Ctx) {
		return viewmodels.QueryResult{
			Result: viewmodels.Result{
				ErrorNo: -1,
				Message: "签名验证失败",
			},
		}
	}
	// c.service.SetContext(c.Ctx)

	data, err := c.service.QueryByParameter(parameter)
	result := viewmodels.QueryResult{Data: data}
	if nil != err {
		result.ErrorNo, result.Message = util.ExtractMessage(err)
	}

	return result
}

func (c *DataController) PostPagingQuery(parameter viewmodels.PagingQueryParameter) viewmodels.PagingQueryResult {
	defer errorExit()

	util.LogDebug("QueryParameter:", parameter)

	if !c.service.CheckSign(&parameter, c.Ctx) {
		return viewmodels.PagingQueryResult{
			QueryResult: viewmodels.QueryResult{
				Result: viewmodels.Result{
					ErrorNo: -1,
					Message: "签名验证失败",
				},
			},
		}
	}

	// c.service.SetContext(c.Ctx)
	return c.service.PagingQueryByParameter(parameter)
}

func (c *DataController) PostSave(parameter viewmodels.SaveParameter) viewmodels.Result {
	defer errorExit()

	if !c.service.CheckSign(&parameter, c.Ctx) {
		return viewmodels.Result{
			ErrorNo: -1,
			Message: "签名验证失败",
		}
	}
	// c.service.SetContext(c.Ctx)
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

	if !c.service.CheckSign(&parameter, c.Ctx) {
		return viewmodels.Result{
			ErrorNo: -1,
			Message: "签名验证失败",
		}
	}
	// c.service.SetContext(c.Ctx)
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
