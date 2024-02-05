// account
package services

import (
	"server/models"
	"server/repositories"
	"server/util"
	"server/viewmodels"
)

type AccountService struct {
}

func (s *AccountService) AuthenticateAndGetUser(model viewmodels.LoginViewModel) (models.User, error) {
	var user models.User
	var repository repositories.TenantRepository
	tenant, err := repository.GetTenantByCode(model.Tenant)
	if err != nil {
		return user, err
	}

	result, err := repository.GetUser(tenant, model.User)
	if err != nil {
		return user, err
	}

	if util.Verify(model.Password, result.UserPassword) {
		result.UserPassword = "*"
		return result, nil
	} else {
		return user, nil
	}
}
