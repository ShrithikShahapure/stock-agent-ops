"""
Embeddings Provider abstraction.

Supports:
- llama.cpp (default) - OpenAI-compatible embeddings API
- Ollama (fallback) - for backward compatibility

Environment variables:
- LLAMA_CPP_EMBED_URL: Embeddings server URL (falls back to LLAMA_CPP_BASE_URL)
- EMBED_MODEL: Embedding model name (default: nomic-embed-text)
- OLLAMA_BASE_URL: Ollama server URL (fallback)
"""

import os
from typing import Optional

from logger.logger import get_logger

logger = get_logger()


def get_embeddings_client():
    """
    Returns a LangChain-compatible embeddings client.

    Priority:
    1. llama.cpp (OpenAI-compatible) if LLAMA_CPP_EMBED_URL or LLAMA_CPP_BASE_URL is set
    2. Ollama (langchain_ollama) if available
    3. None if nothing available

    Returns:
        A LangChain-compatible embeddings client or None
    """
    embed_model = os.getenv("EMBED_MODEL", "nomic-embed-text")
    llama_embed_url = os.getenv("LLAMA_CPP_EMBED_URL", os.getenv("LLAMA_CPP_BASE_URL", ""))
    ollama_url = os.getenv("OLLAMA_BASE_URL", "")

    # Try llama.cpp first (OpenAI-compatible API)
    if llama_embed_url:
        try:
            from langchain_openai import OpenAIEmbeddings

            logger.info(f"Using llama.cpp embeddings at {llama_embed_url}")

            return OpenAIEmbeddings(
                model=embed_model,
                base_url=llama_embed_url,
                api_key="not-needed",
            )

        except ImportError:
            logger.warning("langchain_openai not installed for embeddings, trying Ollama")
        except Exception as e:
            logger.warning(f"llama.cpp embeddings init failed: {e}, trying Ollama")

    # Fallback to Ollama
    if ollama_url:
        try:
            from langchain_ollama import OllamaEmbeddings

            logger.info(f"Using Ollama embeddings at {ollama_url}")

            return OllamaEmbeddings(
                model=embed_model,
                base_url=ollama_url
            )

        except ImportError:
            logger.warning("langchain_ollama not installed for embeddings")
        except Exception as e:
            logger.warning(f"Ollama embeddings init failed: {e}")

    # Return None if no embeddings available (semantic cache will be disabled)
    logger.warning("No embeddings client available, semantic cache will be disabled")
    return None
