package command

import (
	"fmt"
	"github.com/byronzhu-haha/bitset"
	"github.com/byronzhu-haha/portscanner/logger"
	"github.com/byronzhu-haha/portscanner/scanner"
	"github.com/spf13/cobra"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const version = "development"

type Arguments struct {
	isPrintVersion *bool
	isPrint        *bool
	isWroteFile    *bool
	goroutines     *int32
	flushSec       *int32
	timeout        *int
	typ            *string
	filePath       *string
	fileName       *string
	ips            *string
	ports          *string
}

var (
	inArgs = &Arguments{
		isPrintVersion: new(bool),
		isPrint:        new(bool),
		isWroteFile:    new(bool),
		goroutines:     new(int32),
		flushSec:       new(int32),
		timeout:        new(int),
		typ:            new(string),
		filePath:       new(string),
		fileName:       new(string),
		ips:            new(string),
		ports:          new(string),
	}
	scannerCmd = &cobra.Command{
		Use:     "scanner",
		Short:   "a port scanner",
		Long:    "it is a port scanner that help you scan the target port status(open or close) of the target ip.",
		Version: version,
		Run:     run,
	}
	usage = `
Usage:

	./portscanner <command> [arguments]

The commands are:

	--help 输出帮助
	--host -i <target ip> 设置目标ip，e.g. -i baidu.com,192.168.1.1-100
	--ports -p <target port> 设置目标端口，e.g. -p 80,443,22,1-1000
	--version -v 输出程序当前版本号
	--verbose -V 输出详细日志
	--goroutines -g <number> 设置程序最大协程数
	--timeout -t <number> 单个端口扫描允许的超时时间，单位毫秒，默认3秒
	--scan-type -T <string> 端口扫描方式，目前支持的值有："connect"(全连接）
	--log -l 开启日志写入文件
	--filepath -P <string> 日志文件输出路径(绝对路径)，默认当前程序允许路径
	--filename -n <string> 日志文件名称，默认"port_scanner+当前纳秒级时间戳组成"
	--flush -f <number> 日志写入时间间隔，单位秒，默认30秒
`
)

func init() {
	scannerCmd.PersistentFlags().StringVarP(inArgs.ips, "host", "i", "", "设置目标ip，支持域名、0.0.0.0及0.0.0.0-255格式，e.g. -i baidu.com,192.168.1.1-100")
	scannerCmd.PersistentFlags().StringVarP(inArgs.ports, "ports", "p", "", "设置目标端口，支持一次设置多个，e.g. -p 80,443,22,1-1000")
	scannerCmd.PersistentFlags().BoolVarP(inArgs.isPrintVersion, "version", "v", false, "输出程序当前版本号")
	scannerCmd.PersistentFlags().BoolVarP(inArgs.isPrint, "verbose", "V", false, "输出详细日志")
	scannerCmd.PersistentFlags().BoolVarP(inArgs.isWroteFile, "log", "l", false, "开启日志写入文件")
	scannerCmd.PersistentFlags().Int32VarP(inArgs.goroutines, "goroutines", "g", 100, "设置程序最大协程数")
	scannerCmd.PersistentFlags().Int32VarP(inArgs.flushSec, "flush", "f", 60, "日志写入时间间隔，单位秒，默认30秒")
	scannerCmd.PersistentFlags().IntVarP(inArgs.timeout, "timeout", "t", 3000, "单个端口扫描允许的超时时间，单位毫秒，默认3秒")
	scannerCmd.PersistentFlags().StringVarP(inArgs.typ, "scan-type", "T", scanner.Connect, "端口扫描方式，目前支持的值有：\"connect\"(全连接）")
	scannerCmd.PersistentFlags().StringVarP(inArgs.filePath, "filepath", "P", "", "日志文件输出路径(绝对路径)，默认当前程序允许路径")
	scannerCmd.PersistentFlags().StringVarP(inArgs.fileName, "filename", "n", "", "日志文件名称，默认\"port_scanner+当前纳秒级时间戳组成\"")
}

func run(cmd *cobra.Command, args []string) {
	if *inArgs.isPrintVersion {
		println("scanner: ", version)
		return
	}
	scan := scanner.NewScanner(
		inArgs.parseIP(), inArgs.parsePort(), inArgs.parseLogger(),
		scanner.MaxGoroutines(inArgs.parseGoroutines()),
		scanner.Timeout(inArgs.parseTimeout()),
		scanner.TypeOfScanner(inArgs.parseScanType()),
	)
	scan.Start()
	defer scan.Stop()
	scan.Scan()
	scan.Output()
}

func Exec() {
	err := scannerCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func (a *Arguments) parseIP() []string {
	if *a.ips == "" {
		exit("未检测到目标ip的输入\n"+usage, 1)
	}
	_ips := strings.TrimSpace(*a.ips)
	ips := strings.Split(_ips, ",")
	if len(ips) <= 0 {
		exit("未检测到目标ip的输入", 1)
	}
	var (
		res []string
		set = bitset.NewBloomFilter()
	)
	addIP := func(ip string) {
		addr, err := net.ResolveIPAddr("ip", ip)
		if err != nil {
			exit("域名解析失败: "+err.Error(), 1)
			return
		}
		ip = addr.String()
		if net.ParseIP(ip) == nil {
			exit("ip格式不正确", 1)
		}
		if set.Has(ip) {
			return
		}
		set.Add(ip)
		res = append(res, ip)
	}
	for _, ip := range ips {
		if !strings.Contains(ip, "-") {
			addIP(ip)
			continue
		}
		tmp := strings.Split(ip, "-")
		if len(tmp) != 2 {
			continue
		}
		end, err := strconv.Atoi(tmp[1])
		if err != nil {
			continue
		}
		if end > 255 {
			end = 255
		}
		ip = tmp[0]
		idx := strings.LastIndex(ip, ".")
		if idx <= 0 || idx >= len(ip)-1 {
			continue
		}
		start, err := strconv.Atoi(ip[idx+1:])
		if err != nil {
			continue
		}
		if start <= 0 {
			start = 1
		}
		if start == end {
			addIP(ip)
			continue
		}
		if start > end {
			start, end = end, start
		}
		for i := start; i <= end; i++ {
			addIP(ip[:idx+1] + strconv.Itoa(i))
		}
	}

	return res
}

func (a *Arguments) parsePort() []int {
	if *a.ports == "" {
		exit("未检测到目标端口的输入\n"+usage, 1)
	}
	_ports := strings.TrimSpace(*a.ports)
	ports := strings.Split(_ports, ",")
	pLen := len(ports)
	if pLen <= 0 {
		exit("未检测到目标端口的输入", 1)
		return []int{}
	}
	var (
		set = bitset.NewBitSet()
		res []int
	)
	convertPort := func(port string) int {
		p, err := strconv.Atoi(port)
		if err != nil {
			exit("端口格式有误: "+err.Error(), 1)
		}
		return p
	}
	addPort := func(p int) {
		if p < 0 || p > 65535 {
			exit("端口需大于等于0, 小于等于65535", 1)
		}
		if set.Has(p) {
			return
		}
		set.Add(p)
		res = append(res, p)
	}
	for _, port := range ports {
		if !strings.Contains(port, "-") {
			addPort(convertPort(port))
			continue
		}
		tmp := strings.Split(port, "-")
		n := len(tmp)
		if n <= 0 || n > 2 {
			continue
		}
		start, end := 0, 0
		if n == 1 {
			end = convertPort(tmp[0])
		} else {
			start = convertPort(tmp[0])
			end = convertPort(tmp[1])
		}
		if start == end {
			addPort(start)
			continue
		}
		if start > end {
			start, end = end, start
		}
		for i := start; i <= end; i++ {
			addPort(i)
		}
	}
	return res
}

func (a *Arguments) parseGoroutines() int32 {
	if *a.goroutines <= 0 {
		return 100
	}
	return *a.goroutines
}

func (a *Arguments) parseTimeout() time.Duration {
	if *a.timeout <= 0 {
		return time.Second * 3
	}
	return time.Duration(*a.timeout) * time.Millisecond
}

func (a *Arguments) parseScanType() scanner.TypeScanner {
	switch *a.typ {
	case scanner.SYN, scanner.FIN:
		if os.Geteuid() > 0 {
			exit("您需要是特权用户才可执行此类扫描", 1)
		}
		return scanner.TypeScanner(*a.typ)
	case scanner.Connect, scanner.Device:
		return scanner.TypeScanner(*a.typ)
	default:
		return scanner.Connect
	}
}

func (a *Arguments) parseLogger() logger.Logger {
	var loggerOpts []logger.Option
	if *a.isPrint {
		loggerOpts = append(loggerOpts, logger.OpenPrint())
	}
	if !*a.isWroteFile {
		return logger.NewLogger(loggerOpts...)
	}
	filepath := *a.filePath
	if filepath == "" {
		filepath = "."
	}
	filename := *a.fileName
	if filename == "" {
		filename = fmt.Sprintf("port_scanner_%d.log", time.Now().UnixNano())
	} else if !strings.Contains(filename, ".") {
		filename += ".log"
	}
	flushSec := *a.flushSec
	if flushSec <= 0 {
		flushSec = 30
	}
	loggerOpts = append(loggerOpts, logger.OpenWriteFile())
	loggerOpts = append(loggerOpts, logger.LogFilePath(filepath))
	loggerOpts = append(loggerOpts, logger.LogFileName(filename))
	loggerOpts = append(loggerOpts, logger.FlushSec(flushSec))
	return logger.NewLogger(loggerOpts...)
}

func exit(msg string, code int) {
	println(msg)
	os.Exit(code)
}
