from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
import base64

from app.db.repositories import WorkerRepository


@dataclass(frozen=True)
class ArchivePasswordSecret:
    tenant_id: str
    file_id: str
    password_ref_hash: str
    algorithm: str
    is_valid: bool
    validated: bool
    attempt_count: int


class PasswordSecretProvider(ABC):
    """Abstract source of archive passwords for worker extraction."""

    @abstractmethod
    def get_archive_password(self, tenant_id: str, file_id: str) -> ArchivePasswordSecret | None:
        """Return a password reference for the file, or None if unavailable."""

    @abstractmethod
    def resolve_password(self, tenant_id: str, file_id: str) -> str | None:
        """Convenience helper returning the plaintext password, when available."""

    @abstractmethod
    def record_password_attempt(
        self,
        tenant_id: str,
        file_id: str,
        password_ref_hash: str,
        is_valid: bool,
        *,
        increment_attempt: bool = False,
    ) -> None:
        """Persist an extraction validation outcome for the archive password reference."""


class LocalArchivePasswordSecretProvider(PasswordSecretProvider):
    """Local placeholder provider using `archive_password_refs` rows as the secret store."""

    local_algorithm = "local_base64"

    def __init__(self, repository: WorkerRepository) -> None:
        self._repository = repository

    def get_archive_password(self, tenant_id: str, file_id: str) -> ArchivePasswordSecret | None:
        if not tenant_id or not file_id:
            return None

        ref = self._repository.get_latest_archive_password_ref(tenant_id=tenant_id, file_id=file_id)
        if not ref:
            return None

        ref_hash = ref.get("password_ref_hash")
        if not isinstance(ref_hash, str) or not ref_hash:
            return None

        algorithm = str(ref.get("algorithm") or "")
        if algorithm and algorithm != self.local_algorithm:
            return None

        return ArchivePasswordSecret(
            tenant_id=ref.get("tenant_id") or tenant_id,
            file_id=ref.get("file_id") or file_id,
            password_ref_hash=ref_hash,
            algorithm=algorithm or self.local_algorithm,
            is_valid=bool(ref.get("is_valid", False)),
            validated=bool(ref.get("validated", False)),
            attempt_count=int(ref.get("attempt_count", 0) or 0),
        )

    def resolve_password(self, tenant_id: str, file_id: str) -> str | None:
        ref = self.get_archive_password(tenant_id=tenant_id, file_id=file_id)
        if ref is None:
            return None

        if ref.algorithm == self.local_algorithm:
            try:
                decoded = base64.b64decode(ref.password_ref_hash, validate=True)
                return decoded.decode("utf-8")
            except Exception:
                return None

        return None

    def record_password_attempt(
        self,
        tenant_id: str,
        file_id: str,
        password_ref_hash: str,
        is_valid: bool,
        *,
        increment_attempt: bool = False,
    ) -> None:
        try:
            self._repository.update_archive_password_ref_status(
                tenant_id=tenant_id,
                file_id=file_id,
                password_ref_hash=password_ref_hash,
                is_valid=is_valid,
                validated=True,
                increment_attempt_count=increment_attempt,
            )
        except Exception:
            # Password metadata is best-effort for extraction control.
            return
