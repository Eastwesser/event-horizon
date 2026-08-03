📄 ОБНОВЛЁННЫЙ delivery/ansible/README.md

# 🚀 Delivery — Event Horizon

This is the place when our code meets the light, brought to life by Eastwesser.

---

## 🏗️ Архитектура

[GitHub Actions] (в облаке)
        │
        │ SSH + Ansible
        ▼
[Твоя виртуалка]
        │
        │ запускает
        ▼
[docker-compose up -d]
        │
        ▼
[Event Horizon работает]

---

## 📁 Структура

delivery/
├── ansible/
│   ├── site.yml              # Главный плейбук
│   ├── ansible.cfg           # Конфиг Ansible
│   └── inventory/
│       └── dev.ini           # Окружение dev
├── ci-cd/
│   └── .github/workflows/deploy.yml
├── k3s/
│   ├── deployment.yml
│   ├── service.yml
│   └── ingress.yml
└── README.md

---

## 🚀 АВТОДЕПЛОЙ ОДНОЙ КОМАНДОЙ

### Локальный деплой (с виртуалки)

cd /home/denismatveev/event_horizon/delivery/ansible
ansible-playbook -i inventory/dev.ini site.yml

### Проверка (без реальных изменений)

ansible-playbook -i inventory/dev.ini site.yml --check

---

## 🔧 ЧТО ДЕЛАЕТ ANSIBLE

| Шаг | Действие |
| :-- | :------- |
| 1   | Устанавливает Docker и зависимости (поддерживает Arch и Ubuntu) |
| 2   | Запускает и включает Docker service |
| 3   | Создаёт /opt/event-horizon |
| 4   | Клонирует репозиторий (или обновляет) |
| 5   | Создаёт .env файл с переменными окружения |
| 6   | Создаёт папки для volumes (postgres, redis, nats, grafana, prometheus) |
| 7   | Даёт права на Docker socket |
| 8   | Скачивает последние образы с Docker Hub |
| 9   | Запускает docker-compose up -d |
| 10  | Ждёт, пока сервисы станут здоровыми |
| 11  | Проверяет health-эндпоинты |

---

## ⚠️ ЧТО НУЖНО ПОМНИТЬ

### 1. Docker login на виртуалке

**Проблема:** Ansible запускается от root (через become: yes) и не видит твой docker login под пользователем.

**Решение:** Скопировать Docker-конфиг для root:

sudo mkdir -p /root/.docker
sudo cp /home/denismatveev/.docker/config.json /root/.docker/config.json

Или выполнить docker login под root:

sudo docker login -u eastwesser
# Введи Docker Hub токен

### 2. Образы должны быть на Docker Hub

**Проблема:** pull access denied — образ отсутствует на Docker Hub.

**Решение:** Запушить все образы:

cd /home/denismatveev/event_horizon
make docker-build-all
make docker-push-all

### 3. Docker Hub токен должен иметь права на запись

**Проблема:** unauthorized: access token has insufficient scopes

**Решение:** Создать токен с правами "Public Repo Read & Write" (не Read-only!).

### 4. Ansible не видит инвентарь

**Проблема:** Unable to parse inventory

**Решение:** Запускай из папки ansible/:

cd /home/denismatveev/event_horizon/delivery/ansible
ansible-playbook -i inventory/dev.ini site.yml

Или используй ansible.cfg (там уже прописан путь):

ansible-playbook site.yml

### 5. Sudo без пароля для Ansible

**Проблема:** sudo: требуется указать пароль

**Решение:** Добавить пользователя в sudoers без пароля:

echo "denismatveev ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/denismatveev
sudo chmod 440 /etc/sudoers.d/denismatveev

---

## 📋 ПРОВЕРКА СТАТУСА

# Health check (не требует авторизации)
curl -s http://localhost:8079/health | jq '.'

# Получить токен для авторизованных запросов
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

# Проверить товары в магазине (с авторизацией)
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Проверить инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Проверить баланс
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Проверить все контейнеры
docker ps --format "table {{.Names}}\t{{.Status}}"

---

## 🐛 ЧТО МЫ СЕГОДНЯ ИСПРАВЛЯЛИ

| Проблема | Решение |
| :------- | :------ |
| docker-compose pull не мог найти docker-compose.yml | Добавили -f deployments/docker-compose.cluster.yml |
| pull access denied | Сделали docker login под root |
| version: "{{ version }}" рекурсия | Переименовали в commit_hash |
| env: "{{ env }}" рекурсия | Переименовали в deploy_env |
| sudo: требуется указать пароль | Добавили NOPASSWD в sudoers |
| Образов нет на Docker Hub | Собрали и запушили через make docker-* |

---

## 🗺️ НЕСКОЛЬКО ОКРУЖЕНИЙ

# inventory/dev.ini
[dev]
dev-server ansible_host=192.168.1.100 ansible_user=denismatveev

# inventory/staging.ini
[staging]
staging-server ansible_host=192.168.1.101 ansible_user=denismatveev

# inventory/prod.ini
[prod]
prod-1 ansible_host=192.168.1.102 ansible_user=denismatveev
prod-2 ansible_host=192.168.1.103 ansible_user=denismatveev

Запуск на конкретном окружении:

ansible-playbook -i inventory/dev.ini site.yml       # только dev
ansible-playbook -i inventory/staging.ini site.yml  # только staging
ansible-playbook -i inventory/prod.ini site.yml     # только prod

---

## 🔥 ИТОГ

| Что | Статус |
| :-- | :----- |
| make deploy локально | ✅ Работает |
| ansible-playbook | ✅ Работает |
| Docker Hub образы | ✅ Запушены |
| CI/CD (GitHub Actions) | ✅ Готов |
| Inventory Service | ✅ Интегрирован |

Теперь ты можешь развернуть Event Horizon на любой машине одной командой. 🐴🚀

Сделано с ❤️ для Event Horizon