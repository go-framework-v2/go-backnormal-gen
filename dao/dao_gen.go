package dao

import (
	"context"
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-framework-v2/go-backnormal-gen/tool"
	_ "github.com/go-sql-driver/mysql" // 或其他数据库驱动
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed dao.tpl
var daoTplFS embed.FS // 嵌入整个目录或单个文件

//go:embed dao_mongo.tpl
var daoTplFS_mongo embed.FS // 嵌入整个目录或单个文件

func GenDao_Mysql(dsn string, tables []string, daoDir string, boPath string) error {
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
			BoPath:    boPath,
		}

		// 查询表字段信息
		model.Fields, err = tool.GetTableFields(db, table)
		if err != nil {
			return err
		}

		// 从 embed.FS 读取模板内容
		tplContent, err := daoTplFS.ReadFile("dao.tpl") // 路径相对于 //go:embed
		if err != nil {
			return fmt.Errorf("failed to read embedded template: %v", err)
		}

		// 4. 生成 Dao 文件
		// 获取本文件所在目录
		if err := tool.GenerateFromBytes_dao(model, tplContent, daoDir); err != nil {
			return err
		}
	}

	return nil
}

func GenDao_MongoDB_WithConfig(host string, port int, database, username, password string, tables []string, daoDir string, boPath string) error {
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

	// 获取数据库实例
	db := client.Database(database)
	return GenDao_MongoDB_WithDB(db, tables, daoDir, boPath)
}

func GenDao_MongoDB_WithDB(db *mongo.Database, tables []string, daoDir string, boPath string) error {
	for _, table := range tables {
		// 直接使用集合，不需要分析结构
		collection := db.Collection(table)

		// 获取一个样本文档来分析字段结构
		var sampleDoc bson.M
		err := collection.FindOne(context.Background(), bson.M{}).Decode(&sampleDoc)
		if err != nil && err != mongo.ErrNoDocuments {
			return fmt.Errorf("failed to analyze collection %s: %w", table, err)
		}

		// 构建模型信息
		modelName := tool.ToCamelCase(table)
		daoModel := tool.Model{
			ModelName: modelName,
			Tablename: table,
			BoPath:    boPath,
			Fields:    tool.InferFieldsFromMongoSample(sampleDoc),
		}

		// 使用 MongoDB 专用模板
		tplContent, err := daoTplFS_mongo.ReadFile("dao_mongo.tpl")
		if err != nil {
			return fmt.Errorf("failed to read embedded template: %v", err)
		}

		if err := tool.GenerateFromBytes_dao(daoModel, tplContent, daoDir); err != nil {
			return fmt.Errorf("failed to generate DAO file for %s: %w", table, err)
		}
	}
	return nil
}
