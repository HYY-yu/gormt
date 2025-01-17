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
// Code generated by gormt. DO NOT EDIT.
package {{.PackageName}}
import (
	"context"

	"gorm.io/gorm"
)

var globalIsRelated bool = true  // 全局预加载

// prepare for other
type _BaseMgr struct {
	*gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	isRelated bool
}

// Cancel cancel context
func (obj *_BaseMgr) Cancel(c context.Context) {
	obj.cancel()
}

// GetDB get gorm.DB info
func (obj *_BaseMgr) GetDB() *gorm.DB {
	return obj.DB
}

// GetIsRelated Query foreign key Association.获取是否查询外键关联(gorm.Related)
func (obj *_BaseMgr) GetIsRelated() bool {
	return obj.isRelated
}

// SetIsRelated Query foreign key Association.设置是否查询外键关联(gorm.Related)
func (obj *_BaseMgr) SetIsRelated(b bool) {
	obj.isRelated = b
}

type options struct {
	query map[string]queryData
}

type queryData struct {
	data interface{}
	cond string
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

// sort 不能改变 obj.DB ，变成已初始化的DB，将会对后续语句执行造成影响
// 因此主动返回 已初始化的DB ，表示：此sort方法只能使用一次
func (obj *_BaseMgr) sort(userSort, defaultSort string) *gorm.DB {
	if len(userSort) > 0 {
		return obj.DB.Order(userSort)
	} else {
		if len(defaultSort) > 0 {
			return obj.DB.Order(defaultSort)
		}
	}
	return obj.DB
}

	`

	genlogic = `
// Code generated by gormt. DO NOT EDIT.

// 非线程安全
{{$obj := .}}{{$list := $obj.Em}}
type _{{$obj.StructName}}Mgr struct {
	*_BaseMgr
}

// {{$obj.StructName}}Mgr open func
func {{$obj.StructName}}Mgr(ctx context.Context, db *gorm.DB) *_{{$obj.StructName}}Mgr {
	if db == nil {
		panic(fmt.Errorf("{{$obj.StructName}}Mgr need init by db"))
	}
	ctx, cancel := context.WithCancel(ctx)
	return &_{{$obj.StructName}}Mgr{_BaseMgr: &_BaseMgr{DB: db.Table("{{GetTablePrefixName $obj.TableName}}").WithContext(ctx), isRelated: globalIsRelated,ctx:ctx,cancel:cancel}}
}

func (obj *_{{$obj.StructName}}Mgr) WithSelects(idName string, selects ...string) *_{{$obj.StructName}}Mgr {
    if len(idName) > 0 {
		selects = append(selects, idName)
	}
	if len(selects) > 0 {
		// 对Select进行去重
		selectMap := make(map[string]int, len(selects))
		for _, e := range selects {
			if _, ok := selectMap[e]; !ok {
				selectMap[e] = 1
			}
		}

		newSelects := make([]string, 0, len(selects))
		for k := range selectMap {
			if len(k) > 0 {
				newSelects = append(newSelects, k)
			}
		}
		obj.DB = obj.DB.Select(newSelects)
	}
	return obj
}

func (obj *_{{$obj.StructName}}Mgr) WithOptions(opts ...Option) *_{{$obj.StructName}}Mgr {
	obj.Reset()

	options := options{
		query: make(map[string]queryData, len(opts)),
	}
	for _, o := range opts {
		o.apply(&options)
	}
	for k, v := range options.query {
		if v.data == nil {
			obj.DB = obj.DB.Where(k + " " + v.cond)
		} else {
			obj.DB = obj.DB.Where(k+" "+v.cond, v.data)
		}
	}
	return obj
}

// GetTableName get sql table name.获取表名字
func (obj *_{{$obj.StructName}}Mgr) GetTableName() string {
	return "{{GetTablePrefixName $obj.TableName}}"
}

// Tx 开启事务会话
func (obj *_{{$obj.StructName}}Mgr) Tx(tx *gorm.DB) *_{{$obj.StructName}}Mgr {
	obj.DB = tx.Table(obj.GetTableName()).WithContext(obj.ctx)
	return obj
}

// WithPrepareStmt 开启语句 PrepareStmt 功能
// 接下来执行的SQL将会是PrepareStmt的
func (obj *_{{$obj.StructName}}Mgr) WithPrepareStmt() {
	obj.DB = obj.DB.Session(&gorm.Session{Context: obj.ctx, PrepareStmt: true})
}

// Reset 重置gorm会话
func (obj *_{{$obj.StructName}}Mgr) Reset() *_{{$obj.StructName}}Mgr {
	obj.DB = obj.DB.Session(&gorm.Session{NewDB: true, Context: obj.ctx}).Table(obj.GetTableName())
	return obj
}

// Get 获取 
func (obj *_{{$obj.StructName}}Mgr) Get() (result {{$obj.StructName}}, err error) {
	err = obj.DB.Find(&result).Error
	{{GenPreloadList $obj.PreloadList false}}
	return
}

// Gets 获取批量结果
func (obj *_{{$obj.StructName}}Mgr) Gets() (results []{{$obj.StructName}}, err error) {
	err = obj.DB.Find(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}

// Catch 必须获取结果（单条）
func(obj *_{{$obj.StructName}}Mgr) Catch() (results {{$obj.StructName}}, err error) {
	err = obj.DB.Take(&results).Error
	{{GenPreloadList $obj.PreloadList true}}
	return
}

func (obj *_{{$obj.StructName}}Mgr) Count() (count int64, err error) {
	err = obj.DB.Count(&count).Error

	return
}

func (obj *_{{$obj.StructName}}Mgr) HasRecord() (bool, error) {
	count, err := obj.Count()
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

{{range $oem := $obj.Em}}
// With{{$oem.ColStructName}} {{$oem.ColName}}获取 {{$oem.Notes}}
func (obj *_{{$obj.StructName}}Mgr) With{{$oem.ColStructName}}({{CapLowercase $oem.ColStructName}} interface{}, cond ...string) Option {
	return optionFunc(func(o *options) {
		if len(cond) == 0 {
			cond = []string{" = ? "}
		}
		o.query["{{$oem.ColName}}"] = queryData{
			cond: cond[0],
			data: {{CapLowercase $oem.ColStructName}},
		}
	})
}
{{end}}

func (obj *_{{$obj.StructName}}Mgr) Create{{$obj.StructName}}(bean *{{$obj.StructName}}) (err error) {
	err = obj.DB.Create(bean).Error

	return
}

func (obj *_{{$obj.StructName}}Mgr) Update{{$obj.StructName}}(bean *{{$obj.StructName}}) (err error) {
	err = obj.DB.Updates(bean).Error

	return
}

func (obj *_{{$obj.StructName}}Mgr) Delete{{$obj.StructName}}(bean *{{$obj.StructName}}) (err error) {
	err = obj.DB.Delete(bean).Error

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
