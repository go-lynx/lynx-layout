// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt64, Increment: true, Comment: "主键ID"},
		{Name: "num", Type: field.TypeString, Unique: true, Size: 32, Comment: "编号"},
		{Name: "account", Type: field.TypeString, Size: 32, Comment: "账号", Default: ""},
		{Name: "password", Type: field.TypeString, Size: 100, Comment: "密码", Default: ""},
		{Name: "phone", Type: field.TypeString, Size: 11, Comment: "手机号", Default: ""},
		{Name: "nickname", Type: field.TypeString, Size: 20, Comment: "昵称"},
		{Name: "avatar", Type: field.TypeString, Size: 240, Comment: "个人肖像", Default: ""},
		{Name: "stats", Type: field.TypeInt32, Comment: "账号状态 1:正常;2:封禁", Default: 1},
		{Name: "note", Type: field.TypeString, Size: 120, Comment: "个人简介", Default: ""},
		{Name: "register_source", Type: field.TypeInt32, Comment: "注册来源 1:web端;2:app端"},
		{Name: "last_login_at", Type: field.TypeTime, Comment: "最近登录时间", SchemaType: map[string]string{"mysql": "datetime(3)"}},
		{Name: "created_at", Type: field.TypeTime, Comment: "创建时间", SchemaType: map[string]string{"mysql": "datetime(3)"}},
		{Name: "updated_at", Type: field.TypeTime, Comment: "修改时间", SchemaType: map[string]string{"mysql": "datetime(3)"}},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Comment:    "用户表",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		UsersTable,
	}
)

func init() {
	UsersTable.Annotation = &entsql.Annotation{}
}
