package models

import "time"

type Message struct {
	ID     int       `mapstructure:"id"`
	ChatID int       `mapstructure:"chat_id"`
	UserID int       `mapstructure:"user_id"`
	Text   string    `mapstructure:"text"`
	Time   time.Time `mapstructure:"time"`

	Deleted bool `mapstructure:"deleted"`
	DB      bool `mapstructure:"db"`
	Kafka   bool `mapstructure:"kafka"`
}
