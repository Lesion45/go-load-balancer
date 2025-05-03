# Go Load Balancer

Простой, но мощный балансировщик нагрузки на Go с поддержкой health checks, rate limiting и мониторингом через Prometheus. 

## Вопросы
1. Однажды в рамках лабораторной работы мне пришлось лезть в память работающего калькулятора Windows, чтобы заменить дефолтные выражения по типу "Деление на ноль невозможно" на кастомные вещи. Пришлось искать строку в памяти через отладчик, выяснять, как она защищена от модификации, и аккуратно патчить её прямо в процессе.
2. Ошибся в настройке маршрутизации на учебном стенде и положил его. Стенд перестал отвечать, доступ пропал. Пришлось восстанавливать из резервной копии. После этого разобрался в проблеме и исправил конфигурацию.
3. Всегда хотелось принять участие в разработке облачных решений, поскольку очень перспективная и интересная сфера. Поэтому в рамках стажировки хочу еще больше погрузится в нее и прокачать свои скилы по максимум. Буду рад учиться у команды, участвовать в обсуждениях и помогать в решении задач. Для меня важно не только писать код, но и понимать, как всё устроено в реальных решениях.

## Особенности

- **Несколько стратегий балансировки**: Round Robin, Least Connections
- **Health Checks**: Автоматическое исключение нерабочих бэкендов
- **Rate Limiting**: Защита от перегрузки (Token Bucket)
- **Метрики Prometheus**: Встроенная интеграция для мониторинга
- **Grafaна**: Поддержка визуализации метрик
- **Гибкая конфигурация**: Через YAML-файлы и переменные окружения
- **Логирование**: Встроенная поддержка логирования с помощью Zap

## Быстрый старт

### 1. Установка

```bash
git clone https://github.com/your-repo/go-load-balancer.git
cd go-load-balancer
```

### 2. Конфигурация
Создайте файл `secrets.env` для хранения конфиденциальной информации:
```env
# Основные настройки приложения
ENV= # dev, prod
HOST=
PORT=

# Настройки балансировщика нагрузки
BALANCER_STRATEGY_NAME= # round_robin, least_connections

# Настройки health check
HEALTH_CHECK_ENDPOINT=
HEALTH_CHECK_BASE_INTERVAL=
HEALTH_CHECK_MAX_INTERVAL=

# Настройки rate limiting
RATE_LIMITING_STRATEGY_NAME= # token_bucket
```
Пример файла `secrets.env`:
```env
# Основные настройки приложения
ENV=dev
HOST=0.0.0.0
PORT=8080

# Настройки балансировщика нагрузки
BALANCER_STRATEGY_NAME=round_robin

# Настройки health check
HEALTH_CHECK_ENDPOINT=/health
HEALTH_CHECK_BASE_INTERVAL=10s
HEALTH_CHECK_MAX_INTERVAL=2m

# Настройки rate limiting
RATE_LIMITING_STRATEGY_NAME=token_bucket
```
Добавьте бэкенды в файл `backends.yaml`:
```yaml
backends:
  - name:
    address:
    burst: 
    rate: 
```
Пример файла `backends.yaml`:
```yaml
backends:
  - name: backend1
    address: http://host.docker.internal:8085
    burst: 20
    rate: 10
  - name: backend2
    address: http://host.docker.internal:8086
    burst: 20
    rate: 10
  - name: backend3
    address: http://host.docker.internal:8087
    burst: 20
    rate: 10
```
### 3. Запуск
```bash
docker-compose up --build
```

## Дополнительная информация
Готовый шаблон для `Grafana`:
```JSON
{
  "title": "Load Balancer Metrics",
  "description": "Dashboard for monitoring HTTP requests and latency",
  "tags": ["go", "prometheus", "loadbalancer"],
  "panels": [
    {
      "title": "Total Requests",
      "type": "graph",
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[1m])) by (method, code)",
          "legendFormat": "{{method}} ({{code}})",
          "refId": "A"
        }
      ],
      "xaxis": { "mode": "time" },
      "yaxis": { "format": "short" },
      "bars": false,
      "lines": true,
      "datasource": "Prometheus"
    },
    {
      "title": "Requests per Backend",
      "type": "piechart",
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[5m])) by (backend)",
          "legendFormat": "{{backend}}",
          "refId": "B"
        }
      ],
      "datasource": "Prometheus"
    },
    {
      "title": "Request Duration (95th percentile)",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[1m])) by (le, backend))",
          "legendFormat": "p95 {{backend}}",
          "refId": "C"
        },
        {
          "expr": "histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[1m])) by (le, backend))",
          "legendFormat": "p50 {{backend}}",
          "refId": "D"
        }
      ],
      "xaxis": { "mode": "time" },
      "yaxis": { "format": "s" },
      "datasource": "Prometheus"
    },
    {
      "title": "Error Rate",
      "type": "stat",
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{code=~\"5..\"}[1m])) / sum(rate(http_requests_total[1m]))",
          "format": "percentunit",
          "refId": "E"
        }
      ],
      "datasource": "Prometheus",
      "thresholds": "0.01,0.05",
      "colorMode": "value"
    }
  ],
  "templating": {
    "list": [
      {
        "name": "backend",
        "label": "Backend",
        "type": "query",
        "query": "label_values(http_requests_total, backend)"
      }
    ]
  },
  "time": {
    "from": "now-6h",
    "to": "now"
  }
}
```