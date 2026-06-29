import os


class Settings:
    def __init__(self) -> None:
        self.database_url = os.getenv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/fileplatform?sslmode=disable")
        self.redis_addr = os.getenv("REDIS_ADDR", "redis:6379")
        self.redis_password = os.getenv("REDIS_PASSWORD", "")
        self.redis_db = int(os.getenv("REDIS_DB", "0"))
        self.queue_name = os.getenv("QUEUE_NAME", "ingest.jobs")
        self.minio_endpoint = os.getenv("MINIO_ENDPOINT", "minio:9000")
        self.minio_access_key = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
        self.minio_secret_key = os.getenv("MINIO_SECRET_KEY", "minioadmin")
        self.minio_bucket = os.getenv("MINIO_BUCKET", "ingest")
        self.minio_use_ssl = os.getenv("MINIO_USE_SSL", "false").lower() == "true"
        self.opensearch_url = os.getenv("OPENSEARCH_URL", "http://opensearch:9200")
        self.opensearch_user = os.getenv("OPENSEARCH_USER", "admin")
        self.opensearch_pass = os.getenv("OPENSEARCH_PASSWORD", "admin")
        self.opensearch_index = os.getenv("OPENSEARCH_INDEX", "parsed-documents")


settings = Settings()

