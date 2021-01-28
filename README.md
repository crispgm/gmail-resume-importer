# Gmail Resume Importer

从 Gmail 中处理并且下载简历。

目前不能通过配置，需要简单改下代码。

## 配置

1. 创建一个 Label: `job-resumes`, 把 boss 直聘邮件规则打这个标签
2. 在` job-resumes` 下创建 Label: `bot-downloaded`，用于标识已经完成下载
3. 运行程序，第一次需要 OAuth 2.0 登录
4. 先运行 `resume-import -show-labels` 把 `bot-downloaded` label 配置到代码中 `readLabel` 变量中
5. 程序会自动读取 `job-resumes` 中的邮件，完成下载并且标记已经下载

## 安装

1. Clone the repo
2. Build with: `go build -o resume-import`
3. Run `fetch.sh`

## Usage

Show labels:

```sh
resume-import -show-labels
```

Get resumes:

```sh
resume-import main.go
```
