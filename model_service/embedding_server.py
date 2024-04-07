import os
import sys
import torch
import time


from transformers import AutoTokenizer,AutoModel

script_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(script_dir))

print("=======\n")
print("base_path:",script_dir)
print("=======\n")

model_path=os.path.join(os.path.dirname(script_dir),"model_zoo/bge-base-zh-v1.5")

device = "cuda:0" if torch.cuda.is_available() else "cpu"

tokenizer = AutoTokenizer.from_pretrained(model_path)

model = AutoModel.from_pretrained(model_path)
model.to(device)
model.eval()

def embed_query(queries):
    instruction = "为这个句子生成表示以用于检索相关文章："
    query_tokens = tokenizer([instruction+q for q in queries],
                             padding=True,truncation=True,
                             return_tensors='pt',max_length=512).to(device)
    with torch.no_grad():
        model_output = model(**query_tokens)
        query_embeds = model_output[0][:,0]
    query_embeds = torch.nn.functional.normalize(query_embeds,p=2,dim=1)
    
    if "cuda" in device:
        result = query_embeds.cpu().numpy().tolist()
    else:
        result = query_embeds.numpy().tolist()
    
    #print(result)
    return result

def embed_passage(passages):
    
    query_tokens = tokenizer(passages,
                             padding=True,truncation=True,
                             return_tensors='pt',max_length=512).to(device)
    with torch.no_grad():
        model_output = model(**query_tokens)
        query_embeds = model_output[0][:,0]
    query_embeds = torch.nn.functional.normalize(query_embeds,p=2,dim=1)
    
    if "cuda" in device:
        result = query_embeds.cpu().numpy().tolist()
    else:
        result = query_embeds.numpy().tolist()
    
    #print(result)
    return result


def test_emb():
    queries = ["广西哪里好玩","山东哪里好玩"]
    query_embeds = embed_query(queries)

    passages = ["广西哪里好玩","广西不好玩","上海哪里好玩","广西的桂林很好玩，还有漓江也很好玩",
                "桂林山水","阳朔","上海外滩","梯田","去泰山旅游"]
    passage_embeds = embed_passage(passages)
    scores = torch.matmul(torch.tensor(query_embeds),torch.tensor(passage_embeds).T)
    print("scores:",scores)

    for query,scorelist in zip(queries,scores.tolist()):
        for passage,score in zip(passages,scorelist):
            print(score,"query:",query,"passage:",passage)



if __name__ == "__main__":
    test_emb()

