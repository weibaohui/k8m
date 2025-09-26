package dao

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

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
			Queries:  make(map[string]any),
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
			Queries: make(map[string]any),
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

// GenericUpdateColumn 更新某个列
func GenericUpdateColumn[T any](model T, columnName string, value any) error {
	err := DB().Model(model).Update(columnName, value).Error
	return err
}

// GenericSave 是一个通用保存方法，适用于任意模型
func GenericSave[T any](params *Params, model T, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	// 如果params为nil，创建一个默认的params
	if params == nil {
		params = &Params{
			Queries: make(map[string]any),
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
			Queries: make(map[string]any),
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
			Queries:  make(map[string]any),
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
func GetTableName(model any) (string, error) {
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
func GetTableFieldsWithCache(model any) (map[string]bool, error) {
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
func GetTableFields(model any) (map[string]bool, error) {
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

// BuildCreatedAtQuery 构建时间范围查询函数
// 解析时间范围参数并返回对应的查询函数
// 这是一个通用方法，可以被任何需要时间范围过滤的控制器使用
// paramName: 可选参数，指定查询参数名称，不提供则默认使用 "created_at_range"
func BuildCreatedAtQuery(params *Params, paramName ...string) (func(*gorm.DB) *gorm.DB, bool) {
	// 确定参数名称，默认为 "created_at_range"
	queryParam := "created_at_range"
	if len(paramName) > 0 && paramName[0] != "" {
		queryParam = paramName[0]
	}
	
	v, ok := params.Queries[queryParam]
	if !ok || v == "" {
		return nil, false
	}
	timeRange := fmt.Sprintf("%v", v)
	if !strings.Contains(timeRange, ",") {
		return nil, false
	}
	parts := strings.SplitN(timeRange, ",", 2)
	if len(parts) != 2 {
		return nil, false
	}
	startStr, endStr := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	delete(params.Queries, queryParam)
	return func(db *gorm.DB) *gorm.DB {
		if startStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04:05", startStr, time.Local); err == nil {
				db = db.Where("created_at >= ?", t)
			}
		}
		if endStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04:05", endStr, time.Local); err == nil {
				db = db.Where("created_at <= ?", t)
			}
		}
		return db
	}, true
}
