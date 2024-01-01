package template

import "meta-egg/internal/domain/helper"

var TplPkgGormxOption string = helper.PH_META_EGG_HEADER + `
package gormx

import "gorm.io/gorm"

type Option func(tx *gorm.DB) *gorm.DB

// example: Select("id", "name")
// 不和Distinct一起使用
func Select(cols ...interface{}) Option {
	return func(tx *gorm.DB) *gorm.DB {
		if len(cols) > 0 {
			return tx.Select(cols[0], cols[1:]...)
		} else {
			return tx
		}
	}
}

// example: Order("id desc, name asc")
// example: Order("field(id, 45, 44)")
func Order(order string) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order(order)
	}
}

// example: Limit(10)
func Limit(limit int) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Limit(limit)
	}
}

// example: Offset(10)
func Offset(offset int) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Offset(offset)
	}
}

// 参数定义同gorm的Where
// example: Where("id = ?", 1)
func Where(query interface{}, args ...interface{}) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(query, args...)
	}
}

// 参数定义同gorm的Unscoped, 可忽略软删除
func Unscoped() Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Unscoped()
	}
}

// 参数定义同gorm的Distinct
// example: Distinct("id", "name")
// 不和Select一起使用
func Distinct(cols ...interface{}) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Distinct(cols...)
	}
}

// 参数定义同gorm的Group
// example: Group("name")
// 每个查询只能有一个Group
func Group(group string) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Group(group)
	}
}

// 参数定义同gorm的Joins
// 传入JOIN语句
// 范例: Join("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "foo@bar.org")
func Join(join string, args ...interface{}) Option {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Joins(join, args...)
	}
}
`
