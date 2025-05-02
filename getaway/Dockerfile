FROM python:3.12-slim

# Устанавливаем Poetry
RUN pip install poetry

# Настройки Poetry
ENV POETRY_VIRTUALENVS_CREATE=false \
    POETRY_NO_INTERACTION=1

# Устанавливаем рабочую директорию
WORKDIR /getaway

# Копируем все файлы из локальной папки в контейнер
COPY . /getaway

# Создаём README.md, если его нет
RUN test -f README.md || echo "# Getaway API" > README.md

# Устанавливаем зависимости без установки самого проекта
RUN poetry install --only main --no-root

# Добавляем папку /getaway в PYTHONPATH для правильной работы импорта
ENV PYTHONPATH=/getaway:$PYTHONPATH

# Открываем порт
EXPOSE 8000

# Запуск приложения
CMD ["poetry", "run", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
