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

// // 自动事务方式
// err := dao.WithTransaction(context.Background(), func(txSession dao.TransactionSession) error {
//     // 在事务中插入数据
//     user1 := &bo.MycollectionBo{Name: "User1"}
//     if _, err := txSession.Insert(user1); err != nil {
//         return err // 自动回滚
//     }

//     // 在事务中查询数据
//     foundUser, err := txSession.FindByID(user1.ID)
//     if err != nil {
//         return err
//     }

//     // 在事务中更新数据
//     update := bson.M{"$set": bson.M{"name": "UpdatedUser"}}
//     if _, err := txSession.UpdateByID(user1.ID, update); err != nil {
//         return err
//     }

//     return nil // 自动提交
// })
