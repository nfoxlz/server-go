// login
package viewmodels

type LoginViewModel struct {
	ViewModelBase
	Tenant   string `json:"tenant"`
	User     string `json:"user"`
	Password string `json:"password"`
}
