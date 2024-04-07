import os
import sys
import torch
import time
from pathlib import Path

from optimum.onnxruntime import ORTModelForSequenceClassification
from transformers import pipeline, AutoTokenizer

script_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(script_dir))

print("=======\n")
print("base_path:",script_dir)
print("=======\n")
save_directory = os.path.join(os.path.dirname(script_dir),"model_zoo/bge-reranker-base-onnx")

device = "cuda:0" if torch.cuda.is_available() else "cpu"

model = ORTModelForSequenceClassification.from_pretrained(save_directory, file_name="model.onnx").to(device)
tokenizer = AutoTokenizer.from_pretrained(save_directory)


def predict(pairs):
    with torch.no_grad():
        inputs = tokenizer(pairs, padding=True, truncation=True, return_tensors='pt', max_length=512).to(device)
        scores = model(**inputs, return_dict=True).logits.view(-1, ).float()
    if "cuda" in device:
        result = scores.cpu().numpy().tolist()
    else:
        result = scores.numpy().tolist()
    return result

def test_predict():
    pairs1 = [
            ["广西哪里好玩","广西哪里好玩"],
            ["广西哪里好玩","广西不好玩"],
            ["广西哪里好玩","广西的桂林很好玩，还有漓江也很好玩"],
            ["广西哪里好玩","桂林山水"],
            ["广西哪里好玩","阳朔"],
            ["广西哪里好玩","上海外滩很好玩"],
            ["广西哪里好玩","上海外滩"],
            ["广西哪里好玩","去泰山旅游"],
         ]

    pairs2 = [
            ["山东哪里好玩","广西哪里好玩"],
            ["山东哪里好玩","广西不好玩"],
            ["山东哪里好玩","广西的桂林很好玩，还有漓江也很好玩"],
            ["山东哪里好玩","桂林山水"],
            ["山东哪里好玩","阳朔"],
            ["山东哪里好玩","上海外滩很好玩"],
            ["山东哪里好玩","上海外滩"],
            ["山东哪里好玩","去泰山旅游"],
         ]
    scores = predict(pairs1)
    for q_p,score in zip(pairs1,scores):
        print("score:",score,q_p)

    scores = predict(pairs2)
    for q_p,score in zip(pairs2,scores):
        print("score:",score,q_p)

if __name__ == "__main__":
    start = time.time()
    test_predict()
    end = time.time()
    print("time:",(end-start))

    start = time.time()
    test_predict()
    end = time.time()
    print("time:",(end-start))