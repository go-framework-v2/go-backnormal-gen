package model

import (
	"time"
)

// {{.ModelName}}PO 持久化对象
type {{.ModelName}}PO struct {
{{range .Fields}}	{{.GoName}} {{.GoType}} {{.GormTag}}
{{end}}
}

func ({{.ModelName}}PO) TableName() string {
	return "{{.TableName}}"
}
