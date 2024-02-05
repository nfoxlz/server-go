// account
package controllers

import (
	"fmt"
	"log"

	"server/models"
	"server/services"
	"server/util"
	"server/viewmodels"

	"github.com/kataras/iris/v12"
)

type AccountController struct {
	Ctx     iris.Context
	service services.AccountService
}

func (c *AccountController) PostAuthenticate(model viewmodels.LoginViewModel) (models.User, error) {
	defer errorExit()

	result, err := c.service.AuthenticateAndGetUser(model)
	if nil != err {
		return result, err
	}

	token, err := util.AESEncryptWithRandToString([]byte(fmt.Sprintf("%d", result.Tenant.Id)))
	if nil != err {
		return result, err
	}
	// c.Ctx.RemoveCookie("TENANT_TOKEN")
	c.Ctx.SetCookieKV("TENANT_TOKEN", token)

	token, err = util.AESEncryptWithRandToString([]byte(fmt.Sprintf("%d", result.Id)))
	if nil != err {
		return result, err
	}
	// c.Ctx.RemoveCookie("USER_TOKEN")
	c.Ctx.SetCookieKV("USER_TOKEN", token)

	return result, nil
}

func (c *AccountController) Get() string {

	log.Println(c.Ctx.Values().Get("TENANT_TOKEN"))
	log.Println(c.Ctx.Values().Get("USER_TOKEN"))

	return "OK"
}
