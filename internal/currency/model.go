package currency

type Currency struct {
	Code          string `gorm:"column:code;primaryKey"`
	Name          string `gorm:"column:name"`
	DecimalPlaces int16  `gorm:"column:decimal_places"`
}

func (Currency) TableName() string {
	return "currencies"
}
