package rockcli

import "fmt"

const (
	OpHeartbeat Opcode = iota
	OpSubstance
	OpThird
	OpReload
	OpOffline
	OpDeleted
	OpUpgrade
)

const (
	OpEvent Opcode = iota + 100
	OpCoreService
)

const (
	OpAccount Opcode = iota + 200
	OpCPU
	OpDiskIO
	OpFileSystem
	OpListen
	OpMemory
	OpNetwork
	OpProcess
	OpService
	OpSocket
	OpStat
)

var opcodeNames = map[Opcode]string{
	OpHeartbeat: "minion 发出的心跳包",
	OpSubstance: "minion 配置更新",
	OpThird:     "三方文件更新",
	OpReload:    "重新加载指定配置",
	OpOffline:   "节点下线",
	OpDeleted:   "删除节点",
	OpUpgrade:   "节点客户端升级",

	OpEvent:       "上报事件",
	OpCoreService: "上报 rock-go 内部服务运行信息",

	OpAccount:    "上报系统账户信息",
	OpCPU:        "上报 CPU 信息",
	OpDiskIO:     "上报磁盘 I/O",
	OpFileSystem: "上报文件系统",
	OpListen:     "上报端口监听",
	OpMemory:     "上报内存信息",
	OpNetwork:    "上报网络信息",
	OpProcess:    "上报进程信息",
	OpService:    "上报系统服务信息",
	OpSocket:     "上报 socket 连接信息",
	OpStat:       "上报节点基本信息",
}

// Opcode minion 节点操作码
type Opcode uint16

// String implement fmt.Stringer
func (o Opcode) String() string {
	if s, ok := opcodeNames[o]; ok {
		return s
	}
	return fmt.Sprintf("<unnamed minion opcode: %d>", o)
}
