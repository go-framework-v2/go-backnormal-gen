package bo

import "testing"

func TestGenBo_Mysql(t *testing.T) {
	dsn := "root:dev123456@tcp(127.0.0.1:13301)/biz_db?charset=utf8mb4&parseTime=True&loc=Local"
	tables := []string{"test_person"}
	boDir := "/Users/huanlema/Documents/Code/my_code/go-framework_github/go-backnormal-v2/go-backnormal-gen/test/model/bo"
	poPath := "github.com/go-framework-v2/go-backnormal-gen/test/dao/po"

	err := GenBo_Mysql(dsn, tables, boDir, poPath)
	if err != nil {
		t.Error(err)
	}
}
