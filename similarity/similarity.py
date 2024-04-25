from __future__ import annotations
import os
from sentence_transformers import SentenceTransformer
from sentence_transformers import util

def lengths(data: list[str]) -> list[int]:
    """Simple function to test functionality. Returns length of each string in data."""
    out = [len(s) for s in data]
    return out


def similarity(target:str, others: list[str]) -> list[float]:
    sentence_model = _get_sentence_model()
    target_embedding = sentence_model.encode([target], normalize_embeddings=True)
    others_embedding = sentence_model.encode(others, normalize_embeddings=True)
    similarity = util.dot_score(target_embedding, others_embedding)
    return similarity.tolist()[0]

def _get_sentence_model(model_name: str = 'all-MiniLM-L6-v2'):
    current_dir = os.path.dirname(os.path.realpath(__file__))
    path = os.path.join(current_dir, model_name)
    sentence_model = SentenceTransformer(path)
    return sentence_model

# similarities = similarity("I am a test string", ["I am also a testing string", "there is nothing in common with what I say"])
# print(similarities)
# print(type(similarities[0]))