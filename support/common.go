package support

import (
	"bytes"
	"database/sql/driver"
	"github.com/alexedwards/argon2id"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"strings"
	"text/template"
)

func True() *bool {
	value := true
	return &value
}

func False() *bool {
	return new(bool)
}

type Array []interface{}

func (x *Array) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Array) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

const modelTpl = `
// Code generated by bit. DO NOT EDIT.

package model

import (
	"database/sql/driver"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"time"
)

type Array []interface{}

func (x *Array) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Array) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type Object map[string]interface{}

func (x *Object) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Object) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

func True() *bool {
	value := true
	return &value
}

func False() *bool {
	return new(bool)
}

{{range .}}
type {{title .Key}} struct {` +
	"ID     	int64	   `json:\"id\"`\n" +
	"Status     *bool      `gorm:\"default:true\" json:\"status\"`\n" +
	"CreateTime time.Time  `gorm:\"autoCreateTime;default:current_timestamp\" json:\"create_time\"`\n" +
	"UpdateTime time.Time  `gorm:\"autoUpdateTime;default:current_timestamp\" json:\"update_time\"`" + `
	{{range .Schema.Columns}} {{column .}} {{end}}
}
{{end}}

func AutoMigrate(tx *gorm.DB, models ...string) {
	mapper := map[string]interface{}{
		{{range .}}"{{ .Key}}": &{{title .Key}}{},{{end}}
	}

	for _, model := range models {
		if mapper[model] != nil {
			tx.AutoMigrate(mapper[model])
		}
	}
}
`

func GenerateModels(tx *gorm.DB) (buf bytes.Buffer, err error) {
	var resources []Resource
	if err = tx.
		Where("schema ->> 'type' = ?", "collection").
		Find(&resources).Error; err != nil {
		return
	}
	var tmpl *template.Template
	if tmpl, err = template.New("model").Funcs(template.FuncMap{
		"title":  title,
		"column": column,
	}).Parse(modelTpl); err != nil {
		return
	}
	if err = tmpl.Execute(&buf, resources); err != nil {
		return
	}
	return
}

func title(s string) string {
	return strings.Title(s)
}

func dataType(val string) string {
	switch val {
	case "int":
		return "int32"
	case "int8":
		return "int64"
	case "decimal":
		return "float64"
	case "float8":
		return "float64"
	case "varchar":
		return "string"
	case "text":
		return "string"
	case "bool":
		return "*bool"
	case "timestamptz":
		return "time.Time"
	case "uuid":
		return "uuid.UUID"
	case "object":
		return "Object"
	case "array":
		return "Array"
	case "rel":
		return "Array"
	}
	return val
}

func column(val Column) string {
	var b strings.Builder
	b.WriteString(title(val.Key))
	b.WriteString(" ")
	b.WriteString(dataType(val.Type))
	b.WriteString(" `")
	b.WriteString(`gorm:"type:`)
	if val.Type == "object" || val.Type == "array" || val.Type == "rel" {
		b.WriteString("jsonb")
	} else {
		b.WriteString(val.Type)
	}
	if val.Require {
		b.WriteString(`;not null`)
	}
	if val.Unique {
		b.WriteString(`;unique`)
	}
	if val.Default != "" {
		b.WriteString(`;default:`)
		b.WriteString(val.Default)
	}
	b.WriteString(`" json:"`)
	if val.Hide {
		b.WriteString(`-`)
	} else {
		b.WriteString(val.Key)
	}
	b.WriteString(`"`)
	b.WriteString("`\n")
	return b.String()
}

func InitSeeder(tx *gorm.DB) (err error) {
	roles := []map[string]interface{}{
		{
			"key":         "*",
			"name":        "超级管理员",
			"description": "超级管理员拥有完整权限不能编辑，若不使用可以禁用该权限",
			"permissions": Array{},
		},
		{
			"key":         "admin",
			"name":        "管理员",
			"description": "分配管理用户",
			"permissions": Array{
				"resource:*",
				"role:*",
				"admin:*",
			},
		},
	}
	if err = tx.Table("role").Create(&roles).Error; err != nil {
		return
	}
	var password string
	if password, err = argon2id.CreateHash(
		"pass@VAN1234",
		argon2id.DefaultParams,
	); err != nil {
		return
	}
	admins := []map[string]interface{}{
		{
			"username": "admin",
			"password": password,
			"roles":    Array{"*"},
		},
	}
	if err = tx.Table("admin").Create(&admins).Error; err != nil {
		return
	}
	return
}
