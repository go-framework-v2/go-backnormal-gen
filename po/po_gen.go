package po

import (
	"context"
	"fmt"
	"time"

	"github.com/go-framework-v2/go-backnormal-gen/tool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GenPo_MongoDB(mongoURI, database string, tables []string, poDir string) error {
	// 1. 连接 MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(context.TODO())

	// 2. 分析集合结构
	for _, table := range tables {
		model, err := tool.AnalyzeMongoCollection(client, database, table)
		if err != nil {
			return fmt.Errorf("failed to analyze collection %s: %w", table, err)
		}

		// 3. 生成 PO 文件
		if err := tool.GenerateMongoPOFile(model, poDir); err != nil {
			return fmt.Errorf("failed to generate PO file for %s: %w", table, err)
		}
	}

	return nil
}

// GenPo_MongoDB_WithConfig 使用配置参数连接 MongoDB
func GenPo_MongoDB_WithConfig(host string, port int, database, username, password string, tables []string, poDir string) error {
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

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 分析集合结构并生成 PO 文件
	for _, table := range tables {
		model, err := tool.AnalyzeMongoCollection(client, database, table)
		if err != nil {
			return fmt.Errorf("failed to analyze collection %s: %w", table, err)
		}

		if err := tool.GenerateMongoPOFile(model, poDir); err != nil {
			return fmt.Errorf("failed to generate PO file for %s: %w", table, err)
		}
	}

	return nil
}
