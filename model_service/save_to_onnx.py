import os
import sys

from optimum.onnxruntime import ORTModelForSequenceClassification
from transformers import AutoTokenizer
import transformers

script_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(script_dir))

print("=======\n")
print("base_path:",script_dir)
print("=======\n")
model_path=os.path.join(os.path.dirname(script_dir),"model_zoo/bge-reranker-base")
save_directory = os.path.join(os.path.dirname(script_dir),"model_zoo/bge-reranker-base-onnx")


tokenizer = AutoTokenizer.from_pretrained(model_path)
ort_model = ORTModelForSequenceClassification.from_pretrained(model_path,export=True)

ort_model.save_pretrained(save_directory)
tokenizer.save_pretrained(save_directory)

print("=======\n")
print("save end:",save_directory)
print("=======\n")