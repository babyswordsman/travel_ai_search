# 编译
下载源码后，需要进入仓库目录下进行编译
依赖：go版本 >= 1.22.0
```
cd search
go build -o ../ai_search_server server/main.go
```

# 依赖的第三方服务
- elasticsearch-8.14.1 用于文本检索（下载安装包后解压，使用自动的JRE） 
- pika_server 版本3.5.3 持久化的KV存储
- qdrant  版本1.8.3 向量检索（部分场景用了ES没有使用）

# 模型
- embedding model bge-base-zh-v1.5
- reranker model bge-reranker-base 


reranker model转ONNX格式见脚本：
```
model_service/save_to_onnx.py
```
# 下载地址：
- embedding model https://huggingface.co/BAAI/bge-base-zh-v1.5
- reranker model https://huggingface.co/BAAI/bge-reranker-base
- qdrant https://github.com/qdrant/qdrant/releases/tag/v1.8.3
- pika https://github.com/OpenAtomFoundation/pika/releases/tag/v3.5.3
- elasticsearch https://github.com/elastic/elasticsearch/releases/tag/v8.14.1

# 部署目录见下方的部署目录结构
```
$/home/service/app#
├── elasticsearch-8.14.1  # ES的安装根目录
├── model_zoo # 存放模型文件
│   ├── bge-base-zh-v1.5
│   │   ├── 1_Pooling
│   │   │   └── config.json
│   │   ├── config.json
│   │   ├── config_sentence_transformers.json
│   │   ├── modules.json
│   │   ├── pytorch_model.bin
│   │   ├── README.md
│   │   ├── sentence_bert_config.json
│   │   ├── special_tokens_map.json
│   │   ├── tokenizer_config.json
│   │   ├── tokenizer.json
│   │   └── vocab.txt
│   └── bge-reranker-base-onnx
│       ├── config.json
│       ├── model.onnx
│       ├── sentencepiece.bpe.model
│       ├── special_tokens_map.json
│       ├── tokenizer_config.json
│       └── tokenizer.json
├── pika_server # pika部署根目录
│   ├── conf
│   │   └── pika.conf
│   └── pika
├── qdrant_server  # qdrant 部署根目录
│   ├── config
│   │   └── config.yaml
│   ├── qdrant
└── travel_ai_search  # 本项目部署根目录
    ├── ai_search_server # go源码编译后文件
    ├── bin # 启动脚本
    │   ├── build.sh
    │   ├── start_es.sh
    │   ├── start_model_service.sh
    │   ├── start_pika_server.sh
    │   ├── start_qdrant_server.sh
    │   ├── start_server.sh
    │   ├── stop_es.sh
    │   ├── stop_model_service.sh
    │   ├── stop_pika_server.sh
    │   ├── stop_qdrant_server.sh
    │   └── stop_server.sh
    ├── blog # blog静态页面 可以忽略
    ├── config # 配置文件目录
    │   ├── conf.yaml
    ├── data # 导入测试商品等初始化数据文件目录
    ├── es_server -> ../elasticsearch-8.14.1 # ES部署软连
    ├── favicon.ico
    ├── index.html 
    ├── logs  # 日志目录
    ├── model_service # 模型启动脚本目录
    ├── model_zoo -> ../model_zoo  # 模型部署目录软连
    ├── pika_server -> ../pika_server # PIKA部署软连
    ├── qdrant_server -> ../qdrant_server # QDRANT部署软连
    └── resource # 静态页面，需要从源码处拷贝
        ├── chat.tmpl
        └── web
            ├── a.css
            ├── index.html
            ├── output.css
            ├── shopping.html
            └── upload.html
```
# python3环境依赖包
```
Package                  Version
------------------------ ----------------
aiohttp                  3.9.3
aiosignal                1.3.1
annotated-types          0.6.0
anyio                    3.7.1
async-timeout            4.0.3
attrs                    21.2.0
Automat                  20.2.0
Babel                    2.8.0
bcrypt                   3.2.0
blinker                  1.4
certifi                  2020.6.20
chardet                  4.0.0
click                    8.0.3
cloud-init               23.2.2
colorama                 0.4.4
coloredlogs              15.0.1
command-not-found        0.3
configobj                5.0.6
constantly               15.1.0
cryptography             3.4.8
datasets                 2.18.0
dbus-python              1.2.18
decorator                4.4.2
dill                     0.3.8
distro                   1.7.0
distro-info              1.1+ubuntu0.2
evaluate                 0.4.1
exceptiongroup           1.2.0
fastapi                  0.104.1
filelock                 3.13.3
flatbuffers              24.3.25
frozenlist               1.4.1
fsspec                   2024.2.0
h11                      0.14.0
httplib2                 0.20.2
huggingface-hub          0.22.2
humanfriendly            10.0
hyperlink                21.0.0
idna                     3.3
importlib-metadata       4.6.4
incremental              21.3.0
install                  1.3.5
jeepney                  0.7.1
Jinja2                   3.0.3
joblib                   1.3.2
jsonpatch                1.32
jsonpointer              2.0
jsonschema               3.2.0
keyring                  23.5.0
launchpadlib             1.10.16
lazr.restfulclient       0.14.4
lazr.uri                 1.0.6
MarkupSafe               2.0.1
more-itertools           8.10.0
mpmath                   1.3.0
multidict                6.0.5
multiprocess             0.70.16
netifaces                0.11.0
networkx                 3.2.1
numpy                    1.26.4
nvidia-cublas-cu12       12.1.3.1
nvidia-cuda-cupti-cu12   12.1.105
nvidia-cuda-nvrtc-cu12   12.1.105
nvidia-cuda-runtime-cu12 12.1.105
nvidia-cudnn-cu12        8.9.2.26
nvidia-cufft-cu12        11.0.2.54
nvidia-curand-cu12       10.3.2.106
nvidia-cusolver-cu12     11.4.5.107
nvidia-cusparse-cu12     12.1.0.106
nvidia-nccl-cu12         2.19.3
nvidia-nvjitlink-cu12    12.4.127
nvidia-nvtx-cu12         12.1.105
oauthlib                 3.2.0
onnx                     1.16.0
onnxruntime              1.17.1
optimum                  1.18.0
packaging                24.0
pandas                   2.2.1
pexpect                  4.8.0
pillow                   10.3.0
pip                      22.0.2
protobuf                 5.26.1
ptyprocess               0.7.0
pyarrow                  15.0.2
pyarrow-hotfix           0.6
pyasn1                   0.4.8
pyasn1-modules           0.2.1
pydantic                 2.4.2
pydantic_core            2.10.1
PyGObject                3.42.1
PyHamcrest               2.0.2
PyJWT                    2.3.0
pyOpenSSL                21.0.0
pyparsing                2.4.7
pyrsistent               0.18.1
pyserial                 3.5
python-apt               2.4.0+ubuntu3
python-dateutil          2.9.0.post0
python-debian            0.1.43+ubuntu1.1
python-linux-procfs      0.6.3
python-magic             0.4.24
pytz                     2022.1
pyudev                   0.22.0
PyYAML                   5.4.1
regex                    2023.12.25
requests                 2.25.1
responses                0.18.0
safetensors              0.4.2
scikit-learn             1.4.1.post1
scipy                    1.13.0
SecretStorage            3.3.1
sentence-transformers    2.6.1
sentencepiece            0.2.0
service-identity         18.1.0
setuptools               59.6.0
six                      1.16.0
sniffio                  1.3.1
sos                      4.5.6
ssh-import-id            5.11
starlette                0.27.0
sympy                    1.12
systemd-python           234
threadpoolctl            3.4.0
tokenizers               0.15.2
torch                    2.2.2
tqdm                     4.66.2
transformers             4.39.2
triton                   2.2.0
Twisted                  22.1.0
typing_extensions        4.11.0
tzdata                   2024.1
ubuntu-advantage-tools   8001
ubuntu-drivers-common    0.0.0
ufw                      0.36.1
unattended-upgrades      0.1
urllib3                  1.26.5
uvicorn                  0.24.0
wadllib                  1.3.6
wheel                    0.37.1
xkit                     0.0.0
xxhash                   3.4.1
yarl                     1.9.4
zipp                     1.0.0
zope.interface           5.4.0
```

# 部署目录结构
```
$/home/service/app#
├── elasticsearch-8.14.1
├── model_zoo
│   ├── bge-base-zh-v1.5
│   │   ├── 1_Pooling
│   │   │   └── config.json
│   │   ├── config.json
│   │   ├── config_sentence_transformers.json
│   │   ├── modules.json
│   │   ├── pytorch_model.bin
│   │   ├── README.md
│   │   ├── sentence_bert_config.json
│   │   ├── special_tokens_map.json
│   │   ├── tokenizer_config.json
│   │   ├── tokenizer.json
│   │   └── vocab.txt
│   ├── bge-base-zh-v1.5.zip
│   └── bge-reranker-base-onnx
│       ├── config.json
│       ├── model.onnx
│       ├── sentencepiece.bpe.model
│       ├── special_tokens_map.json
│       ├── tokenizer_config.json
│       └── tokenizer.json
├── pika_server
│   ├── conf
│   │   └── pika.conf
│   ├── db
│   │   └── db0
│   └── pika
├── qdrant_server
│   ├── config
│   │   └── config.yaml
│   ├── qdrant
└── travel_ai_search
    ├── ai_search_server
    ├── bin
    │   ├── build.sh
    │   ├── start_es.sh
    │   ├── start_model_service.sh
    │   ├── start_pika_server.sh
    │   ├── start_qdrant_server.sh
    │   ├── start_server.sh
    │   ├── stop_es.sh
    │   ├── stop_model_service.sh
    │   ├── stop_pika_server.sh
    │   ├── stop_qdrant_server.sh
    │   └── stop_server.sh
    ├── blog
    ├── config
    │   ├── conf.yaml
    │   ├── llm-search.cn.cer
    │   └── llm-search.cn.key
    ├── data
    │   ├── jd.txt
    │   ├── walmart
    ├── es_server -> /home/service/app/elasticsearch-8.14.1
    ├── favicon.ico
    ├── index.html
    ├── logs
    │   ├── ai_search.log
    │   ├── es_logs -> /home/service/app/elasticsearch-8.14.1/logs
    │   ├── model_server.log
    │   ├── pika_server.log
    │   └── qdrant_server.log
    ├── model_service
    │   ├── embedding_server.py
    │   ├── model_server.py
    │   ├── rerank_server_onnx.py
    │   ├── rerank_server.py
    │   └── save_to_onnx.py
    ├── model_zoo -> ../model_zoo
    ├── pika_server -> /home/service/app/pika_server
    ├── qdrant_server -> /home/service/app/qdrant_server
    └── resource
        ├── chat.tmpl
        └── web
            ├── a.css
            ├── index.html
            ├── output.css
            ├── shopping.html
            └── upload.html
```

