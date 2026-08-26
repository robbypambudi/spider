from __future__ import annotations

import re
import unicodedata
from typing import Protocol

from pydantic import BaseModel, Field


class PreprocessResult(BaseModel):
    text: str
    original_length: int
    metadata: dict[str, str] = Field(default_factory=dict)


class Preprocessor(Protocol):
    def process(self, text: str) -> PreprocessResult: ...


class DefaultPreprocessor:
    """Normalize unicode and whitespace without destroying signal for detectors."""

    _whitespace = re.compile(r"[ \t]+")
    _newlines = re.compile(r"\n{3,}")

    def process(self, text: str) -> PreprocessResult:
        original_length = len(text)
        normalized = unicodedata.normalize("NFKC", text)
        normalized = normalized.replace("\r\n", "\n").replace("\r", "\n")
        normalized = self._whitespace.sub(" ", normalized)
        normalized = self._newlines.sub("\n\n", normalized)
        return PreprocessResult(
            text=normalized.strip(),
            original_length=original_length,
            metadata={"strategy": "nfkc-whitespace"},
        )
