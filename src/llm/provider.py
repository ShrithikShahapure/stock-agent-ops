"""
LLM Chat Provider abstraction.

Supports:
- llama.cpp (default) - OpenAI-compatible API
- Ollama (fallback) - for backward compatibility

Environment variables:
- LLAMA_CPP_BASE_URL: llama.cpp server URL (default: http://localhost:8080/v1)
- LLM_MODEL: Model name (default: qwen3-7b)
- OLLAMA_BASE_URL: Ollama server URL (fallback)
"""

import os
from typing import Optional

from logger.logger import get_logger

logger = get_logger()


def get_chat_client(tools_list: Optional[list] = None):
    """
    Returns a LangChain-compatible chat client.

    Priority:
    1. llama.cpp (OpenAI-compatible) if LLAMA_CPP_BASE_URL is set
    2. Ollama (langchain_ollama) if available and OLLAMA_BASE_URL is set
    3. Mock client if nothing available

    Args:
        tools_list: Optional list of tools to bind to the LLM

    Returns:
        A LangChain-compatible chat client
    """
    model = os.getenv("LLM_MODEL", "qwen3-7b")
    llama_cpp_url = os.getenv("LLAMA_CPP_BASE_URL", "")
    ollama_url = os.getenv("OLLAMA_BASE_URL", "")

    # Try llama.cpp first (OpenAI-compatible API)
    if llama_cpp_url:
        try:
            from langchain_openai import ChatOpenAI

            logger.info(f"Using llama.cpp at {llama_cpp_url} with model {model}")

            llm = ChatOpenAI(
                model=model,
                base_url=llama_cpp_url,
                temperature=0.3,
                api_key="not-needed",  # llama.cpp doesn't require API key
                max_tokens=2048,
            )

            if tools_list:
                return llm.bind_tools(tools_list)
            return llm

        except ImportError:
            logger.warning("langchain_openai not installed, trying Ollama")
        except Exception as e:
            logger.warning(f"llama.cpp init failed: {e}, trying Ollama")

    # Fallback to Ollama
    if ollama_url:
        try:
            from langchain_ollama import ChatOllama

            # Map model name to Ollama model if needed
            ollama_model = os.getenv("OLLAMA_MODEL", "gpt-oss:20b-cloud")

            logger.info(f"Using Ollama at {ollama_url} with model {ollama_model}")

            llm = ChatOllama(
                model=ollama_model,
                temperature=0.3,
                base_url=ollama_url
            )

            if tools_list:
                return llm.bind_tools(tools_list)
            return llm

        except ImportError:
            logger.warning("langchain_ollama not installed")
        except Exception as e:
            logger.warning(f"Ollama init failed: {e}")

    # Return mock if nothing works
    logger.warning("No LLM available, using mock client")
    return _create_mock_client()


def _create_mock_client():
    """Create a mock LLM client for when no real LLM is available."""
    from langchain_core.messages import AIMessage

    class MockLLM:
        def invoke(self, messages, *args, **kwargs):
            return AIMessage(content="Mock response: LLM unavailable. Please configure LLAMA_CPP_BASE_URL or OLLAMA_BASE_URL.")

        def bind_tools(self, tools):
            return self

    return MockLLM()
