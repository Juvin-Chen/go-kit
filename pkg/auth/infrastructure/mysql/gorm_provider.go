package mysql

import (
	"errors"
	"strings"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var ErrMySQLDSNEmpty = errors.New("mysql dsn is empty")

func OpenGormDB(dsn string) (*gorm.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, ErrMySQLDSNEmpty
	}
	return gorm.Open(gormmysql.Open(strings.TrimSpace(dsn)), &gorm.Config{})
}
