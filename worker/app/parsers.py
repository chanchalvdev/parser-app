import csv
import json
import os
import hashlib
import re
from urllib.parse import urlparse
import xml.etree.ElementTree as ET
from typing import Any, Dict

import chardet
from PyPDF2 import PdfReader
import openpyxl


def detect_kind(path: str) -> str:
    ext = os.path.splitext(path)[1].lower()
    return ext.lstrip(".")


IPV4_PATTERN = re.compile(
    r"\b(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)(?:\.(?:25[0-5]|2[0-4]\d|1?\d?\d)){3})\b"
)
EMAIL_PATTERN = re.compile(r"\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b")
URL_PATTERN = re.compile(r"\bhttps?://[^\s/$.?#].[^\s\"'<>]*", re.IGNORECASE)
DOMAIN_PATTERN = re.compile(r"\b(?:(?:(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}))\b")
HASH_PATTERN = re.compile(r"\b(?:[A-Fa-f0-9]{32}|[A-Fa-f0-9]{40}|[A-Fa-f0-9]{64})\b")


def _unique(values: list[str]) -> list[str]:
    seen = set()
    out: list[str] = []
    for value in values:
        if value in seen:
            continue
        seen.add(value)
        out.append(value)
    return out


def _safe_domain_from_url(raw_url: str) -> list[str]:
    parsed = urlparse(raw_url)
    if not parsed.hostname:
        return []
    host = parsed.hostname.lower()
    if host in {"localhost"}:
        return []
    return [host]


def extract_entities(raw: str) -> Dict[str, list[str]]:
    text = raw or ""
    ips = IPV4_PATTERN.findall(text)
    emails = EMAIL_PATTERN.findall(text)
    urls = URL_PATTERN.findall(text)
    hashes = HASH_PATTERN.findall(text)
    domains = DOMAIN_PATTERN.findall(text)

    domain_values = []
    for raw_url in urls:
        domain_values.extend(_safe_domain_from_url(raw_url))

    normalized_urls = [url.rstrip(").,;!?:") for url in urls]
    return {
        "ip_addresses": _unique(ips),
        "emails": _unique(emails),
        "urls": _unique(normalized_urls),
        "domains": _unique(domains + domain_values),
        "hashes": _unique(hashes),
    }


def _read_text_bytes(raw: bytes) -> str:
    if not raw:
        return ""
    encoding = chardet.detect(raw).get("encoding") or "utf-8"
    try:
        return raw.decode(encoding)
    except Exception:
        return raw.decode("utf-8", errors="ignore")


def parse_txt(path: str) -> Dict[str, Any]:
    with open(path, "rb") as f:
        raw = f.read()
    return {"content": _read_text_bytes(raw), "metadata": {"char_count": len(raw)}}


def parse_csv(path: str) -> Dict[str, Any]:
    rows = []
    with open(path, "r", newline="", encoding="utf-8", errors="ignore") as f:
        reader = csv.reader(f)
        for i, row in enumerate(reader):
            if i > 500:
                break
            rows.append(row)
    return {"content": "\n".join([",".join([str(c) for c in r]) for r in rows]), "metadata": {"rows": len(rows)}}


def parse_json(path: str) -> Dict[str, Any]:
    text = _read_text_bytes(open(path, "rb").read())
    payload: Any = {}
    try:
        payload = json.loads(text)
    except Exception:
        payload = None
    return {"content": json.dumps(payload, ensure_ascii=False) if payload is not None else text, "metadata": {"json": True}}


def parse_jsonl(path: str) -> Dict[str, Any]:
    with open(path, "r", encoding="utf-8", errors="ignore") as f:
        lines = []
        for i, line in enumerate(f):
            if i >= 200:
                break
            t = line.strip()
            if not t:
                continue
            try:
                payload = json.loads(t)
                lines.append(json.dumps(payload, ensure_ascii=False))
            except Exception:
                lines.append(t)
    return {"content": "\n".join(lines), "metadata": {"lines": len(lines)}}


def parse_xml(path: str) -> Dict[str, Any]:
    tree = ET.parse(path)
    root = tree.getroot()
    texts = " ".join(root.itertext())
    return {"content": texts, "metadata": {"root": root.tag}}


def parse_xlsx(path: str) -> Dict[str, Any]:
    wb = openpyxl.load_workbook(path, read_only=True, data_only=True)
    chunks = []
    for sheet in wb.worksheets:
        for i, row in enumerate(sheet.iter_rows(values_only=True)):
            if i > 300:
                break
            chunks.append("|".join(["" if c is None else str(c) for c in row]))
    return {"content": "\n".join(chunks), "metadata": {"sheets": len(wb.worksheets)}}


def parse_pdf(path: str) -> Dict[str, Any]:
    reader = PdfReader(path)
    pages = []
    for i, page in enumerate(reader.pages):
        if i >= 50:
            break
        try:
            pages.append(page.extract_text() or "")
        except Exception:
            pages.append("")
    return {"content": "\n".join(pages), "metadata": {"pages": len(pages)}}


def parse_binary_fallback(path: str) -> Dict[str, Any]:
    with open(path, "rb") as f:
        raw = f.read(2048)
    return {"content": _read_text_bytes(raw), "metadata": {"binary": True}}


def parse(path: str, kind_hint: str | None = None) -> Dict[str, Any]:
    kind = kind_hint or detect_kind(path)
    if kind in {"txt", "log"}:
        return parse_txt(path)
    if kind == "csv":
        return parse_csv(path)
    if kind == "json":
        return parse_json(path)
    if kind == "jsonl":
        return parse_jsonl(path)
    if kind == "xml":
        return parse_xml(path)
    if kind == "xlsx":
        return parse_xlsx(path)
    if kind == "pdf":
        return parse_pdf(path)
    return parse_txt(path) if _is_texty(path) else parse_binary_fallback(path)


def _is_texty(path: str) -> bool:
    with open(path, "rb") as f:
        head = f.read(2048)
    if not head:
        return True
    if any(b > 0x7F for b in head[:100]):
        return False
    return True


def sha256_file(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        while True:
            chunk = f.read(1024 * 1024)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()
