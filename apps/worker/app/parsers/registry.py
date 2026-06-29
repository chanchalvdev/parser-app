from __future__ import annotations

from typing import Mapping, Sequence

from app.processing.file_detector import FileDetectionResult

from app.parsers.base_parser import BaseParser, UnsupportedParserError
from app.parsers.csv_parser import CsvParser
from app.parsers.excel_parser import ExcelParser
from app.parsers.generic_text_parser import GenericTextParser
from app.parsers.json_parser import JsonParser
from app.parsers.pdf_parser import PdfParser
from app.parsers.txt_parser import TxtParser
from app.parsers.xml_parser import XmlParser


class ParserRegistry:
    def __init__(self, parsers: Sequence[BaseParser] | None = None):
        self._parsers: list[BaseParser] = list(
            parsers
            or (
                TxtParser(),
                CsvParser(),
                JsonParser(),
                XmlParser(),
                ExcelParser(),
                PdfParser(),
                GenericTextParser(),
            )
        )

    def register(self, parser: BaseParser) -> None:
        self._parsers.append(parser)

    def select_parser(
        self,
        file_metadata: Mapping[str, object],
        detection_result: FileDetectionResult,
    ) -> BaseParser:
        for parser in self._parsers:
            if parser.can_parse(file_metadata, detection_result):
                return parser

        if detection_result.is_text_like:
            return GenericTextParser()

        raise UnsupportedParserError(
            "no parser available for file type "
            f"{detection_result.detected_file_type} / extension {detection_result.extension}"
        )


def create_default_parser_registry() -> ParserRegistry:
    return ParserRegistry(
        (
            TxtParser(),
            CsvParser(),
            JsonParser(),
            XmlParser(),
            ExcelParser(),
            PdfParser(),
            GenericTextParser(),
        )
    )
