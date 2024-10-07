package models

type User struct {
	Model
	Username string    `gorm:"unique;not null;index:idx_username" json:"username" swagger:"example=exampleuser"`
	Email    string    `gorm:"unique;not null;index:idx_email" json:"email" swagger:"example=user@example.com"`
	Password string    `gorm:"not null" json:"-"` // Password should not be returned in JSON
	ApiKey   string    `gorm:"unique;not null;index:idx_api_key" json:"api_key" swagger:"example=your_api_key"`
	Monitors []Monitor `gorm:"foreignKey:UserID" json:"monitors"`
}
