package model

import "time"

type QARecord struct {
    ID         int64     `gorm:"primaryKey" json:"id"`
    Question   string    `gorm:"type:longtext" json:"question"`
    Answer     string    `gorm:"type:longtext" json:"answer"`
    Fragments  string    `gorm:"type:longtext" json:"fragments"`
    CreatedAt  time.Time `json:"created_at"`
}

