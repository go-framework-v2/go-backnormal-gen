package gen

import (
	"github.com/go-framework-v2/go-backnormal-gen/bo"
	"github.com/go-framework-v2/go-backnormal-gen/dao"
	"github.com/go-framework-v2/go-backnormal-gen/po"
)

func GenPoBoDao(dsn string, tableList []string,
	poDir, boDir, poPath, daoDir, boPath string) error {
	// 1. po
	err := po.GenPo_Mysql(dsn, tableList, poDir)
	if err != nil {
		return err
	}

	// 2. bo
	err = bo.GenBo_Mysql(dsn, tableList, boDir, poPath)
	if err != nil {
		return err
	}

	// 3. dao
	err = dao.GenDao_Mysql(dsn, tableList, daoDir, boPath)
	if err != nil {
		return err
	}

	return nil
}
