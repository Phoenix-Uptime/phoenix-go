package models

type Tag struct {
	Model
	Name        string    `gorm:"unique;not null" json:"name" swagger:"example=important"`
	Description string    `json:"description,omitempty" swagger:"example=Critical tag"`
	Color       string    `json:"color,omitempty" swagger:"example=#FF0000"`
	Monitors    []Monitor `gorm:"many2many:monitor_tags;" json:"monitors"`
}
