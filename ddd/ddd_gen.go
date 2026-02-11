package ddd

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-framework-v2/go-backnormal-gen/tool"
	_ "github.com/go-sql-driver/mysql"
)

//go:embed model_po.tpl
var modelPoTpl []byte

//go:embed repository_impl.tpl
var repositoryImplTpl []byte

// ModelDdd 用于生成 DDD model 的模板数据
type ModelDdd struct {
	ModelName string
	TableName string
	Fields    []tool.FieldGorm
}

// RepoDdd 用于生成 DDD repository 的模板数据
type RepoDdd struct {
	ModelName      string
	DomainPath     string
	DomainPkg      string
	ModelPath      string
	ModelPkg       string // model 包名，用于 model.XXXPO
	IdColumn       string
	IdParamType    string
	IdQueryExpr    string
	ModelNameLower string
	ToDomainBody   string
	ToPOBody       string
}

// GenModel_Mysql 根据表生成 DDD model 目录下的 PO 文件
// dddModelDir: 输出目录，如 "ddd/model"
// tablePrefix: 表名前缀，如 "ddd_" 则表 book 对应 "ddd_book"
func GenModel_Mysql(dsn string, tables []string, dddModelDir string, tablePrefix string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, table := range tables {
		fields, err := tool.GetTableFieldsForGorm(db, table)
		if err != nil {
			return fmt.Errorf("get table fields for %s: %w", table, err)
		}
		tableName := table
		if tablePrefix != "" {
			tableName = tablePrefix + table
		}
		modelName := tool.ToCamelCase(table)
		model := ModelDdd{
			ModelName: modelName,
			TableName: tableName,
			Fields:    fields,
		}
		if err := generateFromTpl(modelPoTpl, model, dddModelDir, tool.ToCamelCase2(modelName)+"_po.go"); err != nil {
			return err
		}
	}
	return nil
}

// GenRepository_Mysql 根据表生成 DDD repository 目录下的仓储实现
// dddRepoDir: 输出目录，如 "ddd/repository"
// domainPath: 领域包引用路径，如 "library-system-ddd/src/internal/domain/book"
// modelPath: model 包引用路径，如 "library-system-ddd/src/internal/infrastructure/persistence/mysql/model"
func GenRepository_Mysql(dsn string, tables []string, dddRepoDir string, domainPath string, modelPath string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

		domainPkg := filepath.Base(domainPath)
		if domainPkg == "." || domainPkg == "/" {
			domainPkg = "domain"
		}
		modelPkg := filepath.Base(modelPath)
		if modelPkg == "." || modelPkg == "/" {
			modelPkg = "model"
		}

	for _, table := range tables {
		fields, err := tool.GetTableFieldsForGorm(db, table)
		if err != nil {
			return fmt.Errorf("get table fields for %s: %w", table, err)
		}
		modelName := tool.ToCamelCase(table)
		modelNameLower := tool.ToCamelCase2(modelName)
		idColumn := "id"
		for _, f := range fields {
			if f.GoName == "ID" || strings.ToLower(f.ColumnName) == "id" {
				idColumn = f.ColumnName
				break
			}
		}
		// 领域层使用值对象，如 book.BookID，查询用 id.Value()
		idParamType := domainPkg + "." + modelName + "ID"
		idQueryExpr := "id.Value()"
		toDomainBody := buildToDomainBody(domainPkg, modelName, fields)
		toPOBody := buildToPOBody(fields)

		repo := RepoDdd{
			ModelName:      modelName,
			DomainPath:     domainPath,
			DomainPkg:      domainPkg,
			ModelPath:      modelPath,
			ModelPkg:       modelPkg,
			IdColumn:       idColumn,
			IdParamType:    idParamType,
			IdQueryExpr:    idQueryExpr,
			ModelNameLower: modelNameLower,
			ToDomainBody:   toDomainBody,
			ToPOBody:       toPOBody,
		}
		outFile := modelNameLower + "_repository_impl.go"
		if err := generateFromTpl(repositoryImplTpl, repo, dddRepoDir, outFile); err != nil {
			return err
		}
	}
	return nil
}

func buildToDomainBody(domainPkg, modelName string, fields []tool.FieldGorm) string {
	if len(fields) == 0 {
		return "\treturn nil // TODO: implement toDomain"
	}
	idField := "ID"
	for _, f := range fields {
		if f.GoName == "ID" {
			idField = f.GoName
			break
		}
	}
	var sb strings.Builder
	sb.WriteString("\tid, _ := ")
	sb.WriteString(domainPkg)
	sb.WriteString(".New")
	sb.WriteString(modelName)
	sb.WriteString("ID(po.")
	sb.WriteString(idField)
	sb.WriteString(")\n\treturn ")
	sb.WriteString(domainPkg)
	sb.WriteString(".Restore")
	sb.WriteString(modelName)
	sb.WriteString("(id")
	for _, f := range fields {
		if f.GoName == "ID" || strings.ToLower(f.ColumnName) == "deleted_at" {
			continue
		}
		sb.WriteString(", po.")
		sb.WriteString(f.GoName)
	}
	sb.WriteString(")")
	return sb.String()
}

func buildToPOBody(fields []tool.FieldGorm) string {
	var sb strings.Builder
	for _, f := range fields {
		if strings.ToLower(f.ColumnName) == "deleted_at" {
			continue
		}
		sb.WriteString("\t\t")
		sb.WriteString(f.GoName)
		sb.WriteString(": ")
		if f.GoName == "ID" {
			sb.WriteString("b.ID().Value()")
		} else {
			sb.WriteString("b.")
			sb.WriteString(f.GoName)
			sb.WriteString("()")
		}
		sb.WriteString(",\n")
	}
	return sb.String()
}

func generateFromTpl(tplContent []byte, data interface{}, outputDir string, filename string) error {
	tpl, err := template.New("ddd").Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
