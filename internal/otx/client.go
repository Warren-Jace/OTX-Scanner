package otx

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// Client 结构用于封装 OTX API 客户端
type Client struct {
	httpClient *http.Client
}

// NewClient 创建一个新的 OTX API 客户端实例
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
	}
}

// 公共请求头
func setCommonHeaders(request *http.Request) {
	request.Header.Set("Host", "otx.alienvault.com")
	request.Header.Set("Connection", "keep-alive") // 使用 keep-alive 提高效率
	request.Header.Set("Accept", "*/*")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36") // 更新 User-Agent
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
}

// fetchAndProcess 执行通用的 GET 请求和处理流程
func (c *Client) fetchAndProcess(url string, domainInput string, processor func(string, string) error) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	setCommonHeaders(request)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("执行请求失败: %w", err)
	}
	defer response.Body.Close()

	log.Printf("请求 %s | 状态码: %d\n", url, response.StatusCode)

	if response.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("读取响应体失败: %w", err)
		}
		return processor(domainInput, string(body)) // 调用具体的处理函数
	} else if response.StatusCode == http.StatusNotFound {
		log.Printf("警告: 域名 %s 在 OTX 未找到相关数据 (%s)。\n", domainInput, url)
		return nil // 404 通常表示没有数据，不视为严重错误
	} else {
		bodyBytes, _ := ioutil.ReadAll(response.Body) // 尝试读取错误信息
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", response.StatusCode, string(bodyBytes))
	}
}

// FetchRelatedDomains 获取关联域名
func (c *Client) FetchRelatedDomains(domainInput string) error {
	url := "https://otx.alienvault.com/otxapi/indicators/domain/whois/" + domainInput
	return c.fetchAndProcess(url, domainInput, dealRelatedJson)
}

// FetchSubdomains 获取子域名
func (c *Client) FetchSubdomains(domainInput string) error {
	url := "https://otx.alienvault.com/otxapi/indicators/domain/passive_dns/" + domainInput
	return c.fetchAndProcess(url, domainInput, dealSubdomainsJson)
}

// FetchUrls 获取 URL 列表（处理分页）
func (c *Client) FetchUrls(domainInput string) error {
	return c.fetchUrlsPage(domainInput, "1") // 从第一页开始
}

// fetchUrlsPage 递归获取 URL 列表的特定页面
func (c *Client) fetchUrlsPage(domainInput string, page string) error {
	url := fmt.Sprintf("https://otx.alienvault.com/otxapi/indicators/domain/url_list/%s?limit=500&page=%s", domainInput, page)

	// 使用闭包来处理分页逻辑
	processor := func(domain string, body string) error {
		hasNext, nextPageNum, err := dealUrlsJson(domain, body) // dealUrlsJson 现在返回是否需要下一页和下一页的页码
		if err != nil {
			return err // 处理 JSON 或写入文件时出错
		}
		if hasNext {
			log.Printf("检测到域名 %s 的 URL 列表有下一页，正在获取第 %d 页...\n", domain, nextPageNum)
			// 递归调用获取下一页
			return c.fetchUrlsPage(domain, strconv.Itoa(nextPageNum))
		}
		return nil // 当前页面处理完成，且没有下一页
	}

	return c.fetchAndProcess(url, domainInput, processor)
}
