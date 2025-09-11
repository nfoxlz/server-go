// business
package components

import (
	"server/models"

	"github.com/kataras/iris/v12"
)

type BusinessComponent struct {
	CurrentTenant models.Tenant
	CurrentUserId int64
	// Ctx           iris.Context
}

func (c *BusinessComponent) CheckSign(viewModel interface{}, ctx iris.Context) bool {

	result := true //c.CheckSignAuthenticate(viewModel, ctx)	// 禁用数据签名。

	if result {
		c.SetContext(ctx)
	}

	return result
}

func (c *BusinessComponent) CheckSignAuthenticate(viewModel interface{}, ctx iris.Context) bool {

	return true

	// 禁用数据签名。
	// util.ModifyField(viewModel, "SignPassword", "F4AE7A53-01EB-4693-8A8A-37753D4B044B")
	// util.ModifyField(viewModel, "Timestamp", time.Now().Format("20060102"))

	// jsonData, err := json.Marshal(viewModel)
	// if err != nil {
	// 	return false
	// }

	// return util.Verify(string(jsonData), ctx.Values().Get(DataSignature).(string))
}

func (c *BusinessComponent) SetContext(ctx iris.Context) {
	var ok bool
	if c.CurrentTenant, ok = (ctx.Values().Get(TenantToken)).(models.Tenant); ok {
		c.CurrentUserId, _ = (ctx.Values().Get(UserToken)).(int64)
	}
}

func (c *BusinessComponent) SetTenantUser(tenant models.Tenant, userId int64) {
	c.CurrentTenant = tenant
	c.CurrentUserId = userId
}

func (c *BusinessComponent) SetComponent(component BusinessComponent) {
	c.CurrentTenant = component.CurrentTenant
	c.CurrentUserId = component.CurrentUserId
	// c.Ctx = component.Ctx
}
