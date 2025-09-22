package bo

import (
	"database/sql"
	"embed"
	"fmt"

	_ "embed"

	"github.com/go-framework-v2/go-backnormal-gen/tool"
	_ "github.com/go-sql-driver/mysql" // 或其他数据库驱动
)

//go:embed bo.tpl
var boTplFS embed.FS // 嵌入整个目录或单个文件

func GenBo_Mysql(dsn string, tables []string, boDir string, poPath string) error {
	// 1. 连接数据库
	//dsn: "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// 2. 查询表名
	// 查询所有表名
	// tables, err := getTables(db)
	// if err != nil {
	// 	return err
	// }

	// 指定生成的表名
	// tables := []string{
	// 	// "user",          // 只生成 user 表
	// }

	// 3. 为每个表生成模型
	for _, table := range tables {
		model := tool.Model{
			ModelName: tool.ToCamelCase(table), // 表名转结构体名（如 "user_info" -> "UserInfo"）
			Tablename: table,                   // 表名
			PoPath:    poPath,
		}

		// 查询表字段信息
		model.Fields, err = tool.GetTableFields(db, table)
		if err != nil {
			return err
		}

		// 从 embed.FS 读取模板内容
		tplContent, err := boTplFS.ReadFile("bo.tpl") // 路径相对于 //go:embed
		if err != nil {
			return fmt.Errorf("failed to read embedded template: %v", err)
		}

		// 4. 生成 Bo 文件
		// 获取本文件所在目录
		if err := tool.GenerateFromBytes_bo(model, tplContent, boDir); err != nil {
			return err
		}
	}

	return nil
}
