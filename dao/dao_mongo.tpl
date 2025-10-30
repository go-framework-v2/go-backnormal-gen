package dao

import (
	"context"
	"time"

	"{{.BoPath}}"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// {{.ModelName}}Dao MongoDB 数据访问接口
type {{.ModelName}}Dao interface {
	// 基础 CRUD 操作
	FindByID(ctx context.Context, id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error)
	FindOne(ctx context.Context, filter bson.M) (*bo.{{.ModelName}}Bo, error)
	Find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*bo.{{.ModelName}}Bo, error)
	Insert(ctx context.Context, obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error)
	InsertMany(ctx context.Context, objs []*bo.{{.ModelName}}Bo) ([]interface{}, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error)
	UpdateOne(ctx context.Context, filter bson.M, update bson.M) (*bo.{{.ModelName}}Bo, error)
	UpdateMany(ctx context.Context, filter bson.M, update bson.M) (int64, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	DeleteOne(ctx context.Context, filter bson.M) error
	DeleteMany(ctx context.Context, filter bson.M) (int64, error)
	Count(ctx context.Context, filter bson.M) (int64, error)

	// 分页查询
	FindWithPagination(ctx context.Context, filter bson.M, page, pageSize int64, sort bson.M) ([]*bo.{{.ModelName}}Bo, int64, error)

	// 聚合操作
	Aggregate(ctx context.Context, pipeline mongo.Pipeline) ([]bson.M, error)

	// 索引管理
	CreateIndexes(ctx context.Context, indexes []mongo.IndexModel) ([]string, error)
}

// custom{{.ModelName}}Dao 自定义 DAO 实现
type custom{{.ModelName}}Dao struct {
	collection *mongo.Collection
	database   *mongo.Database
	client     *mongo.Client
}

// New{{.ModelName}}Dao 创建新的 DAO 实例
func New{{.ModelName}}Dao(client *mongo.Client, databaseName string) {{.ModelName}}Dao {
	database := client.Database(databaseName)
	collection := database.Collection("{{.Tablename}}")
	
	return &custom{{.ModelName}}Dao{
		collection: collection,
		database:   database,
		client:     client,
	}
}

// FindByID 根据 ID 查询文档
func (d *custom{{.ModelName}}Dao) FindByID(ctx context.Context, id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error) {
	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// FindOne 查询单个文档
func (d *custom{{.ModelName}}Dao) FindOne(ctx context.Context, filter bson.M) (*bo.{{.ModelName}}Bo, error) {
	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// Find 查询多个文档
func (d *custom{{.ModelName}}Dao) Find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*bo.{{.ModelName}}Bo, error) {
	cursor, err := d.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*bo.{{.ModelName}}Bo
	for cursor.Next(ctx) {
		var obj bo.{{.ModelName}}Bo
		if err := cursor.Decode(&obj); err != nil {
			return nil, err
		}
		results = append(results, &obj)
	}

	return results, cursor.Err()
}

// Insert 插入单个文档
func (d *custom{{.ModelName}}Dao) Insert(ctx context.Context, obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error) {
	// 设置创建时间和更新时间
	now := time.Now()
	if obj.CreatedAt.IsZero() {
		obj.CreatedAt = now
	}
	obj.UpdatedAt = now

	// 生成 ObjectID 如果不存在
	if obj.ID.IsZero() {
		obj.ID = primitive.NewObjectID()
	}

	_, err := d.collection.InsertOne(ctx, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// InsertMany 批量插入文档
func (d *custom{{.ModelName}}Dao) InsertMany(ctx context.Context, objs []*bo.{{.ModelName}}Bo) ([]interface{}, error) {
	if len(objs) == 0 {
		return nil, nil
	}

	// 准备插入数据
	documents := make([]interface{}, len(objs))
	now := time.Now()

	for i, obj := range objs {
		// 设置时间戳
		if obj.CreatedAt.IsZero() {
			obj.CreatedAt = now
		}
		obj.UpdatedAt = now

		// 生成 ObjectID 如果不存在
		if obj.ID.IsZero() {
			obj.ID = primitive.NewObjectID()
		}

		documents[i] = obj
	}

	result, err := d.collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, err
	}

	return result.InsertedIDs, nil
}

// UpdateByID 根据 ID 更新文档
func (d *custom{{.ModelName}}Dao) UpdateByID(ctx context.Context, id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error) {
	// 添加更新时间
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateOne 更新单个文档
func (d *custom{{.ModelName}}Dao) UpdateOne(ctx context.Context, filter bson.M, update bson.M) (*bo.{{.ModelName}}Bo, error) {
	// 添加更新时间
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateMany 更新多个文档
func (d *custom{{.ModelName}}Dao) UpdateMany(ctx context.Context, filter bson.M, update bson.M) (int64, error) {
	// 添加更新时间
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()

	result, err := d.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// DeleteByID 根据 ID 删除文档
func (d *custom{{.ModelName}}Dao) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := d.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// DeleteOne 删除单个文档
func (d *custom{{.ModelName}}Dao) DeleteOne(ctx context.Context, filter bson.M) error {
	_, err := d.collection.DeleteOne(ctx, filter)
	return err
}

// DeleteMany 删除多个文档
func (d *custom{{.ModelName}}Dao) DeleteMany(ctx context.Context, filter bson.M) (int64, error) {
	result, err := d.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// Count 统计文档数量
func (d *custom{{.ModelName}}Dao) Count(ctx context.Context, filter bson.M) (int64, error) {
	return d.collection.CountDocuments(ctx, filter)
}

// FindWithPagination 分页查询
func (d *custom{{.ModelName}}Dao) FindWithPagination(ctx context.Context, filter bson.M, page, pageSize int64, sort bson.M) ([]*bo.{{.ModelName}}Bo, int64, error) {
	// 计算跳过数量
	skip := (page - 1) * pageSize

	// 设置查询选项
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(pageSize)
	
	if sort != nil {
		findOptions.SetSort(sort)
	}

	// 查询数据
	results, err := d.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := d.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// Aggregate 聚合查询
func (d *custom{{.ModelName}}Dao) Aggregate(ctx context.Context, pipeline mongo.Pipeline) ([]bson.M, error) {
	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// CreateIndexes 创建索引
func (d *custom{{.ModelName}}Dao) CreateIndexes(ctx context.Context, indexes []mongo.IndexModel) ([]string, error) {
	return d.collection.Indexes().CreateMany(ctx, indexes)
}