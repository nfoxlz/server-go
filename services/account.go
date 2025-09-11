// account
package services

import (
	"encoding/base64"
	"server/models"
	"server/repositories"
	"server/util"
	"server/viewmodels"
)

type AccountService struct {
	BusinessService
}

func (s *AccountService) AuthenticateAndGetUser(model viewmodels.LoginViewModel) (models.User, error) {
	var user models.User

	// 使用标准Base64解码
	decodedBytes, err := base64.StdEncoding.DecodeString(model.Password)
	if err != nil {
		return user, err
	}

	ciphertext, err := util.RSADecrypt(decodedBytes)
	util.LogDebug(err)
	if err != nil {
		return user, err
	}

	// util.LogDebug("USER_PASSWORD", string(ciphertext))

	var repository repositories.TenantRepository
	tenant, err := repository.GetTenantByCode(model.Tenant)
	if err != nil {
		return user, err
	}

	result, err := repository.GetUser(tenant, model.User)
	if err != nil {
		return user, err
	}

	if util.Verify(string(ciphertext), result.UserPassword) {
		result.UserPassword = "*"
		return result, nil
	} else {
		return user, nil
	}
}

// func (s *AccountService) GetPublicKey() string {
// 	return []byte{0x41, 0x42, 0x45}
// }
