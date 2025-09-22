package tool

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

type Model struct {
	ModelName string  // 结构体名称（如 "User"）
	Tablename string  // 表名（如 "user"）
	BoPath    string  // 实际的业务对象包路径
	PoPath    string  // 实际的持久化对象包路径
	Fields    []Field // 字段列表
}

type Field struct {
	Name string // 字段名（如 "Username"）
	Type string // Go 类型（如 "string"）
}

// 获取数据库所有表名（MySQL 示例）
func GetTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// 获取表的字段信息（MySQL 示例）
func GetTableFields(db *sql.DB, table string) ([]Field, error) {
	query := `
        SELECT column_name, data_type, is_nullable 
        FROM information_schema.columns 
        WHERE table_schema = DATABASE() AND table_name = ?
    `
	rows, err := db.Query(query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []Field
	for rows.Next() {
		var name, sqlType, isNullable string
		if err := rows.Scan(&name, &sqlType, &isNullable); err != nil {
			return nil, err
		}
		fields = append(fields, Field{
			Name: ToCamelCase(name),
			Type: sqlTypeToGoType(sqlType),
		})
	}
	return fields, nil
}

// Generate
func Generate(model Model, tplPath string, daoDir string) error {
	// 1. 获取模板路径
	_ = tplPath

	// 2. 检查模板文件是否存在
	if _, err := os.Stat(tplPath); os.IsNotExist(err) {
		return err
	}

	// 3. 解析模板
	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		return err
	}

	// 4. 获取 生成dao文件目录
	// 5. 确保 dao/ 目录存在
	if err = os.MkdirAll(daoDir, 0755); err != nil {
		return err
	}

	// 6. 生成目标文件路径（如 dao/biz_userDao.go）bizUserDao.gen.go
	filename := fmt.Sprintf("%sDao.gen.go", ToCamelCase2(model.ModelName))
	outputPath := filepath.Join(daoDir, filename)

	// 7. 创建并写入文件
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 8. 执行模板渲染
	return tpl.Execute(file, model)
}

// GenerateFromBytes_dao 从字节生成文件（替代原来的从文件路径生成）
func GenerateFromBytes_dao(model Model, tplContent []byte, outputDir string) error {
	// 解析模板
	tpl, err := template.New("bo").Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// 渲染模板到缓冲区
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, model); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 写入文件
	// 文件名处理：BizConfins.gen.go -> bizConfinsDao.gen.go
	// 首字母小写，其余驼峰。表名后加Dao
	modelName := ToCamelCase2(model.ModelName) + "Dao"
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.gen.go", modelName))
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// GenerateFromBytes_bo 从字节生成文件（替代原来的从文件路径生成）
func GenerateFromBytes_bo(model Model, tplContent []byte, outputDir string) error {
	// 解析模板
	tpl, err := template.New("bo").Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// 渲染模板到缓冲区
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, model); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 写入文件
	// 文件名处理：BizConfins.gen.go -> bizConfinsBo.gen.go
	// 首字母小写，其余驼峰。表名后加Dao
	modelName := ToCamelCase2(model.ModelName) + "Bo"
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.gen.go", modelName))
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// 获取模板路径 同目录下的bo.tpl
func getTemplatePath(name string) (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller info")
	}

	dir := filepath.Dir(filename)
	tplPath := filepath.Join(dir, name)

	return tplPath, nil
}

// 辅助函数：下划线转驼峰（如 "user_name" -> "UserName"）
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(strings.ToLower(parts[i]))
	}
	return strings.Join(parts, "")
}

// biz_USer -> bizUser
func ToCamelCase2(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		// 手动首字母大写，其余小写
		runes := []rune(parts[i])
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	// 拼接成驼峰命名（首字母小写）
	result := strings.Join(parts, "")
	if len(result) > 0 {
		runes := []rune(result)
		runes[0] = unicode.ToLower(runes[0])
		return string(runes)
	}
	return result
}

// 辅助函数：SQL 类型转 Go 类型，与gorm生成工具的类型映射保持一致
func sqlTypeToGoType(sqlType string) string {
	sqlType = strings.ToLower(sqlType)
	switch {
	// int 类型
	case strings.Contains(sqlType, "int"):
		return "in32"
	case strings.Contains(sqlType, "bigint"):
		return "int64"
	case strings.Contains(sqlType, "tinyint"):
		return "uint8"
	// string 类型
	case strings.Contains(sqlType, "varchar"),
		strings.Contains(sqlType, "char"),
		strings.Contains(sqlType, "text"),
		strings.Contains(sqlType, "longtext"):
		return "string"
	// float 类型
	case strings.Contains(sqlType, "decimal"),
		strings.Contains(sqlType, "float"),
		strings.Contains(sqlType, "double"):
		return "float64"
	// time 类型
	case strings.Contains(sqlType, "timestamp"),
		strings.Contains(sqlType, "datetime"),
		strings.Contains(sqlType, "date"):
		return "time.Time"
	// bool 类型
	case strings.Contains(sqlType, "bool"):
		return "bool"
	default:
		return "interface{}" // 未知类型
	}
}
