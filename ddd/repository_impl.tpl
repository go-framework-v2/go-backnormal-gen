package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"{{.DomainPath}}"
	"{{.ModelPath}}"
)

// {{.ModelName}}RepositoryImpl {{.ModelName}}仓储实现
type {{.ModelName}}RepositoryImpl struct {
	db *gorm.DB
}

// New{{.ModelName}}Repository 创建{{.ModelName}}仓储
func New{{.ModelName}}Repository(db *gorm.DB) *{{.ModelName}}RepositoryImpl {
	return &{{.ModelName}}RepositoryImpl{db: db}
}

// FindByID 根据ID查找
func (r *{{.ModelName}}RepositoryImpl) FindByID(ctx context.Context, id {{.IdParamType}}) (*{{.DomainPkg}}.{{.ModelName}}, error) {
	var po {{.ModelPkg}}.{{.ModelName}}PO
	err := r.db.WithContext(ctx).Where("{{.IdColumn}} = ?", {{.IdQueryExpr}}).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find {{.ModelNameLower}} by id: %w", err)
	}
	return r.toDomain(&po), nil
}

// Save 保存
func (r *{{.ModelName}}RepositoryImpl) Save(ctx context.Context, b *{{.DomainPkg}}.{{.ModelName}}) error {
	po := r.toPO(b)
	return r.db.WithContext(ctx).Save(&po).Error
}

func (r *{{.ModelName}}RepositoryImpl) toDomain(po *{{.ModelPkg}}.{{.ModelName}}PO) *{{.DomainPkg}}.{{.ModelName}} {
{{.ToDomainBody}}
}

func (r *{{.ModelName}}RepositoryImpl) toPO(b *{{.DomainPkg}}.{{.ModelName}}) {{.ModelPkg}}.{{.ModelName}}PO {
	return {{.ModelPkg}}.{{.ModelName}}PO{
{{.ToPOBody}}
	}
}
