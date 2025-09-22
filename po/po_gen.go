package po

import (
	"github.com/go-framework-v2/go-backnormal-gen/tool"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func GenPo_Mysql(dsn string, tables []string, poDir string) error {
	// 1. 配置数据库连接
	db, _ := gorm.Open(mysql.Open(dsn))

	// 2. 配置生成器
	g := gen.NewGenerator(gen.Config{
		OutPath:      poDir, // ✅ 确保路径正确（相对路径或绝对路径）
		ModelPkgPath: poDir, // ✅ 指定模型包路径（避免默认 model）
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	// 自定义文件名（驼峰命名）
	g.WithFileNameStrategy(func(tableName string) string {
		fileName := tool.ToCamelCase2(tableName)
		return fileName
	})

	// 自定义模型名（驼峰命名，首字母大写）(默认)

	// 禁用 JSON 标签（返回空字符串）
	g.WithJSONTagNameStrategy(func(columnName string) string {
		return "" // 返回空字符串，表示不生成 JSON 标签
	})

	// 3. 设置目标数据库
	g.UseDB(db)

	// 4. 生成模型（自动从数据库读取表结构）
	// g.GenerateAllTable()
	for _, table := range tables {
		g.GenerateModel(table)
	}

	// 5. 执行生成
	g.Execute()

	return nil
}
