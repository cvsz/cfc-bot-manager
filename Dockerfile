# Production-Grade Python Runtime Dockerfile
FROM python:3.12-alpine AS runtime

LABEL maintainer="maintainers@example.com"
LABEL description="zTemplate Production-Grade Python Container Baseline"

WORKDIR /app

# Copy necessary application modules, scripts, and tests
COPY scripts/ /app/scripts/
COPY tests/ /app/tests/
COPY README.md ABOUT.md /app/

# Create a secure non-root user and set permissions
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app

USER appuser

ENV PYTHONUNBUFFERED=1

# Default validation command
CMD ["python3", "-m", "unittest", "discover", "-s", "tests", "-v"]
