// business
package components

import (
	"server/models"

	"github.com/kataras/iris/v12"
)

type BusinessComponent struct {
	CurrentTenant models.Tenant
	CurrentUserId int64
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
}
