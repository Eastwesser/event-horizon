# 🚀 Delivery — Event Horizon

This is the place when our code meets the light, brought to life by Eastwesser.

## 🏗️ Architecture

[GitHub Actions] → [Ansible] → [VM / k3s]
        │               │            │
        ▼               ▼            ▼
  Build images   Deploy containers  Push to Hub
                        │
                        ▼
                 Run migrations

## 📁 Structure

delivery/
├── ansible/
│   ├── site.yml              # Main playbook
│   └── inventory/            # Dev/Staging/Prod inventory
│       ├── dev.ini
│       ├── staging.ini
│       └── prod.ini
├── ci-cd/
│   └── .github/workflows/deploy.yml
├── k3s/                      # Kubernetes manifests
│   ├── deployment.yml
│   ├── service.yml
│   └── ingress.yml
└── README.md

## 🚀 Quick start

# Dev deploy
cd /home/denismatveev/event_horizon/delivery/ansible
ansible-playbook -i inventory/dev.ini site.yml

# Deploy specific version
cd /home/denismatveev/event_horizon/delivery/ansible --extra-vars "version=v1.0.6"

# Check service status
curl -s http://localhost:8079/health | jq '.'

## 🔐 Secrets

Add to GitHub Secrets:

- DOCKER_USERNAME — Docker Hub username
- DOCKER_PASSWORD — Docker Hub password/token
- ANSIBLE_SSH_KEY — SSH key for servers"