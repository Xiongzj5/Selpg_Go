package main

import (
	"bufio" // 用于从标准输入流获取数据和将数据写入到标准输入流 //
	"fmt"	// 用于引用io.EOF 来判断错误是否是文件尾导致 //
	"io"	// 用于将错误信息写入到标准错误流 //
	"log"	// 用于打开文件和异常退出时发送状态码 //
	"os"	// 用于打开文件和异常退出时发送状态码 //
	"os/exec"	// 用于开启lp子进程 //
	"strings"	// 用于划分、拼接字符串 //

	flag "github.com/spf13/pflag"	// 用于获取程序运行时用户输入的参数和标识 //
)

func main() {
	//初始化返回的都是对应类型的指针，绑定相应的变量以及赋相应的初始值
	//定义待解析的命令行参数
	startNumber := flag.IntP("startpage", "s", 0, "The page to start printing at [Necessary, no greater than endpage]")
	endNumber := flag.IntP("endpage", "e", 0, "The page to end printing at [Necessary, no less than startpage]")
	lineNumber := flag.IntP("linenumber", "1", 72, "If this flag is used, a page will consist of a fixed number of characters, which is given by you")
	forcePage := flag.BoolP("forcepaging", "f", false, "Change page only if '-f' appears [Cannot be used with -l]")
	destinationPrinter := flag.StringP("destination", "d", "", "Choose a printer to accept the result as a task")

	// StdErr printer //
	l := log.New(os.Stderr, "", 0)

	// Data holder //
	bytes := make([]byte,65535)
	var data string
	var resultData string

	flag.Parse() // 所有flag定义完成后用Parse来解析

	// Are necessary flags given? //
	if *startNumber == 0 || *endNumber == 0 {
		l.Println("Necessary falgs are not given!")
		flag.Usage() // 用于输出所有定义了的命令行参数和帮助信息. 一般，当命令行参数解析出错时，该函数会被调用。
		os.Exit(1)
	}

	// Are flags value valid //
	if(*startNumber > *endNumber) || *startNumber < 0 || *endNumber < 0 || *lineNumber <= 0 {
		l.Println("Invalid flag values!")
		flag.Usage()
		os.Exit(1)
	}

	// Are lineNumber and forcePage set at the same time? //
	if *lineNumber != 72 && *forcePage {
		l.Println("Linenumber and forcepaging cannot be set at the same time!")
		flag.Usage()
		os.Exit(1)
	}

	// Too many arguments? //
	if flag.NArg() > 1 {
		l.Println("Too many arguments!")
		flag.Usage()
		os.Exit(1)
	}

	// StdIn or File? //
	if flag.NArg() == 0 { // 参数中没有能够按照预定义的参数解析的部分，通过flag.Args()即可获取，是一个字符串切片
		// StdIn condition //
		reader := bufio.NewReader(os.Stdin)

		size, err := reader.Read(bytes)

		for size != 0 && err == nil {
			data = data + string(bytes)
			size, err = reader.Read(bytes)
		}

		// Error
		if err != io.EOF {
			l.Println("Error occured when reading from StdIn:\n", err.Error())
			os.Exit(1)
		}

	} else {
		// File condition //
		file, err := os.Open(flag.Args()[0]) // TODO TEST: is PATH needed?
		if err != nil {
			l.Println("Error occured when opening file:\n", err.Error())
			os.Exit(1)
		}

		// 读取整个文件
		size, err := file.Read(bytes)

		for size != 0 && err == nil {
			data = data + string(bytes)
			size, err = file.Read(bytes)
		}

		// Error
		if err != io.EOF {
			l.Println("Error occured when reading file:\n", err.Error())
			os.Exit(1)
		}
	}

	if *forcePage {  // 这里是看换页符的
		pagedData := strings.SplitAfter(data, "\f")

		if len(pagedData) < *endNumber {
			l.Println("Invalid flag values! Too large endNumber!")
			flag.Usage()
			os.Exit(1)
		}

		resultData = strings.Join(pagedData[*startNumber-1:*endNumber], "")
	}
	 else {	// 这里是看换行符的
		lines := strings.SplitAfter(data, "\n")
		if len(lines) < (*endNumber-1)*(*lineNumber)+1 {
			l.Println("Invalid flag values! Too large endNumber!")
			flag.Usage()
			os.Exit(1)
		}
		if len(lines) < *endNumber*(*lineNumber) {
			resultData = strings.Join(lines[(*startNumber)*(*lineNumber)-(*lineNumber):], "")
		} else {
			resultData = strings.Join(lines[(*startNumber)*(*lineNumber)-(*lineNumber):(*endNumber)*(*lineNumber)], "")
		}
	}

	writer := bufio.NewWriter(os.Stdout) // 创建一个Writer

	// StdOut or Printer? //
	if *destinationPrinter == "" {
		// StdOut //
		fmt.Printf("%s", resultData)
	} else {
		// Printer //
		cmd := exec.Command("lp", "-d"+*destinationPrinter)
		lpStdin, err := cmd.StdinPipe() // 连接到命令启动时标准输入的管道

		if err != nil {
			l.Println("Error occured when trying to send data to lp:\n", err.Error())
			os.Exit(1)
		}
		go func() {
			defer lpStdin.Close() // 在return之前关闭管道并输出相关内容
			io.WriteString(lpStdin, resultData)
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			l.Println("Error occured when sending data to lp:\n", err.Error())
			os.Exit(1)
		}

		_, err = writer.Write(out)

		if err != nil {
			l.Println("Error occured when writing information to StdOut:\n", err.Error())
			os.Exit(1)
		}
	}
}
