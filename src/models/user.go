package models

type User struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Age      int     `json:"age"`
	Email    string  `json:"email" gorm:"unique"`
	Password string  `json:"-" gorm:"column:Password_Hash"`
	Role     string  `json:"role" gorm:"default:user"`
	Credit   float64 `json:"credit" gorm:"default:0"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
