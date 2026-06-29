#!/usr/bin/env python3
"""Generate local sample archives from tests/fixtures files.

Creates:
- sample.zip
- sample_nested.zip
- sample.tar.gz
- sample.7z (if py7zr installed)
- password_sample.zip (if pyzipper installed)
"""

from __future__ import annotations

import argparse
import sys
import tarfile
import zipfile
from pathlib import Path
from typing import Iterable


PROJECT_ROOT = Path(__file__).resolve().parents[1]
FIXTURES_DIR = PROJECT_ROOT / "tests" / "fixtures"
DEFAULT_FILES: tuple[str, ...] = (
    "sample.txt",
    "sample.log",
    "sample.csv",
    "sample.json",
    "sample.jsonl",
)


def collect_paths(base_dir: Path, names: Iterable[str]) -> list[Path]:
    return [base_dir / name for name in names]


def create_zip(path: Path, files: list[Path]) -> None:
    with zipfile.ZipFile(path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for file in files:
            zf.write(file, file.name)
    print(f"created: {path}")


def create_nested_zip(path: Path, source_dir: Path, archive_root: str) -> None:
    with zipfile.ZipFile(path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for item in source_dir.rglob("*"):
            if item.is_file():
                rel = item.relative_to(source_dir.parent)
                arcname = Path(archive_root) / rel
                zf.write(item, arcname.as_posix())
    print(f"created: {path}")


def create_tar_gz(path: Path, files: list[Path], source_dir: Path) -> None:
    with tarfile.open(path, "w:gz") as tf:
        for file in files:
            tf.add(file, arcname=file.name)
        tf.add(source_dir, arcname=source_dir.name)
    print(f"created: {path}")


def create_7z(path: Path, files: list[Path], source_dir: Path) -> bool:
    try:
        import py7zr
    except ImportError:
        return False

    with py7zr.SevenZipFile(path, "w") as zf:
        for file in files:
            zf.write(file, file.name)
        zf.writeall(str(source_dir), source_dir.name)
    print(f"created: {path}")
    return True


def create_password_zip(path: Path, files: list[Path], password: str = "changeme") -> bool:
    try:
        import pyzipper
    except ImportError:
        return False

    with pyzipper.AESZipFile(path, "w", compression=pyzipper.ZIP_LZMA) as zf:
        zf.setpassword(password.encode("utf-8"))
        zf.setencryption(pyzipper.WZ_AES, nbits=256)
        for file in files:
            zf.write(file, file.name)
    print(f"created: {path} (password: {password})")
    return True


def ensure_fixtures_exist(base_dir: Path) -> None:
    if not base_dir.exists():
        raise FileNotFoundError(f"fixtures directory missing: {base_dir}")

    missing = [str(p) for p in DEFAULT_FILES if not (base_dir / p).exists()]
    missing += [str(base_dir / "nested_archive_source")] if not (base_dir / "nested_archive_source").exists() else []
    if missing:
        raise FileNotFoundError(
            "Missing fixture files required by archive generator: " + ", ".join(missing)
        )


def main() -> int:
    parser = argparse.ArgumentParser(description="Create local sample archives for parser testing")
    parser.add_argument(
        "--fixtures-dir",
        default=str(FIXTURES_DIR),
        help="Directory containing sample fixture files",
    )
    parser.add_argument(
        "--password",
        default="changeme",
        help="Password for password_sample.zip (default: changeme)",
    )
    args = parser.parse_args()

    fixtures_dir = Path(args.fixtures_dir).resolve()
    ensure_fixtures_exist(fixtures_dir)

    fixture_files = collect_paths(fixtures_dir, DEFAULT_FILES)
    nested_dir = fixtures_dir / "nested_archive_source"

    create_zip(fixtures_dir / "sample.zip", fixture_files)
    create_nested_zip(fixtures_dir / "sample_nested.zip", nested_dir, nested_dir.name)
    create_tar_gz(fixtures_dir / "sample.tar.gz", fixture_files, nested_dir)

    made_7z = create_7z(fixtures_dir / "sample.7z", fixture_files, nested_dir)
    if not made_7z:
        print(
            "skipped: sample.7z (py7zr not installed). "
            "Install with `pip install py7zr` to generate this archive.",
            file=sys.stderr,
        )

    made_password_zip = create_password_zip(
        fixtures_dir / "password_sample.zip",
        fixture_files,
        password=args.password,
    )
    if not made_password_zip:
        print(
            "skipped: password_sample.zip (pyzipper not installed). "
            "Install with `pip install pyzipper` to generate this archive.",
            file=sys.stderr,
        )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
