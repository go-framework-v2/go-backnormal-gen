package bo

import "{{.PoPath}}"

type {{.ModelName}}Bo struct {
	po.{{.ModelName}} `bson:",inline"` // 添加 inline 标签
}