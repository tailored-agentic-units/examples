```sh
aws login
```

```sh
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "What is infrastructure as code? 300 words or less" \
  -stream
```

```sh
cat ./cmd/prompt-agent/config.bedrock.json | jq .
```

```sh
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png" \
  -prompt "Describe what you see." \
  -stream
```
