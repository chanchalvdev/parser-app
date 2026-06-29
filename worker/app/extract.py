import os
import shutil
import tempfile
import zipfile
import tarfile
import py7zr
import rarfile


def safe_extract_path(target_dir: str, member_path: str) -> str:
    if not member_path:
        raise ValueError("archive member path is empty")

    normalized = os.path.normpath(member_path.replace("\\", "/"))
    if normalized in {"", "."}:
        raise ValueError("archive member path is empty")
    if os.path.isabs(normalized) or normalized.startswith("..") or normalized.startswith("../"):
        raise ValueError("archive member escapes extraction root")

    root = os.path.abspath(target_dir)
    target = os.path.abspath(os.path.join(target_dir, normalized))
    if os.path.commonpath([root, target]) != root:
        raise ValueError("archive member escapes extraction root")
    return target


def is_archive(path: str) -> bool:
    lower = path.lower()
    return lower.endswith((".zip", ".rar", ".tar", ".tar.gz", ".tgz", ".gz", ".7z"))


def extract_archive(path: str, target_dir: str, password: str | None = None) -> list[str]:
    extracted = []
    lower = path.lower()
    if lower.endswith(".zip"):
        extracted = _extract_zip(path, target_dir, password)
    elif lower.endswith(".rar"):
        extracted = _extract_rar(path, target_dir, password)
    elif lower.endswith(".tar") or lower.endswith(".tar.gz") or lower.endswith(".tgz") or lower.endswith(".gz"):
        extracted = _extract_tar(path, target_dir)
    elif lower.endswith(".7z"):
        extracted = _extract_7z(path, target_dir, password)
    else:
        raise ValueError(f"unsupported archive: {path}")
    return extracted


def _extract_zip(path: str, target_dir: str, password: str | None) -> list[str]:
    items = []
    with zipfile.ZipFile(path, "r") as zf:
        for info in zf.infolist():
            if info.is_dir():
                continue
            out_path = safe_extract_path(target_dir, info.filename)
            os.makedirs(os.path.dirname(out_path), exist_ok=True)
            if info.flag_bits & 0x1:
                if not password:
                    raise RuntimeError("zip password required")
                with zf.open(info, pwd=password.encode()) as src, open(out_path, "wb") as dst:
                    dst.write(src.read())
            else:
                with zf.open(info) as src, open(out_path, "wb") as dst:
                    dst.write(src.read())
            items.append(out_path)
    return items


def _extract_rar(path: str, target_dir: str, password: str | None) -> list[str]:
    try:
        with rarfile.RarFile(path) as rf:
            for info in rf.infolist():
                if info.isdir():
                    continue
                out_path = safe_extract_path(target_dir, info.filename)
                os.makedirs(os.path.dirname(out_path), exist_ok=True)
                if info.needs_password():
                    if not password:
                        raise RuntimeError("rar password required")
                    with rf.open(info, pwd=password) as src, open(out_path, "wb") as dst:
                        dst.write(src.read())
                else:
                    with rf.open(info) as src, open(out_path, "wb") as dst:
                        dst.write(src.read())
    except rarfile.Error as err:
        if "password required" in str(err).lower() and not password:
            raise RuntimeError("rar password required")
        raise
    return [os.path.join(target_dir, p) for p in _listdir_recursive(target_dir)]


def _extract_7z(path: str, target_dir: str, password: str | None) -> list[str]:
    kwargs = {"mode": "r"}
    if password:
        kwargs["password"] = password
    with py7zr.SevenZipFile(path, **kwargs) as z:
        for fileinfo in z.list():
            if fileinfo.is_directory:
                continue
            _ = safe_extract_path(target_dir, fileinfo.filename)
            target = os.path.join(target_dir, fileinfo.filename)
            os.makedirs(os.path.dirname(target), exist_ok=True)
        z.extractall(target_dir, pwd=password)
    return [os.path.join(target_dir, p) for p in _listdir_recursive(target_dir)]


def _extract_tar(path: str, target_dir: str) -> list[str]:
    extracted = []
    with tarfile.open(path, mode="r:*") as tf:
        for member in tf.getmembers():
            if member.isdir():
                continue
            out_path = safe_extract_path(target_dir, member.name)
            os.makedirs(os.path.dirname(out_path), exist_ok=True)
            src = tf.extractfile(member)
            if src is None:
                continue
            with src, open(out_path, "wb") as dst:
                dst.write(src.read())
            extracted.append(out_path)
    return extracted


def _listdir_recursive(root: str) -> list[str]:
    out = []
    for base, _, files in os.walk(root):
        for name in files:
            rel = os.path.relpath(os.path.join(base, name), root)
            out.append(rel)
    return out


def tempdir(prefix: str) -> str:
    return tempfile.mkdtemp(prefix=prefix)


def cleanup(path: str) -> None:
    shutil.rmtree(path, ignore_errors=True)
