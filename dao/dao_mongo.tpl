package dao

import (
	"context"

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
	UpdateByID(id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error)
	DeleteByID(id primitive.ObjectID) error
	Count(filter bson.M) (int64, error)

	// 事务支持
	BeginTransaction() (TransactionSession, error)
	WithTransaction(fn func(txSession TransactionSession) error) error
}

// TransactionSession 事务会话接口
type TransactionSession interface {
	CommitTransaction() error
	AbortTransaction() error
	EndSession()
	
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

// 基础 CRUD 方法
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
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (d *custom{{.ModelName}}Dao) Insert(obj *bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	
	if obj.ID.IsZero() {
		obj.ID = primitive.NewObjectID()
	}

	_, err := d.collection.InsertOne(ctx, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
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

func (d *custom{{.ModelName}}Dao) DeleteByID(id primitive.ObjectID) error {
	ctx := context.Background()
	_, err := d.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (d *custom{{.ModelName}}Dao) Count(filter bson.M) (int64, error) {
	ctx := context.Background()
	return d.collection.CountDocuments(ctx, filter)
}

// 事务支持
func (d *custom{{.ModelName}}Dao) BeginTransaction() (TransactionSession, error) {
	ctx := context.Background()
	session, err := d.client.StartSession()
	if err != nil {
		return nil, err
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return nil, err
	}

	return &transactionSessionImpl{
		session:    session,
		collection: d.collection,
	}, nil
}

func (d *custom{{.ModelName}}Dao) WithTransaction(fn func(txSession TransactionSession) error) error {
	txSession, err := d.BeginTransaction()
	if err != nil {
		return err
	}
	defer txSession.EndSession()

	err = fn(txSession)
	if err != nil {
		if abortErr := txSession.AbortTransaction(); abortErr != nil {
			return abortErr
		}
		return err
	}

	return txSession.CommitTransaction()
}

// 事务会话实现
func (ts *transactionSessionImpl) CommitTransaction() error {
	ctx := context.Background()
	return ts.session.CommitTransaction(ctx)
}

func (ts *transactionSessionImpl) AbortTransaction() error {
	ctx := context.Background()
	return ts.session.AbortTransaction(ctx)
}

func (ts *transactionSessionImpl) EndSession() {
	ctx := context.Background()
	ts.session.EndSession(ctx)
}

func (ts *transactionSessionImpl) FindByID(id primitive.ObjectID) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	var result bo.{{.ModelName}}Bo
	err := mongo.WithSession(ctx, ts.session, func(sc mongo.SessionContext) error {
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
	ctx := context.Background()
	var result bo.{{.ModelName}}Bo
	err := mongo.WithSession(ctx, ts.session, func(sc mongo.SessionContext) error {
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
	ctx := context.Background()
	
	if obj.ID.IsZero() {
		obj.ID = primitive.NewObjectID()
	}

	err := mongo.WithSession(ctx, ts.session, func(sc mongo.SessionContext) error {
		_, err := ts.collection.InsertOne(sc, obj)
		return err
	})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (ts *transactionSessionImpl) UpdateByID(id primitive.ObjectID, update bson.M) (*bo.{{.ModelName}}Bo, error) {
	ctx := context.Background()
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var result bo.{{.ModelName}}Bo
	err := mongo.WithSession(ctx, ts.session, func(sc mongo.SessionContext) error {
		return ts.collection.FindOneAndUpdate(sc, bson.M{"_id": id}, update, opts).Decode(&result)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (ts *transactionSessionImpl) DeleteByID(id primitive.ObjectID) error {
	ctx := context.Background()
	return mongo.WithSession(ctx, ts.session, func(sc mongo.SessionContext) error {
		_, err := ts.collection.DeleteOne(sc, bson.M{"_id": id})
		return err
	})
}

func (ts *transactionSessionImpl) Count(filter bson.M) (int64, error) {
	ctx := context.Background()
	var count int64
	err := mongo.WithSession(ctx, ts.session, func(sc mongo.SessionContext) error {
		var err error
		count, err = ts.collection.CountDocuments(sc, filter)
		return err
	})
	return count, err
}