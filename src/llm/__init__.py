"""LLM Provider abstraction for llama.cpp integration."""

from src.llm.provider import get_chat_client
from src.llm.embeddings import get_embeddings_client

__all__ = ["get_chat_client", "get_embeddings_client"]
