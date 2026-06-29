from __future__ import annotations

from typing import Iterable, Optional

from app.processing.file_detector import FileDetectionResult
from app.extractors.base_extractor import BaseArchiveExtractor, ArchiveLimits, ExtractedFile
from app.extractors.rar_extractor import RarExtractor
from app.extractors.sevenz_extractor import SevenZipExtractor
from app.extractors.tar_extractor import TarExtractor
from app.extractors.zip_extractor import ZipExtractor


class ArchiveExtractor:
    def __init__(self, extractors: Optional[Iterable[BaseArchiveExtractor]] = None):
        self._extractors = list(
            extractors
            or (
                ZipExtractor(),
                TarExtractor(),
                SevenZipExtractor(),
                RarExtractor(),
            )
        )

    def select(self, file_path: str, detection_result: FileDetectionResult) -> BaseArchiveExtractor:
        for extractor in self._extractors:
            if extractor.can_extract(file_path, detection_result):
                return extractor
        raise ValueError(f"no extractor available for detection {detection_result.detected_file_type}")

    def can_extract(self, file_path: str, detection_result: FileDetectionResult) -> bool:
        try:
            self.select(file_path, detection_result)
            return True
        except ValueError:
            return False

    def extract(
        self,
        file_path: str,
        detection_result: FileDetectionResult,
        destination_dir: str,
        *,
        password: str | None = None,
        limits: ArchiveLimits | None = None,
    ) -> list[ExtractedFile]:
        extractor = self.select(file_path, detection_result)
        return extractor.extract(file_path, destination_dir, password=password, limits=limits)
