//go:build windows
// +build windows

package w_aviation

import (
	"fmt" //输出日志，用于DeBug

	//探针自带的
	"flashcat.cloud/categraf/config" //定义插件配置项
	"flashcat.cloud/categraf/inputs" //引入inputs对象聚合数据
	"flashcat.cloud/categraf/types"  //用于打包发送数据

	//gopsutil,获取硬件信息，跨平台系统
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v3/host"

	//wmi,获取windows信息
	"github.com/StackExchange/wmi"
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

	// CPU
	cpu_modelname := cpu_info[0].ModelName          // cpu型号
	cpu_arch := host_info.KernelArch                // cpu架构
	cpu_Mhz := fmt.Sprintf("%.0f", cpu_info[0].Mhz) // cpu主频
	// cpu睿频 未完成
	cpu_cores := fmt.Sprintf("%d", Cpu_cores)     // cpu物理核心数
	cpu_threads := fmt.Sprintf("%d", Cpu_threads) // cpu逻辑核心数
	prs, _ := cpu.Percent(0.0, true)              // 各逻辑核心使用率

	for index, _ := range prs {

		fields_CPU := map[string]interface{}{
			"Cpu": 1,
		}
		tags_CPU := map[string]string{
			"cpu_index":   fmt.Sprintf("%d", index),
			"cpu_model":   cpu_modelname,
			"cpu_arch":    cpu_arch,
			"cpu_Mhz":     cpu_Mhz,
			"cpu_cores":   cpu_cores,
			"cpu_threads": cpu_threads,
		}

		slist.PushSamples(inputName, fields_CPU, tags_CPU)
	}
	// MEM
	mem_V, _ := mem.VirtualMemory()
	// mem_S, _ := mem.SwapMemory()

	fields_MEM := map[string]interface{}{
		"Mem": 1,
	}

	WinMemInfo := getWinMemInfo()
	if WinMemInfo != nil {
		mem_num := fmt.Sprint(len(WinMemInfo))
		for i, v := range WinMemInfo {
			tags_MEM := map[string]string{
				"mem_index":                  fmt.Sprint(i),
				"mem_total_" + fmt.Sprint(i): fmt.Sprintf("%d", mem_V.Total),     // 总内存大小
				"mem_num":                    mem_num,                            // 物理插槽数量
				"mem_mf":                     v.Manufacturer,                     // 内存品牌
				"mem_type":                   fmt.Sprint(v.MemoryType),           // 内存类型
				"mem_speed":                  fmt.Sprint(v.ConfiguredClockSpeed), // 内存主频
			}
			slist.PushSamples(inputName, fields_MEM, tags_MEM)
		}
	}

	// DISK
	/*
		直接使用disk探针

			磁盘数量 tag
			磁盘分区 tag
			总空间	tag
			可用空间 field
			磁盘IO	field
			挂载文件系统类型 tag
	*/

	// Net
	netinfos := GetNetInfo()
	for _, net := range netinfos {
		Net_fields := map[string]interface{}{
			"Net": 1,
		}
		Net_tags := map[string]string{
			"Name":    net["name"],
			"ipv4_IP": net["ipv4_ip"],
			"ipv6_IP": net["ipv6_ip"],
			"mac":     net["mac"],
			"gateway": net["gateway"],
		}
		slist.PushSamples(inputName, Net_fields, Net_tags)
	}

	// BaseBoard
	cmd := "wmic"
	// 厂商
	BB_Manufacturer_args := []string{"baseboard", "get", "Manufacturer"}
	BB_Manufacturer := output_command(cmd, BB_Manufacturer_args)
	// 序列号
	BB_SerialNumber_args := []string{"baseboard", "get", "SerialNumber"}
	BB_SerialNumber := output_command(cmd, BB_SerialNumber_args)
	// 版本
	BB_Product_args := []string{"baseboard", "get", "Product"}
	BB_Product := output_command(cmd, BB_Product_args)
	BB_fields := map[string]interface{}{
		"BaseBoard": 1,
	}
	BB_tags := map[string]string{
		"Manufacturer": BB_Manufacturer,
		"SerialNumber": BB_SerialNumber,
		"Product":      BB_Product,
	}
	slist.PushSamples(inputName, BB_fields, BB_tags)

	// BIOS
	// 厂商
	BI_Manufacturer_args := []string{"BIOS", "get", "Manufacturer"}
	BI_Manufacturer := output_command(cmd, BI_Manufacturer_args)
	// 版本
	BI_SMBIOSBIOSVersion_args := []string{"BIOS", "get", "SMBIOSBIOSVersion"}
	BI_SMBIOSBIOSVersion := output_command(cmd, BI_SMBIOSBIOSVersion_args)
	// 日期
	BI_ReleaseDate_args := []string{"BIOS", "get", "ReleaseDate"}
	BI_ReleaseDate := output_command(cmd, BI_ReleaseDate_args)
	BI_fields := map[string]interface{}{
		"BIOS": 1,
	}
	BI_tags := map[string]string{
		"Manufacturer":      BI_Manufacturer,
		"SMBIOSBIOSVersion": BI_SMBIOSBIOSVersion,
		"ReleaseDate":       BI_ReleaseDate,
	}
	slist.PushSamples(inputName, BI_fields, BI_tags)
}

type PhysicalMemory struct {
	Manufacturer         string // 品牌
	MemoryType           uint16 // 类型
	ConfiguredClockSpeed uint32 // 运行设置频率
	Speed                uint32 // 运行频率
}

// windwos 获取内存信息 详细信息请访问 https://learn.microsoft.com/zh-cn/windows/win32/cimwin32prov/win32-physicalmemory
func getWinMemInfo() []PhysicalMemory {
	var physicalMemory []PhysicalMemory
	err := wmi.Query("Select * from Win32_PhysicalMemory", &physicalMemory)
	if err != nil {
		return nil
	}
	return physicalMemory
}
