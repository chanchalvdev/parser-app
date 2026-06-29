"""Pluggable parser framework."""

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.parsers.csv_parser import CsvParser
from app.parsers.excel_parser import ExcelParser
from app.parsers.generic_text_parser import GenericTextParser
from app.parsers.json_parser import JsonParser
from app.parsers.pdf_parser import PdfParser
from app.parsers.registry import ParserRegistry, create_default_parser_registry
from app.parsers.txt_parser import TxtParser
from app.parsers.xml_parser import XmlParser

__all__ = [
    "BaseParser",
    "ParsedRecord",
    "UnsupportedParserError",
    "ParserRegistry",
    "create_default_parser_registry",
    "TxtParser",
    "CsvParser",
    "JsonParser",
    "XmlParser",
    "ExcelParser",
    "PdfParser",
    "GenericTextParser",
]
