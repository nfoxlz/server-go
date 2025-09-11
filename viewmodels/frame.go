// frame
package viewmodels

type PeriodYearMonthParameter struct {
	ViewModelBase
	PeriodYearMonth int `json:"periodYearMonth"`
}

type ModifyPasswordParameter struct {
	ViewModelBase
	OriginalPassword string `json:"originalPassword"`
	NewPassword      string `json:"nNewPassword"`
}

type ViewModelBase struct {
	Timestamp    string `json:"timestamp"`
	SignPassword string `json:"Sign - password"`
	// Sign string `json:"sign"`
}
