package support

import (
	"bytes"
	"gorm.io/gorm"
	"strings"
	"text/template"
)

const modelTpl = `
// Code generated by bit. DO NOT EDIT.

package model

import (
	"database/sql/driver"
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
	"ID     	int64\n" +
	"Status     *bool      `gorm:\"default:true\"`\n" +
	"CreateTime time.Time  `gorm:\"autoCreateTime\"`\n" +
	"UpdateTime time.Time  `gorm:\"autoUpdateTime\"`" + `
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
		Where("schema <> ?", "{}").
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
	case "jsonb":
		return "Object"
	case "uuid":
		return "uuid.UUID"
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
	if val.Type != "rel" {
		b.WriteString(val.Type)
	} else {
		b.WriteString("jsonb")
		val.Default = `'[]'`
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
	b.WriteString(`"`)
	if val.Hide {
		b.WriteString(` json:"-"`)
	}
	b.WriteString("`\n")
	return b.String()
}
