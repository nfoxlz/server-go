// data
package models

type SimpleData struct {
	Columns []string
	Rows    [][]any
}

// type SimpleDataTable struct {
// 	TableName string
// 	SimpleData
// }

type PeriodType int

const (
	None PeriodType = iota
	Year
	Quarter
	Month
	TenDays
	Week
	Day
	Hour
	QuarterHour
	Minute
)

var periodTypeMap = map[PeriodType]string{
	None:        "None",
	Year:        "Year",
	Quarter:     "Quarter",
	Month:       "Month",
	TenDays:     "TenDays",
	Week:        "Week",
	Day:         "Day",
	Hour:        "Hour",
	QuarterHour: "QuarterHour",
	Minute:      "Minute",
}

func (t PeriodType) String() string {
	if v, ok := periodTypeMap[t]; ok {
		return v
	}

	return "invalid value"
}

type SequenceInfo struct {
	No         int64
	Name       string
	PeriodType PeriodType
	Format     string
}
