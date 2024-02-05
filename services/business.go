// business
package services

import (
	"server/components"
	"server/repositories"
)

type BusinessService struct {
	components.BusinessComponent
	repository repositories.DataRepository
}
