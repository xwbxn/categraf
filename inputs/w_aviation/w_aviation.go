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
	"github.com/shirou/gopsutil/mem"
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
	/*测试*/
	// host_info, _ := host.Info()
	// fmt.Println("host_info:")
	// fmt.Println(host_info.OS)

	// avg, _ := load.Avg()
	// fmt.Println(avg)
	// misc, _ := load.Misc()
	// fmt.Println(misc)

	// num, _ := cpu.Counts(true) // true为线程，false为核心
	// fmt.Println(num)
	// fmt.Println("cpuinfo:")
	// info, _ := cpu.Info() // true为线程，false为核心
	// fmt.Println(info)
	// fmt.Println("cpu:")
	// pr, _ := cpu.Percent(0.0, true) // true为线程，false为核心
	// fmt.Println(pr)
	// fmt.Println("mem:")
	// sm, _ := mem.SwapMemory()
	// fmt.Println("SwapMemory")
	// fmt.Println(sm)
	// vm, _ := mem.VirtualMemory()
	// fmt.Println("VirtualMemory")
	// fmt.Println(vm)

	return nil
}

// **主要功能区**

func (ins *Instance) Gather(slist *types.SampleList) {
	// 模板
	// fields := map[string]interface{}{}
	// tags := map[string]string{}
	// slist.PushSamples(inputName, fields, tags)

	// 前置准备
	cpu_info, _ := cpu.Info()
	host_info, _ := host.Info()
	Cpu_cores, _ := cpu.Counts(false)
	Cpu_threads, _ := cpu.Counts(true)

	//OS
	OS := host_info.OS
	if OS == "windows" {
		fmt.Println("好耶 Windows")
	}
	fmt.Println("平台：", host_info.PlatformFamily)

	// CPU
	cpu_modelname := cpu_info[0].ModelName          // cpu型号
	cpu_arch := host_info.KernelArch                // cpu架构
	cpu_Mhz := fmt.Sprintf("%.0f", cpu_info[0].Mhz) // cpu主频
	// cpu睿频 未完成
	cpu_cores := fmt.Sprintf("%d", Cpu_cores)     // cpu物理核心数
	cpu_threads := fmt.Sprintf("%d", Cpu_threads) // cpu逻辑核心数
	prs, _ := cpu.Percent(0.0, true)              // 各逻辑核心使用率

	for index, pr := range prs {

		fields := map[string]interface{}{
			"cpu_percent": pr,
		}
		tags := map[string]string{
			"cpu":         fmt.Sprintf("%d", index),
			"cpu_model":   cpu_modelname,
			"cpu_arch":    cpu_arch,
			"cpu_Mhz":     cpu_Mhz,
			"cpu_cores":   cpu_cores,
			"cpu_threads": cpu_threads,
		}

		slist.PushSamples(inputName, fields, tags)
	}
	// MEM
	mem_V, _ := mem.VirtualMemory()
	mem_S, _ := mem.SwapMemory()

	dmi, _ := dmidecode.New()

	MemInfo, _ := dmi.MemoryDevice()
	mem_num := fmt.Sprint(len(MemInfo))

	for i, v := range MemInfo {
		fields := map[string]interface{}{
			"mem_used":        mem_V.Used,        // 已用内存
			"mem_usedpercent": mem_V.UsedPercent, // 已用内存
			"mem_free":        mem_V.Free,        // 可用物理内存
			"mem_free_v":      mem_S.Free,        // 可用虚拟内存
		}
		tags := map[string]string{
			"mem_index":      fmt.Sprint(i),
			"mem_total":      fmt.Sprintf("%d", mem_V.Total),           // 总内存大小
			"mem_num":        mem_num,                                  // 物理插槽数量
			"mem_mf":         v.Manufacturer,                           // 内存品牌
			"mem_type":       fmt.Sprint(v.Type),                       // 内存类型
			"mem_speed_conf": fmt.Sprint(v.ConfiguredMemoryClockSpeed), // 内存主频
			"mem_speed":      fmt.Sprint(v.Speed),                      // 内存主频
		}
		slist.PushSamples(inputName, fields, tags)
	}

	// DISK
	/*
		见disk*探针
	*/

	// NET
	/*
		见net*探针
	*/

	// BaseBoard
	BBInfo, _ := dmi.BaseBoard()
	for _, v := range BBInfo {
		fields := map[string]interface{}{
			"BaseBoard": 0, // 主板

		}

		tags := map[string]string{
			"BB_Manufacturer": v.Manufacturer, // 厂商
			"BB_SerialNumber": v.SerialNumber, // 序列号
			"BB_Product":      v.ProductName,  // 版本
		}
		slist.PushSamples(inputName, fields, tags)
	}

	// BIOS 固件
	BIInfo, _ := dmi.BIOS()
	for _, v := range BIInfo {
		fields := map[string]interface{}{
			"BIOS": 0, // BIOS
		}
		tags := map[string]string{
			"BI_Vendor":      v.Vendor,      // 厂商
			"BI_BIOSVersion": v.BIOSVersion, // 版本
			"BI_ReleaseDate": v.ReleaseDate, // 时间
		}
		slist.PushSamples(inputName, fields, tags)
	}

}
