


```bash
> go mod download
go: github.com/bytedance/sonic@v1.9.1: Get "https://proxy.golang.org/github.com/bytedance/sonic/@v/v1.9.1.mod": dial tcp 142.250.198.81:443: connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.
```

**解决办法：**

尝试使用终端代理：
```bash
# linux/macOS
export https_proxy=http://127.0.0.1:13824 http_proxy=http://127.0.0.1:13824 all_proxy=socks5://127.0.0.1:13825

# Windows PowerShell
$env:https_proxy="http://127.0.0.1:13824"
$env:http_proxy="http://127.0.0.1:13824"
$env:all_proxy="socks5://127.0.0.1:13825"
```