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
	FindByID(id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error)
	FindOne(filter bson.M) (*bo.{{.ModelName}}Bo, error)
	Find(filter bson.M, opts ...*options.FindOptions) ([]*bo.{{.ModelName}}Bo, error)
	Insert(obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error)
	InsertMany(objs []*bo.{{.ModelName}}Bo) ([]primitive.ObjectID, error)
	UpdateByID(id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error)
	UpdateOne(filter bson.M, update bson.M) (*bo.{{.ModelName}}Bo, error)
	UpdateMany(filter bson.M, update bson.M) (int64, error)
	DeleteByID(id primitive.ObjectID) error
	DeleteOne(filter bson.M) error
	DeleteMany(filter bson.M) (int64, error)
	Count(filter bson.M) (int64, error)

	// MongoDB特有操作
	FindWithPagination(filter bson.M, page, pageSize int64, sort bson.M) ([]*bo.{{.ModelName}}Bo, int64, error)
	Aggregate(pipeline mongo.Pipeline) ([]bson.M, error)
	BulkWrite(operations []mongo.WriteModel) (*mongo.BulkWriteResult, error)
	CreateIndexes(indexes []mongo.IndexModel) ([]string, error)
	FindOneAndUpdate(filter bson.M, update bson.M, opts ...*options.FindOneAndUpdateOptions) (*bo.{{.ModelName}}Bo, error)
	FindOneAndDelete(filter bson.M, opts ...*options.FindOneAndDeleteOptions) (*bo.{{.ModelName}}Bo, error)

	// 事务支持
	BeginTransaction(ctx context.Context) (TransactionSession, error)
	WithTransaction(ctx context.Context, fn func(txSession TransactionSession) error) error
}

// TransactionSession 事务会话接口
type TransactionSession interface {
	// 事务操作
	CommitTransaction(ctx context.Context) error
	AbortTransaction(ctx context.Context) error
	EndSession(ctx context.Context)
	
	// 在事务中执行的操作
	FindByID(id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error)
	FindOne(filter bson.M) (*bo.{{.ModelName}}Bo, error)
	Insert(obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error)
	UpdateByID(id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error)
	DeleteByID(id primitive.ObjectID) error
	Count(filter bson.M) (int64, error)
}

// transactionSessionImpl 事务会话实现
type transactionSessionImpl struct {
	session    mongo.Session
	collection *mongo.Collection
	ctx        context.Context
}

// custom{{.ModelName}}Dao 自定义 DAO 实现
type custom{{.ModelName}}Dao struct {
	collection *mongo.Collection
	db         *mongo.Database
	client     *mongo.Client
}

// New{{.ModelName}}Dao 创建新的 DAO 实例
func New{{.ModelName}}Dao(db *mongo.Database) {{.ModelName}}Dao {
	collection := db.Collection("{{.Tablename}}")
	client := db.Client()
	return &custom{{.ModelName}}Dao{
		collection: collection,
		db:         db,
		client:     client,
	}
}

// 基础 CRUD 方法（无上下文）
func (d *custom{{.ModelName}}Dao) FindByID(id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
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

func (d *custom{{.ModelName}}Dao) FindOne(filter bson.M) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
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

func (d *custom{{.ModelName}}Dao) Find(filter bson.M, opts ...*options.FindOptions) ([]*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
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

func (d *custom{{.ModelName}}Dao) Insert(obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	
	// 只生成 ObjectID，不添加额外字段
	if obj.ID.IsZero() {
		obj.ID = primitive.NewObjectID()
	}

	_, err := d.collection.InsertOne(ctx, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (d *custom{{.ModelName}}Dao) InsertMany(objs []*bo.{{.ModelName}}Bo) ([]primitive.ObjectID, error) {
	ctx := context.Background()
	if len(objs) == 0 {
		return nil, nil
	}

	documents := make([]interface{}, len(objs))
	ids := make([]primitive.ObjectID, len(objs))

	for i, obj := range objs {
		// 只生成 ObjectID，不添加额外字段
		if obj.ID.IsZero() {
			obj.ID = primitive.NewObjectID()
		}
		ids[i] = obj.ID
		documents[i] = obj
	}

	_, err := d.collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (d *custom{{.ModelName}}Dao) UpdateByID(id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (d *custom{{.ModelName}}Dao) UpdateOne(filter bson.M, update bson.M) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (d *custom{{.ModelName}}Dao) UpdateMany(filter bson.M, update bson.M) (int64, error) {
	ctx := context.Background()
	result, err := d.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (d *custom{{.ModelName}}Dao) DeleteByID(id primitive.ObjectID) error {
	ctx := context.Background()
	_, err := d.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (d *custom{{.ModelName}}Dao) DeleteOne(filter bson.M) error {
	ctx := context.Background()
	_, err := d.collection.DeleteOne(ctx, filter)
	return err
}

func (d *custom{{.ModelName}}Dao) DeleteMany(filter bson.M) (int64, error) {
	ctx := context.Background()
	result, err := d.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (d *custom{{.ModelName}}Dao) Count(filter bson.M) (int64, error) {
	ctx := context.Background()
	return d.collection.CountDocuments(ctx, filter)
}

// MongoDB特有方法
func (d *custom{{.ModelName}}Dao) FindWithPagination(filter bson.M, page, pageSize int64, sort bson.M) ([]*bo.{{.ModelName}}Bo, int64, error) {
	skip := (page - 1) * pageSize
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(pageSize)
	
	if sort != nil {
		findOptions.SetSort(sort)
	}

	results, err := d.Find(filter, findOptions)
	if err != nil {
		return nil, 0, err
	}

	total, err := d.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (d *custom{{.ModelName}}Dao) Aggregate(pipeline mongo.Pipeline) ([]bson.M, error) {
	ctx := context.Background()
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

func (d *custom{{.ModelName}}Dao) BulkWrite(operations []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	ctx := context.Background()
	return d.collection.BulkWrite(ctx, operations)
}

func (d *custom{{.ModelName}}Dao) CreateIndexes(indexes []mongo.IndexModel) ([]string, error) {
	ctx := context.Background()
	return d.collection.Indexes().CreateMany(ctx, indexes)
}

func (d *custom{{.ModelName}}Dao) FindOneAndUpdate(filter bson.M, update bson.M, opts ...*options.FindOneAndUpdateOptions) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOneAndUpdate(ctx, filter, update, opts...).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (d *custom{{.ModelName}}Dao) FindOneAndDelete(filter bson.M, opts ...*options.FindOneAndDeleteOptions) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	var result bo.{{.ModelName}}Bo
	err := d.collection.FindOneAndDelete(ctx, filter, opts...).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// 事务支持
func (d *custom{{.ModelName}}Dao) BeginTransaction(ctx context.Context) (TransactionSession, error) {
	session, err := d.client.StartSession()
	if err != nil {
		return nil, err
	}

	// 开始事务
	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return nil, err
	}

	return &transactionSessionImpl{
		session:    session,
		collection: d.collection,
		ctx:        ctx,
	}, nil
}

func (d *custom{{.ModelName}}Dao) WithTransaction(ctx context.Context, fn func(txSession TransactionSession) error) error {
	txSession, err := d.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer txSession.EndSession(ctx)

	// 执行事务函数
	err = fn(txSession)
	if err != nil {
		// 出错时回滚事务
		if abortErr := txSession.AbortTransaction(ctx); abortErr != nil {
			return abortErr
		}
		return err
	}

	// 提交事务
	return txSession.CommitTransaction(ctx)
}

// 事务会话实现
func (ts *transactionSessionImpl) CommitTransaction(ctx context.Context) error {
	return ts.session.CommitTransaction(ctx)
}

func (ts *transactionSessionImpl) AbortTransaction(ctx context.Context) error {
	return ts.session.AbortTransaction(ctx)
}

func (ts *transactionSessionImpl) EndSession(ctx context.Context) {
	ts.session.EndSession(ctx)
}

func (ts *transactionSessionImpl) FindByID(id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error) {
	var result bo.{{.ModelName}}Bo
	err := mongo.WithSession(ts.ctx, ts.session, func(sc mongo.SessionContext) error {
		return ts.collection.FindOne(sc, bson.M{"_id": id}).Decode(&result)
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (ts *transactionSessionImpl) FindOne(filter bson.M) (*bo.{{.ModelName}}Bo, error) {
	var result bo.{{.ModelName}}Bo
	err := mongo.WithSession(ts.ctx, ts.session, func(sc mongo.SessionContext) error {
		return ts.collection.FindOne(sc, filter).Decode(&result)
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (ts *transactionSessionImpl) Insert(obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error) {
	// 只生成 ObjectID，不添加额外字段
	if obj.ID.IsZero() {
		obj.ID = primitive.NewObjectID()
	}

	err := mongo.WithSession(ts.ctx, ts.session, func(sc mongo.SessionContext) error {
		_, err := ts.collection.InsertOne(sc, obj)
		return err
	})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (ts *transactionSessionImpl) UpdateByID(id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error) {
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var result bo.{{.ModelName}}Bo
	err := mongo.WithSession(ts.ctx, ts.session, func(sc mongo.SessionContext) error {
		return ts.collection.FindOneAndUpdate(sc, bson.M{"_id": id}, update, opts).Decode(&result)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (ts *transactionSessionImpl) DeleteByID(id primitive.ObjectID) error {
	return mongo.WithSession(ts.ctx, ts.session, func(sc mongo.SessionContext) error {
		_, err := ts.collection.DeleteOne(sc, bson.M{"_id": id})
		return err
	})
}

func (ts *transactionSessionImpl) Count(filter bson.M) (int64, error) {
	var count int64
	err := mongo.WithSession(ts.ctx, ts.session, func(sc mongo.SessionContext) error {
		var err error
		count, err = ts.collection.CountDocuments(sc, filter)
		return err
	})
	return count, err
}