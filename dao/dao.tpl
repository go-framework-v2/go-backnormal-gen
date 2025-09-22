package dao

import (
	"{{.BoPath}}"
	"sync"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// 设计要点：
// 接口组合实现扩展
// 结构体嵌入实现代码复用
type (
	// 基础接口 + 扩展方法
	{{.ModelName}}Dao interface {
		// 基础接口，基于单表
        Truncate() error
		FindOne(id int) (*bo.{{.ModelName}}Bo, error) // 根据id查询
		FindOneByUk()                              // 根据唯一键查询
		Insert(obj bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error)
		Update(obj bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error)

		// 扩展接口
	}

	// 自定义模型实现
	custom{{.ModelName}}Dao struct {
		// 默认实现
		db *gorm.DB // 不带事务的基础数据库连接
		tx *gorm.DB // 带事务的数据库连接

		// 扩展实现
		cache *redis.Client // 新增redis缓存
	}
)

// 对外暴露扩展接口
func New{{.ModelName}}Dao(db *gorm.DB, tx *gorm.DB, cache *redis.Client) {{.ModelName}}Dao {
	return &custom{{.ModelName}}Dao{
		db:    db,
		tx:    tx,
		cache: cache,
	}
}

// 基础方法
// Truncate 清空表并重置自增主键
func (d *custom{{.ModelName}}Dao) Truncate() error {
	// 临时禁用外键检查
	if err := d.tx.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return err
	}

	// 1. 清空表数据
	if err := d.tx.Exec("TRUNCATE TABLE " + "{{.Tablename}}").Error; err != nil {
		return err
	}

	// 2. 重置自增主键为1
	if err := d.tx.Exec("ALTER TABLE " + "{{.Tablename}}" + " AUTO_INCREMENT = 1").Error; err != nil {
		return err
	}

	return nil
}

// FindOne 根据id查询
func (d *custom{{.ModelName}}Dao) FindOne(id int) (*bo.{{.ModelName}}Bo, error) {
	var obj bo.{{.ModelName}}Bo

	err := d.db.First(&obj, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &obj, err
}

// FindOneByUk 根据唯一键查询
func (d *custom{{.ModelName}}Dao) FindOneByUk() {

}

// Insert 插入 (事务内)
func (d *custom{{.ModelName}}Dao) Insert(obj bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error) {
	err := d.tx.Create(&obj).Error
	if err != nil {
		return nil, err
	}

	return &obj, err
}

// Update 更新(事务内)
func (d *custom{{.ModelName}}Dao) Update(obj bo.{{.ModelName}}Bo) (*bo.{{.ModelName}}Bo, error) {
	// 只做字段值有值而不是默认值的更新
	err := d.tx.Model(&obj).Updates(obj).Error
	if err != nil {
		return nil, err
	}

	return &obj, err
}

// 扩展方法