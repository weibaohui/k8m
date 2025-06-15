package dao

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// GenericQuery 是一个通用查询方法，适用于任意模型
func GenericQuery[T any](params *Params, model T, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]T, int64, error) {
	var total int64
	var results []T

	// 如果params为nil，创建一个默认的params
	if params == nil {
		params = &Params{
			OrderBy:  "id",
			OrderDir: "desc",
			Page:     1,
			PerPage:  15,
			Queries:  make(map[string]interface{}),
		}
	}

	// 如果CreatedBy为空，则补全
	if reflect.ValueOf(model).Elem().FieldByName("CreatedBy").String() == "" && params.UserName != "" {
		reflect.ValueOf(model).Elem().FieldByName("CreatedBy").SetString(params.UserName)
	}

	// 构建数据库查询
	dbQuery := DB().Model(model)

	// 定义允许过滤的字段列表
	validFields, err := GetTableFieldsWithCache(model)
	if err != nil {
		return nil, 0, err
	}

	// 动态添加搜索条件
	for key, value := range params.Queries {
		if validFields[key] && value != "" && value != nil {
			dbQuery = dbQuery.Where(key+" like ?", "%"+fmt.Sprintf("%s", value)+"%")
		}
	}

	// 执行自定义查询函数
	for _, fn := range queryFuncs {
		dbQuery = fn(dbQuery)
	}

	// 获取总记录数
	dbQuery.Count(&total)

	// 排序
	// 检查order by 是否设置了值
	order := params.OrderBy + " " + params.OrderDir
	if len(strings.TrimSpace(order)) > 0 {
		dbQuery = dbQuery.Order(order)
	} else {
		// 默认按ID倒排
		dbQuery = dbQuery.Order(" id desc")
	}

	// 分页
	offset := (params.Page - 1) * params.PerPage
	dbQuery = dbQuery.Offset(offset).Limit(params.PerPage)

	// 执行查询
	if err := dbQuery.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
func GenericGetOne[T any](params *Params, model T, queryFuncs ...func(*gorm.DB) *gorm.DB) (T, error) {
	// 如果params为nil，创建一个默认的params
	if params == nil {
		params = &Params{
			Queries: make(map[string]interface{}),
		}
	}

	// 如果CreatedBy为空，则补全
	if reflect.ValueOf(model).Elem().FieldByName("CreatedBy").String() == "" && params.UserName != "" {
		reflect.ValueOf(model).Elem().FieldByName("CreatedBy").SetString(params.UserName)
	}

	dbQuery := DB().Model(model)

	// 执行自定义查询函数
	for _, fn := range queryFuncs {
		dbQuery = fn(dbQuery)
	}

	err := dbQuery.Limit(1).First(model).Error
	return model, err
}

// GenericSave 是一个通用保存方法，适用于任意模型
func GenericSave[T any](params *Params, model T, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	// 如果params为nil，创建一个默认的params
	if params == nil {
		params = &Params{
			Queries: make(map[string]interface{}),
		}
	}

	// 如果CreatedBy为空，则补全
	if reflect.ValueOf(model).Elem().FieldByName("CreatedBy").String() == "" && params.UserName != "" {
		reflect.ValueOf(model).Elem().FieldByName("CreatedBy").SetString(params.UserName)
	}

	dbQuery := DB().Model(model)

	// 执行自定义查询函数
	for _, fn := range queryFuncs {
		dbQuery = fn(dbQuery)
	}

	// 保存数据
	if err := dbQuery.Save(model).Error; err != nil {
		return err
	}

	return nil
}

// GenericDelete 是一个通用的删除方法，适用于任意模型
func GenericDelete[T any](params *Params, model T, ids []int64, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	// 如果params为nil，创建一个默认的params
	if params == nil {
		params = &Params{
			Queries: make(map[string]interface{}),
		}
	}

	// 如果CreatedBy为空，则补全
	if reflect.ValueOf(model).Elem().FieldByName("CreatedBy").String() == "" && params.UserName != "" {
		reflect.ValueOf(model).Elem().FieldByName("CreatedBy").SetString(params.UserName)
	}

	// 构建数据库查询
	dbQuery := DB().Model(model)
	dbQuery = dbQuery.Where(model)

	// 执行自定义查询函数
	for _, fn := range queryFuncs {
		dbQuery = fn(dbQuery)
	}

	// 执行删除操作
	if err := dbQuery.Delete(model, ids).Error; err != nil {
		return err
	}

	return nil
}

// GenericBatchSave 是一个通用的批量插入方法，适用于任意模型切片
func GenericBatchSave[T any](params *Params, models []T, batchSize int, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	// 如果params为nil，创建一个默认的params
	if params == nil {
		params = &Params{
			Queries:  make(map[string]interface{}),
			UserName: "",
		}
	}

	// 如果切片为空，直接返回
	if len(models) == 0 {
		return nil
	}

	// 如果未指定批量大小，设置默认值
	if batchSize <= 0 {
		batchSize = 100
	}

	// 为每个模型补全CreatedBy
	for i := range models {
		if params.UserName != "" {
			if reflect.ValueOf(&models[i]).Elem().FieldByName("CreatedBy").String() == "" {
				reflect.ValueOf(&models[i]).Elem().FieldByName("CreatedBy").SetString(params.UserName)
			}
		}
	}

	dbQuery := DB().Model(&models[0])

	// 执行自定义查询函数
	for _, fn := range queryFuncs {
		dbQuery = fn(dbQuery)
	}

	// 使用CreateInBatches执行批量插入
	if err := dbQuery.CreateInBatches(models, batchSize).Error; err != nil {
		return err
	}

	return nil
}

// GetTableName 使用 GORM 的 Statement 获取模型对应的表名
func GetTableName(model interface{}) (string, error) {
	stmt := &gorm.Statement{DB: DB()}
	err := stmt.Parse(model)
	if err != nil {
		return "", err
	}
	return stmt.Table, nil
}

// 定义缓存存储表字段信息
var fieldCache sync.Map

// GetTableFieldsWithCache 获取表字段并缓存结果
func GetTableFieldsWithCache(model interface{}) (map[string]bool, error) {
	// 获取表名
	tableName, _ := GetTableName(model)

	// 先从缓存中查找是否已经有该表的字段
	if cachedFields, ok := fieldCache.Load(tableName); ok {
		return cachedFields.(map[string]bool), nil
	}

	// 如果缓存中没有，查询数据库字段并缓存
	fields, err := GetTableFields(model)
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	fieldCache.Store(tableName, fields)
	return fields, nil
}

// GetTableFields 自动获取表名并从数据库中获取表的字段
func GetTableFields(model interface{}) (map[string]bool, error) {
	validFields := make(map[string]bool)

	// 使用 Statement 自动获取表名
	stmt := &gorm.Statement{DB: DB()}
	err := stmt.Parse(model)
	if err != nil {
		return nil, err
	}
	tableName := stmt.Table

	// 获取表的列信息
	columns, err := DB().Migrator().ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}

	// 遍历列信息，提取列名
	for _, column := range columns {
		fieldName := column.Name() // 获取字段名
		validFields[fieldName] = true
	}

	return validFields, nil
}
