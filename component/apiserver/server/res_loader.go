package server

import (
	"errors"
	"sync"

	"encoding/json"

	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 当资源重启加载的时候，内存里面的数据会丢失，需要重新从数据库加载规则到资源，建立绑定关联。
*
 */

// LoadNewestInEnd
func LoadNewestInEnd(uuid string, ruleEngine typex.Rhilex) error {
	mInEnd, _ := service.GetMInEndWithUUID(uuid)
	if mInEnd == nil {
		return errors.New("Inend not exists:" + uuid)
	}
	config := map[string]any{}
	if err1 := json.Unmarshal([]byte(mInEnd.Config), &config); err1 != nil {
		glogger.GLogger.Error(err1)
		return err1
	}
	// 所有的更新都先停止资源,然后再加载
	old := ruleEngine.GetInEnd(uuid)
	if old != nil {
		if old.Source.Status() == typex.SOURCE_UP {
			old.Source.Stop()
		}
	}
	ruleEngine.RemoveInEnd(uuid)
	in := typex.NewInEnd(typex.InEndType(mInEnd.Type),
		mInEnd.Name, mInEnd.Description, mInEnd.GetConfig())
	// Important !!!!!!!! in.Id = mInEnd.UUID
	in.UUID = mInEnd.UUID
	BindRules := map[string]typex.Rule{}
	for _, ruleId := range mInEnd.BindRules {
		if ruleId == "" {
			continue
		}
		mRule, err1 := service.GetMRuleWithUUID(ruleId)
		if err1 != nil {
			return err1
		}
		glogger.GLogger.Debugf("Load rule:(%s,%s)", mRule.UUID, mRule.Name)
		RuleInstance := typex.NewLuaRule(
			ruleEngine,
			mRule.UUID,
			mRule.Name,
			mRule.Description,
			mRule.SourceId,
			mRule.DeviceId,
			mRule.Success,
			mRule.Actions,
			mRule.Failed)
		BindRules[mRule.UUID] = *RuleInstance
	}
	// 最新的规则
	in.BindRules = BindRules
	// 最新的配置
	in.Config = mInEnd.GetConfig()
	ctx, cancelCTX := typex.NewCCTX()
	if err2 := ruleEngine.LoadInEndWithCtx(in, ctx, cancelCTX); err2 != nil {
		glogger.GLogger.Error(err2)
		// return err2
	}
	go StartInSupervisor(ctx, in, ruleEngine)
	return nil
}

// LoadNewestOutEnd
func LoadNewestOutEnd(uuid string, ruleEngine typex.Rhilex) error {
	mOutEnd, err := service.GetMOutEndWithUUID(uuid)
	if err != nil {
		return err
	}

	config := map[string]any{}
	if err := json.Unmarshal([]byte(mOutEnd.Config), &config); err != nil {
		return err
	}
	// 所有的更新都先停止资源,然后再加载
	old := ruleEngine.GetOutEnd(uuid)
	if old != nil {
		old.Target.Stop()
	}
	ruleEngine.RemoveOutEnd(uuid)
	out := typex.NewOutEnd(typex.TargetType(mOutEnd.Type),
		mOutEnd.Name, mOutEnd.Description, config)
	// Important !!!!!!!!
	out.UUID = mOutEnd.UUID
	out.Config = mOutEnd.GetConfig()
	ctx, cancelCTX := typex.NewCCTX()
	if err := ruleEngine.LoadOutEndWithCtx(out, ctx, cancelCTX); err != nil {
		glogger.GLogger.Error(err)
	}
	go StartOutSupervisor(ctx, out, ruleEngine)
	return nil

}

/*
*
* 当资源重启加载的时候，内存里面的数据会丢失，需要重新从数据库加载规则到资源，建立绑定关联。
*
 */
var loadDeviceLocker = sync.Mutex{}

// LoadNewestDevice
func LoadNewestDevice(uuid string, ruleEngine typex.Rhilex) error {
	loadDeviceLocker.Lock()
	defer loadDeviceLocker.Unlock()
	mDevice, err := service.GetMDeviceWithUUID(uuid)
	if err != nil {
		return err
	}
	config := map[string]any{}
	if err := json.Unmarshal([]byte(mDevice.Config), &config); err != nil {
		return err
	}
	// 所有的更新都先停止资源,然后再加载
	old := ruleEngine.GetDevice(uuid)
	if old != nil {
		old.Device.Stop()
	}
	ruleEngine.RemoveDevice(uuid) // 删除内存里面的
	dev := typex.NewDevice(typex.DeviceType(mDevice.Type), mDevice.Name,
		mDevice.Description, mDevice.GetConfig())
	// Important !!!!!!!!
	dev.UUID = mDevice.UUID // 本质上是配置和内存的数据映射起来
	BindRules := map[string]typex.Rule{}
	for _, ruleId := range mDevice.BindRules {
		mRule, err1 := service.GetMRuleWithUUID(ruleId)
		if err1 != nil {
			return err1
		}
		glogger.GLogger.Debugf("Load rule:(%s,%s)", mRule.UUID, mRule.Name)
		RuleInstance := typex.NewLuaRule(
			ruleEngine,
			mRule.UUID,
			mRule.Name,
			mRule.Description,
			mRule.SourceId,
			mRule.DeviceId,
			mRule.Success,
			mRule.Actions,
			mRule.Failed)
		BindRules[mRule.UUID] = *RuleInstance
	}
	// 最新的规则
	dev.BindRules = BindRules
	// 最新的配置
	dev.Config = mDevice.GetConfig()
	// 参数传给 --> startDevice()
	ctx, cancelCTX := typex.NewCCTX()
	err2 := ruleEngine.LoadDeviceWithCtx(dev, ctx, cancelCTX)
	if err2 != nil {
		glogger.GLogger.Error(err2)
		// return err2
	}
	go StartDeviceSupervisor(ctx, dev, ruleEngine)
	return nil

}
