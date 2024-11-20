package entities

type User struct {
	ID      uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name    string  `gorm:"unique;not null" json:"name"`
	Surname string  `gorm:"not null" json:"surname"`
	Age     int     `json:"age"`
	Credit  float64 `json:"credit" gorm:"default:0"`
}
