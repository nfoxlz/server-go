// frame
package controllers

import (
	"time"

	"server/models"
	"server/repositories"
	"server/services"

	"github.com/kataras/iris/v12"
)

type FrameController struct {
	Ctx     iris.Context
	service services.FrameService
}

func (c *FrameController) PostMenus() ([]models.Menu, error) {
	defer errorExit()

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

func (c *FrameController) PostClearCache() {
	repositories.InitCache()
}
