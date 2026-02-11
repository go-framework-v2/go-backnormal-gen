package gen

import (
	"path/filepath"

	"github.com/go-framework-v2/go-backnormal-gen/bo"
	"github.com/go-framework-v2/go-backnormal-gen/dao"
	"github.com/go-framework-v2/go-backnormal-gen/ddd"
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

// GenDdd 生成 DDD 目录下的 model（PO）与 repository 实现
// dddDir: 根目录，其下将创建 model 与 repository 子目录，如 "ddd"
// tablePrefix: 表名前缀，如 "ddd_" 则表 book 对应 "ddd_book"
// domainPath: 领域包引用路径，如 "your-module/src/internal/domain/book"
// modelPath: model 包引用路径，如 "your-module/src/internal/infrastructure/persistence/mysql/model"
func GenDdd(dsn string, tableList []string, dddDir string, tablePrefix string, domainPath string, modelPath string) error {
	modelDir := filepath.Join(dddDir, "model")
	if err := ddd.GenModel_Mysql(dsn, tableList, modelDir, tablePrefix); err != nil {
		return err
	}
	repoDir := filepath.Join(dddDir, "repository")
	if err := ddd.GenRepository_Mysql(dsn, tableList, repoDir, domainPath, modelPath); err != nil {
		return err
	}
	return nil
}
