package w_aviation

import (
	"context"
	"fmt" //输出日志，用于DeBug
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"flashcat.cloud/categraf/config" //定义插件配置项
	"flashcat.cloud/categraf/inputs" //引入inputs对象聚合数据

	//自定义cpu组件, 修改自psutil
	"flashcat.cloud/categraf/types" //用于打包发送数据

	//gopsutil,获取硬件信息，跨平台系统
	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"

	// "github.com/shirou/gopsutil/v3/net"
	"github.com/toolkits/pkg/logger"

	// 获取范式电源信息
	"github.com/distatus/battery"

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
	ImpiTimeout config.Duration `toml:"ipmi_timeout"`
	Path        string          `toml:"path"`
}

// 初始化
func (ins *Instance) Init() error {
	if ins.ImpiTimeout == 0 {
		ins.ImpiTimeout = config.Duration(time.Second * 5)
	}
	return nil
}

// **主要功能区**

func (ins *Instance) Gather(slist *types.SampleList) {
	// 模板
	// fields := map[string]interface{}{}
	// tags := map[string]string{}
	// slist.PushSamples(inputName, fields, tags)

	// CPU
	ins.GetCpuInfo(slist)
	// MEM
	ins.GetMemInfo(slist)
	// NET
	ins.GetNetInfo(slist)
	// DISK
	ins.GetDiskInfo(slist)
	// BaseBoard
	ins.GetBaseBoardInfo(slist)
	// BIOS
	ins.GetBIOSInfo(slist)
	// OS
	ins.GetOSInfo(slist)
	// BUS
	ins.GetBusInfo(slist)
	// Battery
	ins.GetBatteryInfo(slist)
	// Power Supply
	ins.GetPowerInfo(slist)
}

// func for get CPU info
func (ins *Instance) GetCpuInfo(slist *types.SampleList) error {
	dmi, error := dmidecode.New()
	if error != nil {
		return error
	}
	processors, error := dmi.Processor()
	if error != nil {
		return error
	}

	host_info, _ := host.Info() // cpu架构
	cpu_arch := host_info.KernelArch

	for index, cpu_info := range processors {
		fields := map[string]interface{}{
			"cpu": 1,
		}
		tags := map[string]string{
			"index":         fmt.Sprintf("%d", index),                 // cpu序号
			"model":         cpu_info.Version,                         // cpu型号
			"arch":          cpu_arch,                                 // cpu架构
			"frequency":     fmt.Sprintf("%d", cpu_info.CurrentSpeed), // 主频
			"core_count":    fmt.Sprintf("%d", cpu_info.CoreCount),    // 系统总物理核心数
			"thread_count":  fmt.Sprintf("%d", cpu_info.ThreadCount),  // 系统总逻辑核心数
			"max_frequency": fmt.Sprintf("%d", cpu_info.MaxSpeed),     //最大主频
		}
		slist.PushSamples(inputName, fields, tags) // cpu主频
	}
	return nil
}

// func for get MEM info
func (ins *Instance) GetMemInfo(slist *types.SampleList) error {
	dmi, error := dmidecode.New()
	if error != nil {
		return error
	}
	memDevice, error := dmi.MemoryDevice()
	if error != nil {
		return error
	}
	memArray, error := dmi.MemoryArray()
	if error != nil {
		return error
	}
	num_device := 0
	if len(memArray) > 0 {
		num_device = int(memArray[0].NumberOfMemoryDevices)
	}

	for i, v := range memDevice {
		fields := map[string]interface{}{
			"memory": 1,
		}
		tags := map[string]string{
			"index":      fmt.Sprint(i),
			"capacity":   fmt.Sprintf("%d", v.Size),                // 内存大小
			"brand":      v.Manufacturer,                           // 内存品牌
			"type":       fmt.Sprint(v.Type),                       // 内存类型
			"frequency":  fmt.Sprint(v.ConfiguredMemoryClockSpeed), // 内存主频
			"num_device": fmt.Sprintf("%d", num_device),            // 物理插槽数量
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

func (ins *Instance) GetDiskInfo(slist *types.SampleList) error {
	diskinfos, err := disk.Partitions(false)
	if err != nil {
		return err
	}
	for _, info := range diskinfos {
		fields := map[string]interface{}{
			"disk": 1,
		}
		capacity := 0
		usage, err := disk.Usage(info.Mountpoint)
		if err == nil {
			capacity = int(usage.Total)
		}
		tags := map[string]string{
			"name":     info.Mountpoint,
			"type":     info.Fstype,
			"capacity": fmt.Sprintf("%d", capacity),
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

// func for get NET info
func (ins *Instance) GetNetInfo(slist *types.SampleList) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	netinfo, err := ghw.Network()
	if err != nil {
		fmt.Printf("Error getting network info: %v", err)
		return err
	}
	netType := map[string]string{}
	for _, nic := range netinfo.NICs {
		if nic.IsVirtual {
			netType[strings.ToUpper(nic.MacAddress)] = "虚拟网卡"
		} else {
			netType[strings.ToUpper(nic.MacAddress)] = "物理网卡"
		}
	}

	defaultgw := GetGateway()
outerloop:
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		var ipv4, ipv6, gateway string
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if v4 := ipnet.IP.To4(); v4 != nil {
					ipv4 = v4.String()
					gateway = defaultgw
				} else if v6 := ipnet.IP.To16(); v6 != nil {
					ipv6 = v6.String()
				}
			} else {
				continue outerloop
			}
		}

		fields := map[string]interface{}{
			"net": 1,
		}
		tags := map[string]string{
			"name":       iface.Name,
			"address":    ipv4,
			"address_v6": ipv6,
			"mac":        iface.HardwareAddr.String(),
			"gateway":    gateway,
			"type":       netType[strings.ToUpper(iface.HardwareAddr.String())],
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

// func for get BaseBoard info
func (ins *Instance) GetBaseBoardInfo(slist *types.SampleList) error {
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
			"board": 1, // 主板
		}
		tags := map[string]string{
			"manufacturers": v.Manufacturer, // 厂商
			"serial_num":    v.SerialNumber, // 序列号
			"version":       v.ProductName,  // 版本
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

// func for get NET info
func (ins *Instance) GetBIOSInfo(slist *types.SampleList) error {
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
			"bios": 1, // BIOS
		}
		tags := map[string]string{
			"manufacturers": v.Vendor,      // 厂商
			"version":       v.BIOSVersion, // 版本
			"release_date":  v.ReleaseDate, // 时间
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

func (ins *Instance) GetOSInfo(slist *types.SampleList) error {
	info, err := host.Info()
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"os": 1, // os
	}
	tags := map[string]string{
		"name":    info.Platform,                   // 名称
		"version": info.PlatformVersion,            // 版本
		"vendor":  info.PlatformFamily,             // 厂商
		"env":     strings.Join(os.Environ(), ";"), //环境变量
	}
	slist.PushSamples(inputName, fields, tags)
	return nil
}

func (ins *Instance) GetBatteryInfo(slist *types.SampleList) error {
	batteries, err := battery.GetAll()
	if err != nil {
		fmt.Println("Could not get battery info!")
		return err
	}

	for i, battery := range batteries {
		fields := map[string]interface{}{
			"power": 1,
		}
		tags := map[string]string{
			"index":  fmt.Sprint(i),
			"status": battery.State.String(),
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}

func (ins *Instance) GetPowerInfo(slist *types.SampleList) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ins.ImpiTimeout))
	defer cancel()
	cmd := exec.CommandContext(ctx, ins.Path, "sensor", "-s", "-g", "\"Power Supply\"")
	out, err := cmd.CombinedOutput()
	logger.Debug(string(out))
	if err != nil {
		return fmt.Errorf("failed to run command %q: %w - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		ipmiFields := ExtractFieldsFromRegex(line)
		if len(ipmiFields) == 0 {
			continue
		}

		val := float64(0)
		if ipmiFields["status"] == "OK" || ipmiFields["status"] == "Present" {
			val = 1
		}

		fields := map[string]interface{}{
			"power": val, // power
		}
		tags := map[string]string{
			"sensor": ipmiFields["name"],
			"type":   ipmiFields["sdr"],
		}
		slist.PushSamples(inputName, fields, tags)
	}

	return nil
}
