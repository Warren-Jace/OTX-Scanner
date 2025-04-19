# OTX Scanner (OTX 信息抓取工具)

本工具使用 Go 语言编写，用于从 AlienVault OTX (Open Threat Exchange) API 获取指定域名的相关信息，包括：

*   **关联域名 (Related Domains):** 基于 WHOIS 数据查找可能相关的其他域名。
*   **子域名 (Subdomains):** 基于被动 DNS (Passive DNS) 数据查找子域名。
*   **URL 列表 (URLs):** 获取与该域名相关的已知 URL。

## 功能特性

*   从标准输入逐行读取域名进行查询。
*   为每个输入的域名分别调用 OTX API 获取关联域名、子域名和 URL。
*   将获取到的数据分别保存到不同的文件中：
    *   关联域名: `<domain>_related_domain.csv` (CSV 格式)
    *   子域名: `<domain>_subdomains.txt` (每行一个子域名)
    *   URL 列表: `<domain>_urls.txt` (每行一个 URL，自动处理分页)
*   采用标准 Go 项目结构，易于维护和扩展。
*   使用共享的 `http.Client` 提高网络请求效率。
*   包含基本的日志输出，显示处理进度和错误信息。

## 项目结构
```bash
otx-scanner/
├── cmd/
│   └── otx-scanner/
│       └── main.go        # 程序入口和主逻辑
├── internal/
│   └── otx/
│       ├── client.go      # OTX API 客户端和请求逻辑
│       └── processor.go   # JSON 处理和文件写入逻辑
├── go.mod                 # Go 模块文件
└── README.md              # 项目说明文件
```


## 环境要求

*   Go 1.16 或更高版本

## 安装与构建

1.  **克隆仓库 (如果项目在 Git 仓库中):**
    ```bash
    git clone <repository-url>
    cd otx-scanner
    ```
    或者，如果你只有这些代码文件，请确保它们位于 `otx-scanner` 目录下，并按照上面的结构组织好。

2.  **构建可执行文件:**
    在 `otx-scanner` 根目录下运行：
    ```bash
    go build -o otx-scanner_app ./cmd/otx-scanner/
    ```
    这将在根目录下生成一个名为 `otx-scanner_app` (Linux/macOS) 或 `otx-scanner_app.exe` (Windows) 的可执行文件。

## 使用方法

1.  **运行程序:**
    ```bash
    ./otx-scanner_app
    ```
    或者在 Windows 上：
    ```bash
    .\otx-scanner_app.exe
    ```

2.  **输入域名:**
    程序启动后会提示输入域名。每行输入一个域名，然后按 Enter。

3.  **结束输入:**
    *   在 Linux 或 macOS 上，按 `Ctrl + D`。
    *   在 Windows 上，按 `Ctrl + Z` 然后按 `Enter`。

4.  **查看输出:**
    程序会为每个输入的域名在当前目录下生成对应的 `.csv` 和 `.txt` 文件。同时，控制台会输出处理日志。

**示例:**

```bash
./otx-scanner_app
请输入域名，每行一个，按 Ctrl+D (Unix/Linux) 或 Ctrl+Z 后回车 (Windows) 结束输入:
example.com
google.com
^D  # (或者 Ctrl+Z Enter)
=================== 输入结束，开始处理 ====================
正在处理域名: example.com
请求 https://otx.alienvault.com/otxapi/indicators/domain/whois/example.com | 状态码: 200
成功获取 example.com 的关联域名数据并写入文件。
请求 https://otx.alienvault.com/otxapi/indicators/domain/passive_dns/example.com | 状态码: 200
成功获取 example.com 的子域名数据并写入文件。
请求 https://otx.alienvault.com/otxapi/indicators/domain/url_list/example.com?limit=500&page=1 | 状态码: 200
成功获取 example.com 的URL列表数据并写入文件。
域名 example.com 处理完成。
-----------------------------------------------------
正在处理域名: google.com
请求 https://otx.alienvault.com/otxapi/indicators/domain/whois/google.com | 状态码: 200
成功获取 google.com 的关联域名数据并写入文件。
请求 https://otx.alienvault.com/otxapi/indicators/domain/passive_dns/google.com | 状态码: 200
成功获取 google.com 的子域名数据并写入文件。
请求 https://otx.alienvault.com/otxapi/indicators/domain/url_list/google.com?limit=500&page=1 | 状态码: 200
检测到域名 google.com 的 URL 列表有下一页，正在获取第 2 页...
请求 https://otx.alienvault.com/otxapi/indicators/domain/url_list/google.com?limit=500&page=2 | 状态码: 200
... (可能会有多页)
成功获取 google.com 的URL列表数据并写入文件。
域名 google.com 处理完成。
-----------------------------------------------------
=================== 所有处理已完成 ========================
之后，你会在目录下找到 example.com_related_domain.csv, example.com_subdomains.txt, example.com_urls.txt, google.com_related_domain.csv, google.com_subdomains.txt, google.com_urls.txt 等文件。

## 依赖
本项目仅使用 Go 标准库。
