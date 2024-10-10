package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type AlertTypes []AlertType

type AlertType string

const (
	AlertSMTP     AlertType = "smtp"
	AlertTelegram AlertType = "telegram"
	AlertNone     AlertType = "none"
)

func (a AlertTypes) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *AlertTypes) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal AlertTypes value")
	}

	return json.Unmarshal(bytes, a)
}
