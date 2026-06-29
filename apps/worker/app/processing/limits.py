from __future__ import annotations


def is_file_too_large(file_size_bytes: int, max_size_mb: int) -> bool:
    if file_size_bytes < 0:
        raise ValueError("file size cannot be negative")
    max_bytes = max_size_mb * 1024 * 1024
    return file_size_bytes > max_bytes
