package otx

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// dealRelatedJson 处理关联域名 JSON 数据并写入 CSV 文件
// 注意：函数名小写，变为包内私有
func dealRelatedJson(domainInput string, jsonstr string) error {
	var jsons map[string]interface{}
	err := json.Unmarshal([]byte(jsonstr), &jsons)
	if err != nil {
		return fmt.Errorf("解析关联域名 JSON 失败: %w", err)
	}

	relatedData, ok := jsons["related"].([]interface{}) // OTX API 返回的是一个数组
	if !ok {
		// 有可能 related 字段不存在或类型不正确
		log.Printf("警告: 域名 %s 未找到 'related' 字段或格式不正确。\n", domainInput)
		return nil // 不视为错误，可能就是没有关联数据
	}

	if len(relatedData) == 0 {
		log.Printf("域名 %s 没有关联域名数据。\n", domainInput)
		return nil
	}

	fileName := domainInput + "_related_domain.csv"
	fileObj, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666) // O_TRUNC 确保每次覆盖
	if err != nil {
		return fmt.Errorf("打开或创建文件 %s 失败: %w", fileName, err)
	}
	defer fileObj.Close()

	wr := bufio.NewWriter(fileObj)
	_, err = wr.WriteString("domain,related,related_type\n") // 写入 CSV 头
	if err != nil {
		return fmt.Errorf("写入 CSV 头失败: %w", err)
	}

	for _, item := range relatedData {
		v, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("警告: 关联域名数据项格式不正确: %v\n", item)
			continue
		}

		// 安全地提取字段，处理可能的类型断言失败或字段缺失
		domain, _ := v["domain"].(string)
		related, _ := v["related"].(string)
		relatedType, _ := v["related_type"].(string)

		line := fmt.Sprintf("%s,%s,%s\n", domain, related, relatedType)
		_, err = wr.WriteString(line)
		if err != nil {
			// 记录错误但尝试继续处理其他行
			log.Printf("写入关联域名数据行失败 (%s): %v\n", line, err)
		}
	}

	err = wr.Flush() // 确保所有缓冲数据写入文件
	if err != nil {
		return fmt.Errorf("刷新文件缓冲区失败 (%s): %w", fileName, err)
	}
	return nil
}

// dealSubdomainsJson 处理子域名 JSON 数据并写入 TXT 文件
// 注意：函数名小写，变为包内私有
func dealSubdomainsJson(domainInput string, jsonstr string) error {
	var jsons map[string]interface{}
	err := json.Unmarshal([]byte(jsonstr), &jsons)
	if err != nil {
		return fmt.Errorf("解析子域名 JSON 失败: %w", err)
	}

	passiveDnsData, ok := jsons["passive_dns"].([]interface{})
	if !ok {
		log.Printf("警告: 域名 %s 未找到 'passive_dns' 字段或格式不正确。\n", domainInput)
		return nil
	}

	if len(passiveDnsData) == 0 {
		log.Printf("域名 %s 没有子域名数据。\n", domainInput)
		return nil
	}

	fileName := domainInput + "_subdomains.txt"
	fileObj, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666) // O_TRUNC 确保每次覆盖
	if err != nil {
		return fmt.Errorf("打开或创建文件 %s 失败: %w", fileName, err)
	}
	defer fileObj.Close()

	wr := bufio.NewWriter(fileObj)
	uniqueHostnames := make(map[string]bool) // 用于主机名去重

	for _, item := range passiveDnsData {
		v, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("警告: 子域名数据项格式不正确: %v\n", item)
			continue
		}

		hostname, okHostname := v["hostname"].(string)
		// address, okAddress := v["address"].(string) // IP 地址

		if okHostname && hostname != "" && !uniqueHostnames[hostname] {
			uniqueHostnames[hostname] = true // 标记为已添加
			_, err = wr.WriteString(hostname + "\n")
			if err != nil {
				log.Printf("写入子域名数据行失败 (%s): %v\n", hostname, err)
			}
			// 如果需要 IP 地址，可以在这里添加
			// if okAddress {
			//     _, err = wr.WriteString(fmt.Sprintf("%s,%s\n", hostname, address)) // 或者其他格式
			// }
		}
	}

	err = wr.Flush() // 确保所有缓冲数据写入文件
	if err != nil {
		return fmt.Errorf("刷新文件缓冲区失败 (%s): %w", fileName, err)
	}
	return nil
}

// dealUrlsJson 处理 URL 列表 JSON 数据并写入 TXT 文件
// 返回值: hasNext (bool), nextPageNum (int), error
// 注意：函数名小写，变为包内私有
func dealUrlsJson(domainInput string, jsonstr string) (bool, int, error) {
	var jsons map[string]interface{}
	err := json.Unmarshal([]byte(jsonstr), &jsons)
	if err != nil {
		return false, 0, fmt.Errorf("解析 URL 列表 JSON 失败: %w", err)
	}

	urlListData, ok := jsons["url_list"].([]interface{})
	if !ok {
		// OTX 在没有 URL 时可能不返回 url_list 字段
		log.Printf("信息: 域名 %s 在当前页未找到 'url_list' 字段。\n", domainInput)
		// 仍然需要检查分页
	}

	fileName := domainInput + "_urls.txt"
	// 使用 O_APPEND 模式，因为分页是递归调用的
	fileObj, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return false, 0, fmt.Errorf("打开或创建文件 %s 失败: %w", fileName, err)
	}
	defer fileObj.Close()

	wr := bufio.NewWriter(fileObj)

	if len(urlListData) > 0 {
		for _, item := range urlListData {
			v, ok := item.(map[string]interface{})
			if !ok {
				log.Printf("警告: URL 数据项格式不正确: %v\n", item)
				continue
			}

			url, okUrl := v["url"].(string)
			if okUrl && url != "" {
				_, err = wr.WriteString(url + "\n")
				if err != nil {
					log.Printf("写入 URL 数据行失败 (%s): %v\n", url, err)
				}
			}
		}
		err = wr.Flush() // 确保写入
		if err != nil {
			// 记录错误，但仍需检查分页
			log.Printf("刷新文件缓冲区失败 (%s): %v\n", fileName, err)
		}
	} else {
		log.Printf("信息: 域名 %s 在当前页没有 URL 数据。\n", domainInput)
	}

	// 检查分页
	hasNext, _ := jsons["has_next"].(bool)
	nextPageNum := 0
	if hasNext {
		// OTX API 返回的 page_num 是当前页码，需要加 1 得到下一页
		currentPageNumFloat, ok := jsons["page_num"].(float64)
		if !ok {
			return false, 0, fmt.Errorf("解析页码失败，'page_num' 字段类型不正确或缺失")
		}
		nextPageNum = int(currentPageNumFloat) + 1
	}

	return hasNext, nextPageNum, nil // 返回分页信息
}
