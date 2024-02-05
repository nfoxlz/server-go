// interceptors
package web

import (
	"log"
	"strconv"

	"server/components"
	"server/repositories"
	"server/util"

	"github.com/kataras/iris/v12"
)

func getCookie(ctx iris.Context, name string) (int64, error) {
	cookie, err := ctx.GetRequestCookie(name)
	if nil != err {
		log.Println(err)
		return 0, err
	}

	value, err := util.AESDecryptWithRandString(cookie.Value)
	if nil != err {
		return 0, err
	}

	id, err := strconv.ParseInt(string(value), 10, 64)
	if nil != err {
		log.Println(err)
		return 0, err
	}

	return id, nil
}

func AuthorizationInterceptor() iris.Handler {
	return func(ctx iris.Context) {
		if "/api/Account/authenticate" != ctx.Path() {
			tenantId, err := getCookie(ctx, components.TenantToken)
			if nil != err {
				return
			}

			var repository repositories.TenantRepository
			tenant, err := repository.GetTenant(tenantId)
			if nil != err || tenant.Id == 0 {
				return
			}

			userId, err := getCookie(ctx, components.UserToken)
			if nil != err {
				return
			}

			ctx.Values().Set(components.TenantToken, tenant)
			ctx.Values().Set(components.UserToken, userId)
		}

		ctx.Next()
	}
}
