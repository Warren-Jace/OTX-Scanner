package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"otx-scanner/internal/otx" // 导入内部包
)

func main() {
	// 创建一个可复用的 HTTP 客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second, // 设置超时
	}

	// 创建 OTX 客户端实例
	otxClient := otx.NewClient(httpClient)

	log.Println("请输入域名，每行一个，按 Ctrl+D (Unix/Linux) 或 Ctrl+Z 后回车 (Windows) 结束输入:")

	inputReader := bufio.NewReader(os.Stdin)
	for {
		// 以\n为分隔符读取字符串
		domainInput, err := inputReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("=================== 输入结束，开始处理 ====================")
				break // 输入结束
			}
			log.Printf("读取输入时出错: %v\n", err)
			break // 发生其他错误，退出
		}

		// 清理输入
		domainInput = strings.TrimSpace(domainInput) // 去除首尾空格和换行符

		if domainInput == "" {
			continue // 跳过空行
		}

		log.Printf("正在处理域名: %s\n", domainInput)

		// 获得关联主域名
		err = otxClient.FetchRelatedDomains(domainInput)
		if err != nil {
			log.Printf("获取 %s 的关联域名失败: %v\n", domainInput, err)
		} else {
			log.Printf("成功获取 %s 的关联域名数据并写入文件。\n", domainInput)
		}

		// 获得子域名
		err = otxClient.FetchSubdomains(domainInput)
		if err != nil {
			log.Printf("获取 %s 的子域名失败: %v\n", domainInput, err)
		} else {
			log.Printf("成功获取 %s 的子域名数据并写入文件。\n", domainInput)
		}

		// 获得urls (从第一页开始)
		err = otxClient.FetchUrls(domainInput)
		if err != nil {
			log.Printf("获取 %s 的URL列表失败: %v\n", domainInput, err)
		} else {
			log.Printf("成功获取 %s 的URL列表数据并写入文件。\n", domainInput)
		}

		log.Printf("域名 %s 处理完成。\n", domainInput)
		log.Println("-----------------------------------------------------")

	}
	log.Println("=================== 所有处理已完成 ========================")
}
