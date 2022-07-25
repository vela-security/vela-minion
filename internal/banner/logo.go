//go:build !windows
package banner

const logo = "\u001B[1;33m" +
	"   ______________  _____\n" +
	"  / ___/ ___/ __ \\/ ___/\n" +
	" (__  |__  ) /_/ / /__  \n" +
	"/____/____/\\____/\\___/  \u001B[0m  \u001B[1;35mMINION\u001B[0m\n" +
	"Powered By: 东方财富安全团队\n\n" +
	"\t进程 PID: \u001B[1;1m%d\u001B[0m\n" +
	"\t操作系统: \u001B[1;1m%s\u001B[0m\n" +
	"\t系统架构: \u001B[1;1m%s\u001B[0m\n" +
	"\t软件版本: \u001B[1;1m%s\u001B[0m\n" +
	"\t编译时间: \u001B[1;1m%s\u001B[0m\n" +
	"\t修订版本: \u001B[1;1m%s\u001B[0m\n\n\n"
