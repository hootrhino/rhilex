package apis

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/datacenter"
	"github.com/hootrhino/rhilex/component/dataschema"
	"github.com/hootrhino/rhilex/component/interdb"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"gorm.io/gorm"
)

/*
*
* 模型管理
*
 */
func InitDataSchemaApi() {

	schemaApi := server.RouteGroup(server.ContextUrl("/schema"))
	{
		// 物模型
		schemaApi.POST("/create", server.AddRoute(CreateDataSchema))
		schemaApi.DELETE("/del", server.AddRoute(DeleteDataSchema))
		schemaApi.PUT("/update", server.AddRoute(UpdateDataSchema))
		schemaApi.GET("/list", server.AddRoute(ListDataSchema))
		schemaApi.GET(("/detail"), server.AddRoute(DataSchemaDetail))
		schemaApi.POST(("/publish"), server.AddRoute(PublishSchema))
		schemaApi.POST(("/fix"), server.AddRoute(FixSchema))
		// 属性
		schemaApi.POST(("/properties/create"), server.AddRoute(CreateIotSchemaProperty))
		schemaApi.PUT(("/properties/update"), server.AddRoute(UpdateIotSchemaProperty))
		schemaApi.DELETE(("/properties/del"), server.AddRoute(DeleteIotSchemaProperty))
		schemaApi.GET(("/properties/list"), server.AddRoute(IotSchemaPropertyPageList))
		schemaApi.GET(("/properties/detail"), server.AddRoute(IotSchemaPropertyDetail))
		// 模板
		schemaApi.GET(("/getTemplates"), server.AddRoute(GetTemplates))
		schemaApi.GET(("/getTemplateFields"), server.AddRoute(GetTemplateFields))
		schemaApi.POST(("/genTemplate"), server.AddRoute(GenerateSchemaTemplate))

	}
}

/*
*
* 新建模型
*
 */
type IoTSchemaVo struct {
	UUID        string `json:"uuid,omitempty"`
	Published   *bool  `json:"published"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

/*
*
* 属性 @ component/dataschema/iot_schema_define
*
 */
type IotPropertyVo struct {
	UUID        string            `json:"uuid"`        // UUID
	SchemaId    string            `json:"schemaId"`    //模型ID
	Label       string            `json:"label"`       // UI显示的那个文本
	Name        string            `json:"name"`        // 变量关联名
	Description string            `json:"description"` // 额外信息
	Type        string            `json:"type"`        // 类型, 只能是上面几种
	Rw          string            `json:"rw"`          // R读 W写 RW读写
	Unit        string            `json:"unit"`        // 单位 例如：摄氏度、米、牛等等
	Rule        IoTPropertyRuleVo `json:"rule"`        // 规则,IoTPropertyRule
}
type IoTPropertyRuleVo struct {
	DefaultValue interface{} `json:"defaultValue"` // 默认值
	Max          *int        `json:"max"`          // 最大值
	Min          *int        `json:"min"`          // 最小值
	TrueLabel    string      `json:"trueLabel"`    // 真值label
	FalseLabel   string      `json:"falseLabel"`   // 假值label
	Round        *int        `json:"round"`        // 小数点位
}

func (O IoTPropertyRuleVo) Check() error {
	return nil
}
func (O IoTPropertyRuleVo) GetDefaultValue() string {
	switch T := O.DefaultValue.(type) {
	case string:
		return T
	case int32:
		return fmt.Sprintf("%d", T)
	case int64:
		return fmt.Sprintf("%d", T)
	case bool:
		if T {
			return "1"
		}
		return "0"
	default:
		return "0"
	}
}

/*
*
* 属性
*
 */
func (O IoTPropertyRuleVo) String() string {
	if O.Max == nil {
		O.Max = new(int)
	}
	if O.Min == nil {
		O.Min = new(int)
	}
	if O.Round == nil {
		O.Round = new(int)
	}
	if O.DefaultValue == nil {
		O.DefaultValue = ""
	}
	if bytes, err := json.Marshal(O); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

/*
*
* 从数据库反解析
*
 */
func (O IoTPropertyRuleVo) IoTPropertyRuleFromString(s string) error {
	if err := json.Unmarshal([]byte(s), &O); err != nil {
		return err
	}
	return nil
}

/*
*
* 从数据库保存的String字符串反解析规则
*
 */
func (O *IoTPropertyRuleVo) ParseRuleFromModel(s string) error {
	if O.Max == nil {
		O.Max = new(int)
	}
	if O.Min == nil {
		O.Min = new(int)
	}
	if O.Round == nil {
		O.Round = new(int)
	}

	if O.DefaultValue == nil {
		O.DefaultValue = ""
	}

	if err := json.Unmarshal([]byte(s), &O); err != nil {
		return err
	} else {
		return nil
	}
}

/*
*
* 新建一个物模型表
*
 */
func CreateDataSchema(c *gin.Context, ruleEngine typex.Rhilex) {

	IoTSchemaVo := IoTSchemaVo{}
	if err := c.ShouldBindJSON(&IoTSchemaVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	MIotSchema := model.MIotSchema{
		UUID:        utils.DataSchemaUuid(),
		Published:   new(bool),
		Name:        IoTSchemaVo.Name,
		Description: IoTSchemaVo.Description,
	}
	if err := service.InsertDataSchema(MIotSchema); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 更新模型
*
 */
func UpdateDataSchema(c *gin.Context, ruleEngine typex.Rhilex) {
	IoTSchemaVo := IoTSchemaVo{}
	if err := c.ShouldBindJSON(&IoTSchemaVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	MIotSchema := model.MIotSchema{
		UUID:        IoTSchemaVo.UUID,
		Name:        IoTSchemaVo.Name,
		Description: IoTSchemaVo.Description,
	}
	if err := service.UpdateDataSchema(MIotSchema); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 删除模型
*
 */
func DeleteDataSchema(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	if err := service.DeleteDataSchemaAndProperty(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 修复数据仓库，本质上是删除可能存在的垃圾数据表，然后重新建表
*
 */
func FixSchema(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	if err := service.ResetSchema(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 发布模型
*
 */
func PublishSchema(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	MSchema, err := service.GetDataSchemaWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if *MSchema.Published {
		c.JSON(common.HTTP_OK, common.Error("Data Schema Already published"))
		return
	}
	var records []model.MIotProperty
	result := interdb.InterDb().Order("created_at DESC").
		Find(&records, &model.MIotProperty{SchemaId: MSchema.UUID})
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	if len(records) == 0 {
		c.JSON(common.HTTP_OK, common.Error("Must contain at least one property"))
		return
	}
	DDLColumns := []datacenter.DDLColumn{}
	// 默认加入PK
	DDLColumns = append(DDLColumns, datacenter.DDLColumn{
		Name: "id", Type: "INTEGER", Description: "PRIMARY KEY",
	})
	DDLColumns = append(DDLColumns, datacenter.DDLColumn{
		Name: "create_at", Type: "DATETIME", Description: "DATETIME", DefaultValue: "CURRENT_TIMESTAMP",
	})
	ioTPropertyRuleVo := IoTPropertyRuleVo{}
	for _, record := range records {
		if err0 := ioTPropertyRuleVo.ParseRuleFromModel(record.Rule); err0 != nil {
			c.JSON(common.HTTP_OK, common.Error400(err0))
			return
		}
		DDLColumns = append(DDLColumns, datacenter.DDLColumn{
			Name:         record.Name,
			Type:         record.Type,
			Description:  record.Description,
			DefaultValue: ioTPropertyRuleVo.GetDefaultValue(),
		})
	}
	txErr := interdb.InterDb().Transaction(func(tx *gorm.DB) error {
		// Publish Schema
		True := true
		MSchema.Published = &True
		err3 := tx.Model(MSchema).Where("uuid=?", MSchema.UUID).Updates(&MSchema).Error
		if err3 != nil {
			return err3
		}
		// 发布完了还得生成表结构
		tableName := fmt.Sprintf("data_center_%s", MSchema.UUID)
		sql, err1 := datacenter.GenerateSQLiteCreateTableDDL(datacenter.SchemaDDL{
			SchemaUUID: tableName,
			DDLColumns: DDLColumns,
		})
		if err1 != nil {
			return err1
		}
		errTransaction := datacenter.DataCenterDb().Transaction(func(datacenterTx *gorm.DB) error {
			if err2 := datacenter.DataCenterDb().Exec(sql).Error; err2 != nil {
				return err2
			}
			idxSql1 := `CREATE INDEX IF NOT EXISTS idx_id ON %s (id DESC);`
			if errIdx1 := datacenter.DataCenterDb().Exec(fmt.Sprintf(idxSql1, tableName)).Error; errIdx1 != nil {
				return errIdx1
			}
			idxSql2 := `CREATE INDEX IF NOT EXISTS idx_create_at ON %s (create_at DESC);`
			if errIdx2 := datacenter.DataCenterDb().Exec(fmt.Sprintf(idxSql2, tableName)).Error; errIdx2 != nil {
				return errIdx2
			}
			// 防止数据爆炸撑死数据库，10000条数据大概吃10Mb硬盘.
			// 通常来说这些数据完全足够代表最新数据了，历史数据反而是垃圾
			trigger := `
CREATE TRIGGER IF NOT EXISTS "%s"
AFTER INSERT ON "%s"
WHEN ((SELECT COUNT(*) FROM "%s") / 100) * 100 = (SELECT COUNT(*) FROM "%s")
AND (SELECT COUNT(*) FROM "%s") > 10000
BEGIN
    DELETE FROM "%s"
    WHERE id IN (
        SELECT id FROM "%s"
        ORDER BY id ASC
        LIMIT 100
    );
END;
`
			if errTrigger := datacenter.DataCenterDb().Exec(fmt.Sprintf(trigger, tableName, tableName,
				tableName, tableName, tableName, tableName, tableName)).Error; errTrigger != nil {
				return errTrigger
			}
			return nil
		})

		return errTransaction
	})
	if txErr != nil {
		c.JSON(common.HTTP_OK, common.Error400(txErr))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 模型列表
*
 */
func ListDataSchema(c *gin.Context, ruleEngine typex.Rhilex) {
	DataSchemas := []IoTSchemaVo{}
	for _, vv := range service.AllDataSchema() {
		IoTSchemaVo := IoTSchemaVo{
			UUID:        vv.UUID,
			Published:   vv.Published,
			Name:        vv.Name,
			Description: vv.Description,
		}
		DataSchemas = append(DataSchemas, IoTSchemaVo)
	}
	c.JSON(common.HTTP_OK, common.OkWithData(DataSchemas))
}

/*
*
* 模型详情
*
 */
func DataSchemaDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	MIotSchema, err := service.GetDataSchemaWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	IoTSchemaVo := IoTSchemaVo{
		UUID:        MIotSchema.UUID,
		Published:   MIotSchema.Published,
		Name:        MIotSchema.Name,
		Description: MIotSchema.Description,
	}
	c.JSON(common.HTTP_OK, common.OkWithData(IoTSchemaVo))
}

// CreateIotSchemaProperty
func CreateIotSchemaProperty(c *gin.Context, ruleEngine typex.Rhilex) {
	IotPropertyVo := IotPropertyVo{}
	if err := c.ShouldBindJSON(&IotPropertyVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := dataschema.CheckPropertyType(IotPropertyVo.Type); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := dataschema.ValidateRw(IotPropertyVo.Rw); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Schema, err := service.GetDataSchemaWithUUID(IotPropertyVo.SchemaId)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if *Schema.Published {
		c.JSON(common.HTTP_OK, common.Error("Data Schema Already published"))
		return
	}
	// Check Published
	if *Schema.Published {
		c.JSON(common.HTTP_OK, common.Error("Schema Already Published:"+IotPropertyVo.Name))
		return
	}
	// 不允许重复name
	count := service.CountIotSchemaProperty(IotPropertyVo.Name, Schema.UUID)
	if count > 0 {
		c.JSON(common.HTTP_OK, common.Error("Already Exists Property:"+IotPropertyVo.Name))
		return
	}

	err2 := service.InsertIotSchemaProperty(model.MIotProperty{
		SchemaId:    IotPropertyVo.SchemaId,
		UUID:        utils.MakeUUID("PROPER"),
		Label:       IotPropertyVo.Label,
		Name:        IotPropertyVo.Name,
		Description: IotPropertyVo.Description,
		Type:        IotPropertyVo.Type,
		Rw:          IotPropertyVo.Rw,
		Unit:        IotPropertyVo.Unit,
		Rule:        IotPropertyVo.Rule.String(), // 规则
	})
	if err2 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err2))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

// 更新属性
func UpdateIotSchemaProperty(c *gin.Context, ruleEngine typex.Rhilex) {
	IotPropertyVo := IotPropertyVo{}
	if err := c.ShouldBindJSON(&IotPropertyVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := dataschema.CheckPropertyType(IotPropertyVo.Type); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Schema, err := service.GetDataSchemaWithUUID(IotPropertyVo.SchemaId)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if *Schema.Published {
		c.JSON(common.HTTP_OK, common.Error("Data Schema Already published"))
		return
	}
	err2 := service.UpdateIotSchemaProperty(model.MIotProperty{
		SchemaId:    IotPropertyVo.SchemaId,
		UUID:        IotPropertyVo.UUID,
		Label:       IotPropertyVo.Label,
		Name:        IotPropertyVo.Name,
		Description: IotPropertyVo.Description,
		Type:        IotPropertyVo.Type,
		Rw:          IotPropertyVo.Rw,
		Unit:        IotPropertyVo.Unit,
		Rule:        IotPropertyVo.Rule.String(), // 规则
	})
	if err2 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err2))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

// 删除属性
func DeleteIotSchemaProperty(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	// Find Schema By UUID
	Property, err := service.FindIotSchemaProperty(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Schema, err := service.GetDataSchemaWithUUID(Property.SchemaId)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if *Schema.Published {
		c.JSON(common.HTTP_OK, common.Error("Data Schema Already published"))
		return
	}
	err1 := service.DeleteIotSchemaProperty(uuid)
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 分页查找数据
*
 */
func IotSchemaPropertyDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	record, err := service.FindIotSchemaProperty(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	IotPropertyVo := IotPropertyVo{
		SchemaId:    record.SchemaId,
		UUID:        record.UUID,
		Label:       record.Label,
		Name:        record.Name,
		Description: record.Description,
		Type:        record.Type,
		Rw:          record.Rw,
		Unit:        record.Unit,
	}
	IoTPropertyRuleVo := IoTPropertyRuleVo{}
	if err0 := IoTPropertyRuleVo.ParseRuleFromModel(record.Rule); err0 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err0))
		return
	}
	IotPropertyVo.Rule = IoTPropertyRuleVo
	c.JSON(common.HTTP_OK, common.OkWithData(IotPropertyVo))
}

/*
*
* 列表
*
 */
func IotSchemaPropertyPageList(c *gin.Context, ruleEngine typex.Rhilex) {
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	schemaUuid, _ := c.GetQuery("schema_uuid")
	db := interdb.InterDb()
	tx := db.Scopes(service.Paginate(*pager))
	var count int64
	err1 := interdb.InterDb().Model(&model.MIotProperty{}).
		Where("schema_id=?", schemaUuid).Count(&count).Error
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	var records []model.MIotProperty
	result := tx.Model(&model.MIotProperty{}).
		Where("schema_id=?", schemaUuid).Find(&records).Order("created_at DESC")
	if result.Error != nil {
		c.JSON(common.HTTP_OK, common.Error400(result.Error))
		return
	}
	recordsVoList := []IotPropertyVo{}
	for _, record := range records {
		IoTPropertyRuleVo := IoTPropertyRuleVo{}
		if err0 := IoTPropertyRuleVo.ParseRuleFromModel(record.Rule); err0 != nil {
			c.JSON(common.HTTP_OK, common.Error400(err0))
			return
		}
		IotPropertyVo := IotPropertyVo{
			SchemaId:    record.SchemaId,
			UUID:        record.UUID,
			Label:       record.Label,
			Name:        record.Name,
			Description: record.Description,
			Type:        record.Type,
			Rw:          record.Rw,
			Unit:        record.Unit,
		}
		IotPropertyVo.Rule = IoTPropertyRuleVo
		recordsVoList = append(recordsVoList, IotPropertyVo)
	}
	Result := service.WrapPageResult(*pager, recordsVoList, count)
	c.JSON(common.HTTP_OK, common.OkWithData(Result))
}
