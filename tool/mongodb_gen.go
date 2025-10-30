package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoField MongoDB 字段信息
type MongoField struct {
	Name      string
	Type      string
	BSONTag   string
	OmitEmpty bool
}

// MongoModel MongoDB 模型信息
type MongoModel struct {
	ModelName      string
	TableName      string
	CollectionName string
	Fields         []MongoField
}

// AnalyzeMongoCollection 分析 MongoDB 集合结构
func AnalyzeMongoCollection(client *mongo.Client, database, collectionName string) (*MongoModel, error) {
	collection := client.Database(database).Collection(collectionName)

	// 获取样本文档来分析结构
	var sampleDoc bson.M
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&sampleDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	model := &MongoModel{
		ModelName:      ToCamelCase(collectionName),
		TableName:      collectionName,
		CollectionName: collectionName,
	}

	// 从样本文档推断字段结构
	fields := InferFieldsFromSample(sampleDoc)
	model.Fields = fields

	return model, nil
}

// InferFieldsFromSample 从样本文档推断字段类型
func InferFieldsFromSample(doc bson.M) []MongoField {
	var fields []MongoField

	// 总是包含 ID 字段
	fields = append(fields, MongoField{
		Name:      "ID",
		Type:      "primitive.ObjectID",
		BSONTag:   "_id,omitempty",
		OmitEmpty: true,
	})

	for key, value := range doc {
		// 跳过 _id 字段，因为我们已经手动添加了
		if key == "_id" {
			continue
		}

		field := MongoField{
			Name:    ToCamelCase(key),
			BSONTag: key,
		}

		// 根据值的类型推断 Go 类型
		switch value.(type) {
		case string:
			field.Type = "string"
		case int, int32:
			field.Type = "int32"
		case int64:
			field.Type = "int64"
		case float32, float64:
			field.Type = "float64"
		case bool:
			field.Type = "bool"
		case primitive.DateTime:
			field.Type = "time.Time"
		case primitive.ObjectID:
			field.Type = "primitive.ObjectID"
		case []interface{}:
			field.Type = "[]interface{}"
		case map[string]interface{}, bson.M:
			field.Type = "bson.M"
		default:
			field.Type = "interface{}"
		}

		fields = append(fields, field)
	}

	// 添加常用的时间字段
	commonTimeFields := []string{"created_at", "updated_at", "deleted_at"}
	for _, timeField := range commonTimeFields {
		if _, exists := doc[timeField]; !exists {
			// 如果样本中没有这些字段，但通常会有，可以手动添加
			fields = append(fields, MongoField{
				Name:      ToCamelCase(timeField),
				Type:      "time.Time",
				BSONTag:   timeField,
				OmitEmpty: true,
			})
		}
	}

	return fields
}

// GenerateMongoPOFile 生成 MongoDB PO 文件
func GenerateMongoPOFile(model *MongoModel, poDir string) error {
	// 确保目录存在
	if err := os.MkdirAll(poDir, 0755); err != nil {
		return err
	}

	fileName := ToCamelCase2(model.TableName) + ".go"
	filePath := filepath.Join(poDir, fileName)

	content := generatePOContent(model)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// generatePOContent 生成 PO 文件内容
func generatePOContent(model *MongoModel) string {
	var fieldsBuilder strings.Builder

	for _, field := range model.Fields {
		bsonTag := field.BSONTag
		if field.OmitEmpty {
			bsonTag += ",omitempty"
		}

		fieldsBuilder.WriteString(fmt.Sprintf("\t%s %s `bson:\"%s\"`\n",
			field.Name, field.Type, bsonTag))
	}

	return fmt.Sprintf(`package po

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionName%s = "%s"

// %s MongoDB 文档结构
type %s struct {
%s}

// CollectionName 返回集合名称
func (*%s) CollectionName() string {
	return CollectionName%s
}
`, model.ModelName, model.CollectionName,
		model.ModelName, model.ModelName,
		fieldsBuilder.String(),
		model.ModelName, model.ModelName)
}

// CreateDefaultMongoModel 创建默认的 MongoDB 模型
func CreateDefaultMongoModel(tableName string) *MongoModel {
	// 确保模型名首字母大写
	modelName := strings.Title(ToCamelCase2(tableName))

	return &MongoModel{
		ModelName:      modelName,
		TableName:      tableName,
		CollectionName: tableName,
		Fields: []MongoField{
			{Name: "ID", Type: "primitive.ObjectID", BSONTag: "_id,omitempty", OmitEmpty: true},
			{Name: "Name", Type: "string", BSONTag: "name"},
			{Name: "CreatedAt", Type: "time.Time", BSONTag: "created_at"},
			{Name: "UpdatedAt", Type: "time.Time", BSONTag: "updated_at"},
		},
	}
}

// ConvertMongoFieldsToBOFields 将 MongoDB 字段转换为 BO 字段
func ConvertMongoFieldsToBOFields(mongoFields []MongoField) []Field {
	var boFields []Field

	for _, mongoField := range mongoFields {
		boField := Field{
			Name: mongoField.Name,
			Type: mongoField.Type,
		}
		boFields = append(boFields, boField)
	}

	return boFields
}
