package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type RoleList []string

const RoleAdmin = "ADMIN"
const RoleUser = "USER"

func (roleList *RoleList) Scan(src any) error {
	roleString, ok := src.(string)
	if !ok {
		return fmt.Errorf("src value %v cannot cast to []byte", src)
	}
	*roleList = strings.Split(roleString, ",")
	return nil
}

func (roleList RoleList) Value() (driver.Value, error) {
	if len(roleList) == 0 {
		return nil, nil
	}
	return strings.Join(roleList, ","), nil
}
