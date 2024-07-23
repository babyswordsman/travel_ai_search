import uvicorn
import time
from uvicorn.config import LOGGING_CONFIG
from fastapi import FastAPI,Body
from pydantic import BaseModel,Field
from typing import List,Any
from typing_extensions import Annotated

from pprint import pprint
import embedding_server
# import rerank_server_onnx

class QueryRequest(BaseModel):
    queries:List[str]

class PassageRequest(BaseModel):
    passages:List[str]

class EmbedResponse(BaseModel):
    embs:List[List[float]]
 
class RerankerRequest(BaseModel):
    q_p_pairs:List[List[str]]

class RerankerResponse(BaseModel):
    scores:List[float]

app = FastAPI()

@app.post("/embedding/query",response_model=EmbedResponse) 
async def embed_query(req:QueryRequest) -> Any:
    start_time = time.time()
    print(len(req.queries),req.queries)
    res = embedding_server.embed_query(req.queries)
    resp = EmbedResponse(embs=res)
    print("embed_query exec times:",time.time()-start_time)
    return resp

@app.post("/embedding/passage",response_model=EmbedResponse) 
async def embed_passage(req:PassageRequest) -> Any:
    start_time = time.time()
    print(len(req.passages),req.passages)
    res = embedding_server.embed_passage(req.passages)
    resp = EmbedResponse(embs=res)
    print("embed_passage exec times:",time.time()-start_time)
    return resp

# @app.post("/reranker/predict",response_model=RerankerResponse) 
# async def reranker_predict(req:RerankerRequest) -> Any:
#     start_time = time.time()
#     res = rerank_server_onnx.predict(req.q_p_pairs)
#     resp = RerankerResponse(scores = res)
#     print("reranker_predict exec times:",time.time()-start_time)
#     return resp

def getTimestamp():
    return time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())

if __name__ == "__main__":
    print(getTimestamp(),"start ...")
    uvicorn.run(app="model_server:app",host="0.0.0.0",port=8080,log_level="info",workers=1)
    embedding_server.embed_query(["一段文字"])
    rerank_server_onnx.predict([["问题","回答"]])
    print(getTimestamp(),"running ...")
