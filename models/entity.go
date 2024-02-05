// entity
package models

type Entity struct {
	Id   int64
	Code string
	Name string
}

type Tenant struct {
	Entity
	DbServerName         string
	ReadOnlyDbServerName string
}

type User struct {
	Entity
	UserPassword string
	Tenant       Tenant
	Role         Entity
}
