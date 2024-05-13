# travel_ai_search
## 代办事项
- [x] 提取网页正文内容
- [ ] 登录
- [ ] 历史对话的相关性内容提取,长问题的对话进行摘要
- [x] query 改写
- [ ] 判断是否需要进行外部检索
- [ ] 思维链

```
apt install python3-pip


pip3 install uvicorn==0.24 fastapi==0.104.1 pydantic==2.4.2
pip3 install transformers==4.39.2
pip3 install sentence_transformers==2.6.1
pip3 install --upgrade-strategy eager install optimum[onnxruntime]==1.18.0
```

Successfully installed coloredlogs-15.0.1 datasets-2.18.0 dill-0.3.8 evaluate-0.4.1 flatbuffers-24.3.25 fsspec-2024.2.0 humanfriendly-10.0 install-1.3.5 multiprocess-0.70.16 onnx-1.16.0 onnxruntime-1.16.3 optimum-1.18.0 pandas-2.0.3 protobuf-5.26.1 pyarrow-15.0.2 pyarrow-hotfix-0.6 python-dateutil-2.9.0.post0 pytz-2024.1 responses-0.18.0 sentencepiece-0.2.0 six-1.16.0 tzdata-2024.1 xxhash-3.4.1
Successfully installed tokenizers-0.15.2 transformers-4.39.2

BAAI模型下载网址：https://model.baai.ac.cn/models
 

 https://huggingface.co/docs/optimum/v1.2.1/en/quickstart

模型服务接口测试
 ```
 curl -X POST -H "Content-Type:application/json" \
-d'{"queries":["德天瀑布","广西哪里好玩"]}' \
http://127.0.0.1:8080/embedding/query

curl -X POST -H "Content-Type:application/json" \
-d'{"passages":["德天瀑布","广西哪里好玩"]}' \
http://127.0.0.1:8080/embedding/passage

curl -X POST -H "Content-Type:application/json" \
-d'{"q_p_pairs":[["山东哪里好玩","广西哪里好玩"],["山东哪里好玩","广西不好玩"],["山东哪里好玩","广西的桂林很好玩，还有漓江也很好玩"]]}' \
http://127.0.0.1:8080/reranker/predict
```

爬虫：
Beautiful Soup   lxml 网页解析

requests 网络请求

scrapy 一个python写的爬虫框架

分布式： scrapy-redis/scrapy cluster


3、模拟浏览器：selenium   splash 

反爬： IP  header校验  cookie  加密   JS混淆   动态内容拼接

升级GLIBC 
vim /etc/apt/sources.list
```
#增加
deb http://mirrors.aliyun.com/ubuntu/ jammy main
apt-get update
apt install libc6

strings /lib/x86_b4-linux-gnu/libc.so.6 | grep GLIBC_
```

# tika-server
文本提取使用了tika，启动tika见命令行
https://github.com/apache/tika/tree/main/tika-server
下载页面：
https://tika.apache.org/download.html
```
$ java -jar tika-server/target/tika-server.jar --help
   usage: tikaserver
    -?,--help           this help message
    -h,--host <arg>     host name (default = localhost)
    -l,--log <arg>      request URI log level ('debug' or 'info')
    -p,--port <arg>     listen port (default = 9998)
    -s,--includeStack   whether or not to return a stack trace
                        if there is an exception during 'parse'
```

# docconv
https://github.com/sajari/docconv
```
# 安装依赖
apt-get install poppler-utils wv unrtf tidy
```

# 编译tailwindcss
```
npx tailwindcss -i ./web/a.css -o ./web/output.css --watch
```