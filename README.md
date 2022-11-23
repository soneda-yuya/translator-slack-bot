# createing translator slack bot using apigateway * lambda * go

Since this is an in-house tool, our policy is to operate it in a single environment (staging).

## Deploy
```shell
make stage/${app_name}/apply OPT=-auto-approve
...
base_url = "https://xxxxx.execute-api.ap-northeast-1.amazonaws.com/${app_name}"
function_name = "slack-translator"
```

## Run App
```shell
curl -X ${METHOD} "https://xxxxx.execute-api.ap-northeast-1.amazonaws.com/${app_name}/exec"
```

## App
### Translator bot app

```shell
curl -X GET https://xxxxx.execute-api.ap-northeast-1.amazonaws.com/slack-translator/exec
```
