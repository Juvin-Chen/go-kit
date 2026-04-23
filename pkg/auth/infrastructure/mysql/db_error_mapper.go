package mysql

import (
	"errors"

	gomysql "github.com/go-sql-driver/mysql"
)

// isDuplicateKeyError 判断是否为 MySQL 唯一键冲突
// 识别规则：
// 1) 先把通用 error 解包为 *mysql.MySQLError
// 2) 再判断错误码是否为 1062（Duplicate entry）
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlError *gomysql.MySQLError
	return errors.As(err, &mysqlError) && mysqlError.Number == 1062
}
