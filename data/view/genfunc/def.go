package genfunc

const (
	genTnf = `
// TableName get sql table name.获取数据库表名
func (m *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`
	genColumn = `
// {{.StructName}}Columns get sql column name.获取数据库列名
var {{.StructName}}Columns = struct { {{range $em := .Em}}
	{{$em.StructName}} string{{end}}    
	}{ {{range $em := .Em}}
		{{$em.StructName}}:"{{$em.ColumnName}}",  {{end}}           
	}
`
	genBase = `
package {{.PackageName}}
import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var globalIsRelated bool = true  // 全局预加载

// prepare for other
type _BaseMgr struct {
	*gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
	isRelated bool
}

// SetTimeOut set timeout
func (obj *_BaseMgr) SetTimeOut(timeout time.Duration) {
	obj.ctx, obj.cancel = context.WithTimeout(obj.ctx, timeout)
	obj.timeout = timeout
}

// Cancel cancel context
func (obj *_BaseMgr) Cancel(c context.Context) {
	obj.cancel()
}

// GetDB get gorm.DB info
func (obj *_BaseMgr) GetDB() *gorm.DB {
	return obj.DB
}

// UpdateDB update gorm.DB info
func (obj *_BaseMgr) UpdateDB(db *gorm.DB) {
	obj.DB = db
}

// GetIsRelated Query foreign key Association.获取是否查询外键关联(gorm.Related)
func (obj *_BaseMgr) GetIsRelated() bool {
	return obj.isRelated
}

// SetIsRelated Query foreign key Association.设置是否查询外键关联(gorm.Related)
func (obj *_BaseMgr) SetIsRelated(b bool) {
	obj.isRelated = b
}

// New new gorm.新gorm,重置条件
func (obj *_BaseMgr) new() {
	obj.DB = obj.newDB()
}

// NewDB new gorm.新gorm
func (obj *_BaseMgr) newDB() *gorm.DB {
	return obj.DB.Session(&gorm.Session{NewDB: true, Context: obj.ctx})
}

type options struct {
	query map[string]interface{}
}

// Option overrides behavior of Connect.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// OpenRelated 打开全局预加载
func OpenRelated() {
	globalIsRelated = true
}

// CloseRelated 关闭全局预加载
func CloseRelated() {
	globalIsRelated = true
}

// -------- sql where helper ----------

type CheckWhere func(v interface{}) bool
type DoWhere func(*gorm.DB, interface{}) *gorm.DB

// AddWhere
// CheckWhere 函数 如果返回true，则表明 DoWhere 的查询条件需要加到sql中去
func (obj *_BaseMgr) addWhere(v interface{}, c CheckWhere, d DoWhere) *_BaseMgr {
	if c(v) {
		obj.DB = d(obj.DB, v)
	}
	return obj
}

func (obj *_BaseMgr) sort(userSort, defaultSort string) *_BaseMgr {
	if len(userSort) > 0 {
		obj.DB = obj.DB.Order(userSort)
	} else {
		if len(defaultSort) > 0 {
			obj.DB = obj.DB.Order(defaultSort)
		}
	}
	return obj
}

	`

	genlogic = `{{$obj := .}}{{$list := $obj.Em}}
type _{{$obj.StructName}}Mgr struct {
	*_BaseMgr
}

// {{$obj.StructName}}Mgr open func
func {{$obj.StructName}}Mgr(db *gorm.DB) *_{{$obj.StructName}}Mgr {
	if db == nil {
		panic(fmt.Errorf("{{$obj.StructName}}Mgr need init by db"))
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &_{{$obj.StructName}}Mgr{_BaseMgr: &_BaseMgr{DB: db.Table("{{GetTablePrefixName $obj.TableName}}"), isRelated: globalIsRelated,ctx:ctx,cancel:cancel,timeout:-1}}
}


// WithContext set context to db
func (obj *_{{$obj.StructName}}Mgr) WithContext(c context.Context) *_{{$obj.StructName}}Mgr {
	if c != nil {
		obj.ctx = c
	}
	return obj
}

func (obj *_{{$obj.StructName}}Mgr) WithSelects(idName string, selects ...string) *_{{$obj.StructName}}Mgr {
	if len(selects) > 0 {
		if len(idName) > 0 {
			selects = append(selects, idName)
		}
		// 对Select进行去重
		selectMap := make(map[string]int, len(selects))
		for _, e := range selects {
			if _, ok := selectMap[e]; !ok {
				selectMap[e] = 1
			}
		}

		newSelects := make([]string, 0, len(selects))
		for k, _ := range selectMap {
			newSelects = append(newSelects, k)
		}

		obj.DB = obj.DB.Select(newSelects)
	}
	return obj
}

func (obj *_{{$obj.StructName}}Mgr) WithOmit(omit ...string) *_{{$obj.StructName}}Mgr {
	if len(omit) > 0 {
		obj.DB = obj.DB.Omit(omit...)
	}
	return obj
}

func (obj *_{{$obj.StructName}}Mgr) WithOptions(opts ...Option) *_{{$obj.StructName}}Mgr {
	options := options{
		query: make(map[string]interface{}, len(opts)),
	}
	for _, o := range opts {
		o.apply(&options)
	}
	obj.DB = obj.DB.Where(options.query)
	return obj
}

// GetTableName get sql table name.获取数据库名字
func (obj *_{{$obj.StructName}}Mgr) GetTableName() string {
	return "{{GetTablePrefixName $obj.TableName}}"
}

// Reset 重置gorm会话
func (obj *_{{$obj.StructName}}Mgr) Reset() *_{{$obj.StructName}}Mgr {
	obj.new()
	return obj
}

// Get 获取 
func (obj *_{{$obj.StructName}}Mgr) Get() (result {{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}

// Gets 获取批量结果
func (obj *_{{$obj.StructName}}Mgr) Gets() (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}

func (obj *_{{$obj.StructName}}Mgr) Count(count *int64) (tx *gorm.DB) {
	return obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Count(count)
}

{{range $oem := $obj.Em}}
// With{{$oem.ColStructName}} {{$oem.ColName}}获取 {{$oem.Notes}}
func (obj *_{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) Option {
	return optionFunc(func(o *options) { o.query["{{$oem.ColName}}"] = {{CapLowercase $oem.ColStructName}} })
}
{{end}}

{{range $oem := $obj.Em}}
// GetFrom{{$oem.ColStructName}} 通过{{$oem.ColName}}获取内容 {{$oem.Notes}} {{if $oem.IsMulti}}
func (obj *_{{$obj.StructName}}Mgr) GetFrom{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Where("{{$oem.ColNameEx}} = ?", {{CapLowercase $oem.ColStructName}}).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
{{else}}
func (obj *_{{$obj.StructName}}Mgr)  GetFrom{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} {{$oem.Type}}) (result {{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Where("{{$oem.ColNameEx}} = ?", {{CapLowercase $oem.ColStructName}}).Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}
{{end}}
// GetBatchFrom{{$oem.ColStructName}} 批量查找 {{$oem.Notes}}
func (obj *_{{$obj.StructName}}Mgr) GetBatchFrom{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}}s []{{$oem.Type}}) (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Where("{{$oem.ColNameEx}} IN (?)", {{CapLowercase $oem.ColStructName}}s).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
 {{end}}

func (obj *_{{$obj.StructName}}Mgr) Create{{$obj.StructName}}(bean *{{$obj.StructName}}) (err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Create(bean).Error

	return
}

func (obj *_{{$obj.StructName}}Mgr) Update{{$obj.StructName}}(bean *{{$obj.StructName}}) (err error) {
	err = obj.DB.WithContext(obj.ctx).Model(bean).Updates(bean).Error

	return
}

func (obj *_{{$obj.StructName}}Mgr) Delete{{$obj.StructName}}(bean *{{$obj.StructName}}) (err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Delete(bean).Error

	return
}

 {{range $ofm := $obj.Index}}
 // {{GenFListIndex $ofm 1}}  获取多个内容
 func (obj *_{{$obj.StructName}}Mgr) {{GenFListIndex $ofm 1}}({{GenFListIndex $ofm 2}}) (results []*{{$obj.StructName}}, err error) {
	err = obj.DB.WithContext(obj.ctx).Model({{$obj.StructName}}{}).Where("{{GenFListIndex $ofm 3}}", {{GenFListIndex $ofm 4}}).Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}
 {{end}}

`
	genPreload = `if err == nil && obj.isRelated { {{range $obj := .}}{{if $obj.IsMulti}}
		if err = obj.NewDB().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", result.{{$obj.ColStructName}}).Find(&result.{{$obj.ForeignkeyStructName}}List).Error;err != nil { // {{$obj.Notes}}
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}	
			} {{else}} 
		if err = obj.NewDB().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", result.{{$obj.ColStructName}}).Find(&result.{{$obj.ForeignkeyStructName}}).Error; err != nil { // {{$obj.Notes}} 
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}
			}{{end}} {{end}}}
`
	genPreloadMulti = `if err == nil && obj.isRelated {
		for i := 0; i < len(results); i++ { {{range $obj := .}}{{if $obj.IsMulti}}
		if err = obj.NewDB().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", results[i].{{$obj.ColStructName}}).Find(&results[i].{{$obj.ForeignkeyStructName}}List).Error;err != nil { // {{$obj.Notes}}
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}
			} {{else}} 
		if err = obj.NewDB().Table("{{$obj.ForeignkeyTableName}}").Where("{{$obj.ForeignkeyCol}} = ?", results[i].{{$obj.ColStructName}}).Find(&results[i].{{$obj.ForeignkeyStructName}}).Error; err != nil { // {{$obj.Notes}} 
				if err != gorm.ErrRecordNotFound { // 非 没找到
					return
				}
			} {{end}} {{end}}
	}
}`
)
