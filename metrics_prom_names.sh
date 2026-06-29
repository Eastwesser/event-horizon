#!/bin/bash

# Функция для поиска метрики по шаблону и возврата первого значения
query_metric_flexible() {
    local pattern=$1
    # Ищем все метрики, подходящие под шаблон
    local metric_names=$(curl -s "http://localhost:9090/api/v1/label/__name__/values" | jq -r ".data[] | select(contains(\"$pattern\"))" | head -1)
    
    if [[ -z "$metric_names" ]]; then
        echo "null"
        return
    fi
    
    # Берем первую найденную метрику и делаем запрос
    local query="rate(${metric_names}[1m])"
    local result=$(curl -s -G --data-urlencode "query=$query" "http://localhost:9090/api/v1/query" | jq -r '.data.result[0].value[1] // "null"')
    echo "$result"
}

# Использование:
# RPS=$(query_metric_flexible "gateway_requests")
# LATENCY=$(query_metric_flexible "gateway_request_duration_seconds")
# GOROUTINES=$(query_metric_flexible "go_goroutines")
