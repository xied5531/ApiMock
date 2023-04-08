# ApiMock

REST接口打桩模拟服务，支持自定义接口请求和动态响应。

# 运行

`ApiMock [mock_data.yaml] [api_mock.log]`

- 参数1：服务接口配置数据，可选，默认值：`mock_data.yaml`
- 参数2：日志记录问题，可选，默认值：当前目录下`gin-xxx.log`，默认保留最近10份

# 配置

## 模型

```yaml
name: "mock_server"                          # 必选，服务名
address: ":5555"                             # 必选，服务地址，例如“127.0.0.1:5555”
read_timeout_s: 5                            # 可选，读超时配置，默认5，单位秒
write_timeout_s: 10                          # 可选，写超时配置，默认10，单位秒
cert_file: "test.crt"                        # 可选，服务证书配置，未配置时，协议为http，否则为https
key_file: "test.key"                         # 可选，服务私钥配置，未配置时，协议为http，否则为https
apis:                                        # 固定，接口列表说明
  - request:                                 # 固定，请求体说明
      url: "/core/:id"                       # 必选，请求URL，路径参数以:开始，例如:id
      method: "POST"                         # 必选，标准请求方法，例如GET/POST/PUT/DELETE/HEAD
      metadata:                              # 可选，元数据说明，用于响应体自动获取变量内容，获取语法go template，标识符<<>>
        path_vars:                           # 可选，路径参数列表的说明
          - "id"                             
        header_keys:                         # 可选，请求头参数列表的说明
          - "h1"
        query_params:                        # 可选，请求参数列表的说明
          - "p1"
        json_body_keys:                      # 可选，请求体关键字列表的说明，仅支持JSON格式的请求体
          - "b1"
          - "b2.k"
    response:                                # 固定，响应体说明
      status: 200                            # 必选，响应码
      content_type: "application/json"       # 可选，响应结构，默认值application/json
      body: '{"key1":"value1", "path_id": <<.path_vars.id>>, "body_b1": <<.json_body_keys.b1>>, "body_b2_k": <<.json_body_keys.b2.k>>}'  # 可选，响应体，支持go template自动变量替换
      headers:                               # 可选，响应头说明
        - key: "Header1"                     # 响应头，Key，首字符自动大写
          value: "value1"                    # 响应头，Value，支持go template自动变量替换
        - key: "HeaderKes"
          value: "h1=<<.header_keys.h1>>"
```

## 样例

> 参考`mock_data.yaml`文件

```
# curl --request GET -k --url https://127.0.0.1:5555/core

# curl --request PUT -k --url https://127.0.0.1:5555/core/with_path_var/1

# curl --request POST -k --url https://127.0.0.1:5555/core/with_path_var/2/with_meta_data/body
{"key":"value", "path_id": 2}

# curl --request POST -k --url https://127.0.0.1:5555/core/with_path_var/3/with_meta_data/body/header   --header 'h1: hello'
{"key1":"value1", "path_id": 3}

# curl --request POST -k  --url 'https://127.0.0.1:5555/core/with_path_var/5/with_meta_data/body/header/multi?p1=world'   --header 'content-type: multipart/form-data'   --header 'h1: hello'   --form f1=haha
{"key1":"value1", "path_id": 5, "head_h1": hello, "param_p1": world, "form_f1": haha}

# curl --request POST -k  --url 'https://127.0.0.1:5555/core/with_path_var/6/with_meta_data/body/header/all?p1=world'   --header 'content-type: application/json'   --header 'h1: hello'   --data '{
    "b1": "test",
    "b2": {
        "k": "toooooo"
    }
}'
{"key1":"value1", "path_id": 6, "head_h1": hello, "param_p1": world, "body_b1": test, "body_b2_k": toooooo}
```
