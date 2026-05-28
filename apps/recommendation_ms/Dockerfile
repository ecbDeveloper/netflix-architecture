FROM python:3.12-slim-bookworm

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY gen/python/ gen
COPY apps/recommendation_ms/ .

EXPOSE 50052

ENV PYTHONPATH=/app/gen/python

CMD ["python", "main.py"]
