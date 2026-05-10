# Newsy 插件编写指南

Newsy 插件系统允许你为没有 RSS 的网站编写自定义抓取脚本。插件可以是任何语言的可执行文件，通过 stdin/stdout JSON 与主程序通信。

---

## 默认目录

默认情况下，Newsy 会从 XDG 风格的用户目录读取配置和插件，而不是依赖仓库当前目录：

```text
~/.config/newsy/
├── config.yaml
└── plugins/
    └── my-scraper
```

数据和日志默认位于：

```text
~/.local/share/newsy/newsy.db
~/.cache/newsy/newsy.log
```

你可以用下面的命令查看当前程序实际使用的路径：

```bash
newsy --print-paths
```

如果不想放在默认插件目录，也可以在 `config.yaml` 中用 `plugin_path` 指向任意可执行文件。

---

## 编写插件

### 插件协议

可执行文件接收一个命令行参数并读取标准输入中的 JSON：

```
/path/to/plugin validate  < 标准输入  # 验证配置
/path/to/plugin fetch     < 标准输入  # 获取文章
```

标准输入的内容是 `config.yaml` 中该 source 的 `config` 部分序列化为 JSON。插件配置中以 `plugin_` 为前缀的键（如 `plugin_path`）是保留键，不会被传给插件。

### 通信格式

#### 配置输入

config.yaml 中有以下配置时：

```yaml
sources:
  - key: my-site
    provider_type: plugin
    name: 我的网站
    enabled: true
    config:
      plugin: my-scraper     # 指定默认插件目录中的插件名称
      url: https://example.com
      selector: "h2.title a"     # 插件自定义参数
```

插件启动时会收到标准输入：

```json
{"url":"https://example.com","selector":"h2.title a"}
```

#### validate 输出

验证配置是否合法。成功时 stdout 返回：

```json
{"valid": true}
```

失败时 stdout 返回：

```json
{"valid": false, "error": "url is required"}
```

超时：10 秒。

#### fetch 输出

获取文章列表。成功时 stdout 返回：

```json
{
  "articles": [
    {
      "external_id": "https://example.com/article-1",
      "title": "文章标题",
      "link": "https://example.com/article-1",
      "author": "作者（可选）",
      "summary": "摘要文字（可选）",
      "content": "<p>全文HTML（可选）</p>",
      "published_at": "2026-05-10T12:00:00Z（可选，RFC 3339格式，UTC）"
    }
  ]
}
```

| 字段 | 必须 | 说明 |
|------|------|------|
| `external_id` | 是 | 文章唯一标识（同一 source 内去重依据），推荐用文章链接 |
| `title` | 是 | 文章标题 |
| `link` | 是 | 文章链接 |
| `author` | 否 | 作者名 |
| `summary` | 否 | 文章摘要/简介 |
| `content` | 否 | 文章全文（HTML 文本） |
| `published_at` | 否 | 发布时间，RFC 3339 格式（如 `2026-05-10T12:00:00Z`），空值则显示当前时间 |

失败时 stdout 返回：

```json
{"error": "网络请求失败：连接超时"}
```

超时：30 秒。

#### 错误输出

有两种方式报告错误：

1. **优雅错误**：退出码 0，stdout 返回包含 `error` 字段的 JSON
2. **严重错误**：退出码非 0，错误信息写入 stderr

推荐使用优雅错误方式，更安全。

---

## 示例插件

### Bash 示例

```bash
#!/usr/bin/env bash
set -euo pipefail

action="$1"
config=$(cat)

if [ "$action" = "validate" ]; then
  echo '{"valid":true}'
  exit 0
fi

if [ "$action" = "fetch" ]; then
  url=$(echo "$config" | python3 -c "import sys,json;print(json.load(sys.stdin).get('url',''))")
  html=$(curl -sL --max-time 15 "$url")

  echo '{"articles":[{"external_id":"1","title":"示例文章","link":"https://example.com/1"}]}'
  exit 0
fi

echo "unknown action: $action" >&2
exit 1
```

### Python 示例

```python
#!/usr/bin/env python3
import json, sys, requests
from bs4 import BeautifulSoup

config = json.load(sys.stdin)
action = sys.argv[1]

if action == 'validate':
    if 'url' not in config:
        print(json.dumps({"valid": False, "error": "url is required"}))
    else:
        print(json.dumps({"valid": True}))
    sys.exit(0)

if action == 'fetch':
    resp = requests.get(config['url'], timeout=15)
    soup = BeautifulSoup(resp.text, 'html.parser')
    articles = []
    for item in soup.select(config.get('selector', 'article')):
        link_el = item.find('a')
        if not link_el:
            continue
        articles.append({
            "external_id": link_el.get('href'),
            "title": link_el.get_text(strip=True),
            "link": link_el.get('href'),
        })
    print(json.dumps({"articles": articles}))
    sys.exit(0)

print(json.dumps({"error": f"unknown action: {action}"}))
sys.exit(0)
```

---

## 测试插件

```bash
# 测试 validate
printf '{"url":"https://example.com"}' | ~/.config/newsy/plugins/my-scraper validate

# 测试 fetch
printf '{"url":"https://example.com"}' | ~/.config/newsy/plugins/my-scraper fetch
```

也可以使用你在 `plugin_path` 里配置的绝对路径直接测试。

---

## 注意事项

- **超时**：validate 最长 10 秒，fetch 最长 30 秒，超时进程会被 kill
- **去重**：主程序通过 `source_key + external_id` 去重，重复抓取会更新已有文章
- **不要用交互式命令**：插件运行期间没有终端交互，所有输入通过 stdin 传入
- **HTTP 请求**：建议在 fetch 中设置合理的超时（如 15 秒），避免触发 30 秒插件超时
- **插件路径覆盖**：如果不想把插件放在默认插件目录，也可以在 config 中指定 `plugin_path`：
  ```yaml
  config:
    plugin_path: /home/user/custom-scripts/my-scraper.py
  ```
