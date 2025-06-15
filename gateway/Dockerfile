FROM python:3.12-slim

# Устанавливаем Poetry
RUN pip install poetry --index-url https://mirrors.aliyun.com/pypi/simple/

# Настройки Poetry
ENV POETRY_VIRTUALENVS_CREATE=false \
    POETRY_NO_INTERACTION=1

# Установка зеркала
RUN poetry config repositories.pypi https://mirrors.aliyun.com/pypi/simple/

# Устанавливаем рабочую директорию
WORKDIR /gateway

# Копируем все файлы из локальной папки в контейнер
COPY . /gateway

# Устанавливаем зависимости без установки самого проекта
RUN poetry install --only main --no-root

# Добавляем папку /getaway в PYTHONPATH для правильной работы импорта
ENV PYTHONPATH=/gateway:$PYTHONPATH
EXPOSE 8001

# Запуск приложения
CMD ["poetry", "run", "uvicorn", "gateway.main:app", "--host", "0.0.0.0", "--port", "8001"]
