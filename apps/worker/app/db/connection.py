from __future__ import annotations

from contextlib import contextmanager
from typing import Iterator
import psycopg2


class DatabaseConnection:
    def __init__(self, database_url: str):
        self.database_url = database_url

    @contextmanager
    def connect(self) -> Iterator[psycopg2.extensions.connection]:
        connection = psycopg2.connect(self.database_url)
        try:
            yield connection
            connection.commit()
        except Exception:
            connection.rollback()
            raise
        finally:
            connection.close()
