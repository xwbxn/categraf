//go:build !windows
// +build !windows

package w_aviation

import (
	"fmt" //输出日志，用于DeBug

	"flashcat.cloud/categraf/config" //定义插件配置项
	"flashcat.cloud/categraf/inputs" //引入inputs对象聚合数据
	"flashcat.cloud/categraf/types"  //用于打包发送数据

	//gopsutil,获取硬件信息，跨平台系统
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/host"

	// dmidecode,获取linux信息
	"github.com/yumaojun03/dmidecode"
)

// 插件名称
const inputName = "w_aviation"

// 插件配置参数
type w_aviation struct {
	config.PluginConfig             //插件全局参数
	Instances           []*Instance `toml:"instances"` //插件自定义参数
}

// 插件的一些基础的函数，改名字即可
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &w_aviation{}
	})
}
func (wa *w_aviation) Init() error {
	if len(wa.Instances) == 0 {
		return types.ErrInstancesEmpty
	}
	for i := 0; i < len(wa.Instances); i++ {
		if err := wa.Instances[i].Init(); err != nil {
			return err
		}
	}

	return nil
}
func (wa *w_aviation) Clone() inputs.Input {
	return &w_aviation{}
}
func (wa *w_aviation) Name() string {
	return inputName
}
func (wa *w_aviation) GetInstances() []inputs.Instance {
	ret := make([]inputs.Instance, len(wa.Instances))
	for i := 0; i < len(wa.Instances); i++ {
		ret[i] = wa.Instances[i]
	}
	return ret
}

func (wa *w_aviation) Drop() {}

func (wa *w_aviation) Gather(slist *types.SampleList) {}

// 用于配置插件详细参数
type Instance struct {
	config.InstanceConfig
	HelloWorld string `toml:"HelloWorld"`
}

// 初始化
func (ins *Instance) Init() error {
	return nil
}

// **主要功能区**

func (ins *Instance) Gather(slist *types.SampleList) {
	// 模板
	// fields := map[string]interface{}{}
	// tags := map[string]string{}
	// slist.PushSamples(inputName, fields, tags)

	// CPU
	GetCpuInfo(slist)
	// MEM
	GetMemInfo(slist)
	// NET
	GetNetInfoS(slist)
	// DISK
	/*
		见disk*探针
	*/
	// BaseBoard
	GetBaseBoardInfo(slist)
	// BIOS
	GetBIOSInfo(slist)
}

// func for get CPU info
func GetCpuInfo(slist *types.SampleList) error {
	host_info, _ := host.Info() // cpu架构
	cpu_arch := host_info.KernelArch
	cpu_infos, _ := cpu.Info()
	if len(cpu_infos) == 0 { // 如果没有cpu就直接跳过不发送cpu信息
		return nil
	}
	Cpu_cores, _ := cpu.Counts(false)  //系统物理核心数
	Cpu_threads, _ := cpu.Counts(true) //系统虚拟核心数（线程数）
	// cpu睿频 未完成
	cpu_cores := fmt.Sprintf("%d", Cpu_cores)     // 系统cpu物理核心数
	cpu_threads := fmt.Sprintf("%d", Cpu_threads) // 系统cpu逻辑核心数

	for index, cpu_info := range cpu_infos {
		cpu_modelname := cpu_info.ModelName // cpu型号
		cpu_Mhz := fmt.Sprintf("%.0f", cpu_info.Mhz)

		fields := map[string]interface{}{
			"Cpu": 1,
		}
		tags := map[string]string{
			"cpu_index":   fmt.Sprintf("%d", index), // cpu序号
			"cpu_model":   cpu_modelname,            // cpu型号
			"cpu_arch":    cpu_arch,                 // cpu架构
			"cpu_Mhz":     cpu_Mhz,                  // 主频
			"cpu_cores":   cpu_cores,                // 系统总物理核心数
			"cpu_threads": cpu_threads,              // 系统总逻辑核心数
		}

		slist.PushSamples(inputName, fields, tags) // cpu主频
	}
	return nil
}

// func for get MEM info
func GetMemInfo(slist *types.SampleList) error {
	dmi, error := dmidecode.New()
	if error != nil {
		return error
	}
	MemInfo, error := dmi.MemoryDevice()
	if error != nil {
		return error
	}
	// mem_num := fmt.Sprint(len(MemInfo))

	for i, v := range MemInfo {
		fields := map[string]interface{}{
			"Mem": 1,
		}
		tags := map[string]string{
			"mem_index":      fmt.Sprint(i),
			"mem_total":      fmt.Sprintf("%d", v.Size),                // 内存大小
			"mem_mf":         v.Manufacturer,                           // 内存品牌
			"mem_type":       fmt.Sprint(v.Type),                       // 内存类型
			"mem_speed_conf": fmt.Sprint(v.ConfiguredMemoryClockSpeed), // 内存主频
			"mem_speed":      fmt.Sprint(v.Speed),                      // 内存主频
			// "mem_num":        mem_num,                                  // 物理插槽数量
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

// func for get NET info
func GetNetInfoS(slist *types.SampleList) error {
	netinfos := GetNetInfo()
	if len(netinfos) == 0 {
		return nil
	}
	for _, net := range netinfos {
		fields := map[string]interface{}{
			"Net": 1,
		}
		tags := map[string]string{
			"Name":    net["name"],
			"ipv4_IP": net["ipv4_IP"],
			"ipv6_IP": net["ipv6_IP"],
			"mac":     net["mac"],
			"gateway": net["gateway"],
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

// func for get BaseBoard info
func GetBaseBoardInfo(slist *types.SampleList) error {
	dmi, error := dmidecode.New()
	if error != nil {
		return error
	}
	BBInfo, error := dmi.BaseBoard()
	if error != nil {
		return error
	}
	for _, v := range BBInfo {
		fields := map[string]interface{}{
			"BaseBoard": 1, // 主板
		}
		tags := map[string]string{
			"BB_Manufacturer": v.Manufacturer, // 厂商
			"BB_SerialNumber": v.SerialNumber, // 序列号
			"BB_Product":      v.ProductName,  // 版本
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

// func for get NET info
func GetBIOSInfo(slist *types.SampleList) error {
	dmi, error := dmidecode.New()
	if error != nil {
		return error
	}
	BIInfo, error := dmi.BIOS()
	if error != nil {
		return error
	}
	for _, v := range BIInfo {
		fields := map[string]interface{}{
			"BIOS": 1, // BIOS
		}
		tags := map[string]string{
			"BI_Vendor":      v.Vendor,      // 厂商
			"BI_BIOSVersion": v.BIOSVersion, // 版本
			"BI_ReleaseDate": v.ReleaseDate, // 时间
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}
