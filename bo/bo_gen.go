package bo

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "embed"

	"github.com/go-framework-v2/go-backnormal-gen/tool"
	_ "github.com/go-sql-driver/mysql" // 或其他数据库驱动
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed bo.tpl
var boTplFS embed.FS // 嵌入整个目录或单个文件

//go:embed bo_mongo.tpl
var boTplFS_mongo embed.FS // 嵌入整个目录或单个文件

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

func GenBo_MongoDB_WithConfig(host string, port int, database, username, password string, tables []string, boDir string, poPath string) error {
	// 1. 连接 MongoDB
	// 构建连接URI（使用您已验证的格式）
	dsn := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", username, password, host, port, database)

	// 设置客户端选项
	clientOptions := options.Client().
		ApplyURI(dsn).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(100).
		SetMinPoolSize(5)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 建立连接
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// 验证连接 - 使用相同的上下文
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 2. 为每个表生成 BO 模型
	for _, table := range tables {
		// 分析集合结构
		model, err := tool.AnalyzeMongoCollection(client, database, table)
		if err != nil {
			// 如果分析失败，使用默认结构
			fmt.Printf("Warning: failed to analyze collection %s, using default structure: %v\n", table, err)
			model = tool.CreateDefaultMongoModel(table)
		}

		// 转换为 BO 模型
		boModel := tool.Model{
			ModelName: model.ModelName,
			Tablename: model.TableName,
			PoPath:    poPath,
			Fields:    tool.ConvertMongoFieldsToBOFields(model.Fields),
		}

		// 从 embed.FS 读取模板内容
		tplContent, err := boTplFS_mongo.ReadFile("bo_mongo.tpl")
		if err != nil {
			return fmt.Errorf("failed to read embedded template: %v", err)
		}

		// 3. 生成 BO 文件
		if err := tool.GenerateFromBytes_bo(boModel, tplContent, boDir); err != nil {
			return fmt.Errorf("failed to generate BO file for %s: %w", table, err)
		}

		fmt.Printf("✅ Successfully generated BO file for: %s\n", table)
	}

	return nil
}
