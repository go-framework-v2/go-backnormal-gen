package dao

import "testing"

func TestGenDao_Mysql(t *testing.T) {
	// dsn := "root:dev123456@tcp(127.0.0.1:13301)/biz_db?charset=utf8mb4&parseTime=True&loc=Local"
	// tables := []string{"test_person"}
	// daoDir := "/Users/huanlema/Documents/Code/my_code/go-framework_github/go-backnormal-v2/go-backnormal-gen/test/dao"
	// boPath := "github.com/go-framework-v2/go-backnormal-gen/test/model/bo"

	// err := GenDao_Mysql(dsn, tables, daoDir, boPath)
	// if err != nil {
	// 	t.Error(err)
	// }
}

func TestGenDao_MongoDB(t *testing.T) {
	// host := "127.0.0.1"
	// port := 27017
	// database := "test_1029"
	// username := "appuser"
	// password := "app123"
	// tables := []string{"myCollection"}
	// daoDir := "/Users/huanlema/Documents/Code/my_code/github_go-framework-v2/go-backnormal-gen/dao"
	// boPath := "github.com/go-framework-v2/go-backnormal-gen/bo"

	// err := GenDao_MongoDB_WithConfig(host, port, database, username, password, tables, daoDir, boPath)
	// if err != nil {
	// 	t.Error(err)
	// }
}

// // 使用示例
// // 自动事务方式
// err := dao.WithTransaction(context.Background(), func(txSession dao.TransactionSession) error {
//     user1 := &bo.MycollectionBo{Name: "User1"}
//     if _, err := txSession.Insert(user1); err != nil {
//         return err // 自动回滚
//     }

//     user2 := &bo.MycollectionBo{Name: "User2"}
//     if _, err := txSession.Insert(user2); err != nil {
//         return err // 自动回滚
//     }

//     return nil // 自动提交
// })

// // 手动事务方式
// txSession, err := dao.BeginTransaction(context.Background())
// if err != nil {
//     return err
// }
// defer txSession.EndSession(context.Background())

// // 在事务中执行操作
// _, err = txSession.Insert(&user)
// if err != nil {
//     txSession.AbortTransaction(context.Background())
//     return err
// }

// // 提交事务
// err = txSession.CommitTransaction(context.Background())
