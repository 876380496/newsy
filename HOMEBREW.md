# Homebrew 安装说明

## 目录约定

安装后的 `newsy` 默认使用 XDG 风格目录：

- 配置：`~/.config/newsy/config.yaml`
- 插件：`~/.config/newsy/plugins/`
- 数据库：`~/.local/share/newsy/newsy.db`
- 日志：`~/.cache/newsy/newsy.log`

查看实际路径：

```bash
newsy --print-paths
```

## 本地测试 formula

仓库根目录下保留了一份 `newsy.rb`，方便本地直接测试：

```bash
brew install --build-from-source ./newsy.rb
```

## 发布到 tap 仓库

推荐仓库结构：

- 主仓库：`876380496/newsy`
- tap 仓库：`876380496/homebrew-tap`

在 tap 仓库里放置：

```text
Formula/newsy.rb
```

本仓库也提供了同内容的标准路径版本：

```text
Formula/newsy.rb
```

用户安装：

```bash
brew tap 876380496/tap
brew install newsy
```

## formula 发布前要替换的字段

`Formula/newsy.rb` 中这几项需要替换成真实值：

- `sha256`
- `license`

## 生成 release tarball 的 sha256

假设你发布的是 `v0.1.0`：

```bash
curl -L -o newsy-v0.1.0.tar.gz https://github.com/876380496/newsy/archive/refs/tags/v0.1.0.tar.gz
shasum -a 256 newsy-v0.1.0.tar.gz
```

把输出的 hash 填进：

```ruby
sha256 "这里替换成真实值"
```

## 发布流程建议

1. 在主仓库打 tag，例如：
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
2. 创建 GitHub Release
3. 下载发布 tarball 并计算 sha256：
   ```bash
   curl -L -o newsy-v0.1.0.tar.gz https://github.com/876380496/newsy/archive/refs/tags/v0.1.0.tar.gz
   shasum -a 256 newsy-v0.1.0.tar.gz
   ```
4. 更新 tap 仓库中的 `Formula/newsy.rb`
5. 提交并推送 tap 仓库：
   ```bash
   git add Formula/newsy.rb
   git commit -m "Add newsy v0.1.0"
   git push origin main
   ```
6. 本地验证安装：
   ```bash
   brew update
   brew tap 876380496/tap
   brew install newsy
   newsy --print-paths
   ```
7. 如需升级测试：
   ```bash
   brew upgrade newsy
   ```

## 首次运行

首次运行会自动生成默认配置文件；如需插件，可放到 `~/.config/newsy/plugins/`，或在配置中使用 `plugin_path` 指向任意可执行文件。
