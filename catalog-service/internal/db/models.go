package db

import "time"

type Product struct {
	ID 			string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name		string `gorm:"not null"`
	Active 		bool `gorm:"not null;default:true"`
	CreatedAt 	time.Time
}
