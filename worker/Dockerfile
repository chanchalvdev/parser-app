FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

HEALTHCHECK --interval=20s --timeout=5s --start-period=10s --retries=10 \
  CMD python -c "import redis, json; r=redis.Redis(host='redis', port=6379); r.ping(); print('ok')"

CMD ["python", "-m", "app.main"]

