FROM python:3.12-slim

# 1. Устанавливаем системные зависимости (слой кэшируется, если вы не меняете этот блок)
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ffmpeg libsm6 libxext6 \
 && rm -rf /var/lib/apt/lists/*

# 2. Ставим Poetry
RUN pip install poetry

# 3. Отключаем venv внутри Poetry и отключаем интерактив
ENV POETRY_VIRTUALENVS_CREATE=false \
    POETRY_NO_INTERACTION=1

# 4. Рабочая директория
WORKDIR /detection

# 5. Копируем только манифесты зависимостей (для поэтапной установки poetry deps)
COPY pyproject.toml poetry.lock* /detection/

# 6. Устанавливаем python‑зависимости без копирования всего приложения
RUN poetry install --only main --no-root

# 7. Копируем остальной код
COPY . /detection

# 8. Экспортируем PYTHONPATH и порт
ENV PYTHONPATH=/detection:$PYTHONPATH
EXPOSE 8004

# 9. Запуск приложения
CMD ["uvicorn", "detection.main:app", "--host", "0.0.0.0", "--port", "8004"]
