[English](README.md) | [中文](README.zh-CN.md)

# iframe-extractor

一个轻量级的 Go 工具，使用 headless Chrome 从网页中提取 iframe 标签。非常适合分析嵌入内容、抓取 iframe 源或审计第三方集成。

## 功能特性

- **Headless Browser**：通过 [chromedp](https://github.com/chromedp/chromedp) 使用 Chrome DevTools Protocol 完整加载页面，包括动态渲染的内容
- **并发处理**：使用可配置的 worker pools 并行处理多个 URL
- **智能等待**：随机睡眠间隔确保动态内容有足够时间加载
- **灵活输入**：通过命令行参数或输入文件接受 URL
- **去重**：自动过滤重复的 iframe 源
- **恢复支持**：跳过已处理的 URL，避免重复工作
- **可自定义输出**：配置输出目录，选择保存所有 iframe 或仅保存第一个

## 安装

### 前置要求

- Go 1.25.0 或更高版本
- 系统已安装 Chrome/Chromium 浏览器

### 从源码构建

```bash
git clone <repository-url>
cd iframe-extractor
go mod download
go build
```

## 使用方法

### 基本用法

从单个 URL 提取 iframe：

```bash
./iframe-extractor https://example.com
```

处理多个 URL：

```bash
./iframe-extractor https://example.com https://test.com https://site.org
```

### 使用 URL 文件

创建一个文本文件，每行一个 URL：

```text
# urls.txt
https://example.com
https://test.com

# 注释和空行会被忽略
https://another-site.org
```

处理该文件：

```bash
./iframe-extractor --urls-file urls.txt
```

### 配置选项

| 标志 | 默认值 | 说明 |
|------|---------|------|
| `--urls-file` | - | 包含 URL 的文件路径（每行一个） |
| `--skip-already-exists-files` | `true` | 跳过输出文件已存在的 URL |
| `--max-workers` | `3` | 并行处理的 worker 数量 |
| `--timeout-seconds` | `70` | 加载每个页面的浏览器超时时间（秒） |
| `--min-sleep-seconds` | `10` | 页面加载后等待内容渲染的最短时间 |
| `--max-sleep-seconds` | `25` | 页面加载后等待内容渲染的最长时间 |
| `--save-all-iframes` | `true` | 保存找到的所有 iframe（false = 仅保存第一个 iframe） |
| `--output-dir` | `__output_iframes` | 提取的 iframe HTML 文件保存目录 |

### 示例

仅提取每个页面的第一个 iframe：

```bash
./iframe-extractor --save-all-iframes=false https://example.com
```

使用 10 个并发 worker 和较短的超时时间处理：

```bash
./iframe-extractor --max-workers 10 --timeout-seconds 30 --urls-file urls.txt
```

为加载缓慢的动态内容自定义等待时间：

```bash
./iframe-extractor --min-sleep-seconds 15 --max-sleep-seconds 30 https://example.com
```

保存到自定义目录：

```bash
./iframe-extractor --output-dir my_iframes https://example.com
```

重新处理所有 URL（不跳过已存在的文件）：

```bash
./iframe-extractor --skip-already-exists-files=false --urls-file urls.txt
```

## 工作原理

1. **解析输入**：从命令行参数或输入文件读取 URL
2. **过滤 URL**：可选择跳过已处理的 URL
3. **并发处理**：启动 worker goroutine 并行处理 URL
4. **浏览器自动化**：对于每个 URL：
   - 启动一个 headless Chrome 实例
   - 导航到 URL
   - 等待页面 body 准备就绪
   - 睡眠一个随机时长（在最小/最大秒数之间），以允许动态内容加载
   - 使用正则表达式模式匹配从 HTML 中提取 iframe 标签
   - 过滤掉 `about:blank` iframe 和重复项
5. **保存结果**：创建包含提取的 iframe 标签的 HTML 文件
6. **报告**：显示进度和遇到的任何错误

## 输出格式

提取的 iframe 保存为简单的 HTML 文件：

```html
<!DOCTYPE html>
<html>
<body>
<iframe src="https://example.com/embed/video123">
<iframe src="https://another-site.com/widget">
</body>
</html>
```

### 输出文件名

文件名派生自 URL 的最后一个路径段：
- `https://example.com/page/video` → `video.html`
- `https://example.com/page/` → `page.html`
- `https://example.com/` → `index.html`
- 查询参数会被移除：`page?id=123` → `page.html`

## 依赖项

- [chromedp/chromedp](https://github.com/chromedp/chromedp) - Go 的 Chrome DevTools Protocol 驱动
