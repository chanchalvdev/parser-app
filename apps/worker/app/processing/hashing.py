from __future__ import annotations

from pathlib import Path


def compute_sha256(file_path: str) -> str:
    digest = __import__("hashlib").sha256()
    path = Path(file_path)
    with path.open("rb") as fh:
        for block in iter(lambda: fh.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()
