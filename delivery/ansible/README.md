[GitHub Actions] (в облаке)
        │
        │ SSH + Ansible
        ▼
[Твоя виртуалка] (где ты сейчас сидишь)
        │
        │ запускает
        ▼
[docker-compose up -d]
        │
        ▼
[Event Horizon работает на виртуалке]


🧠 Альтернатива: Ansible на виртуалке
Если GitHub Actions не нужен, можешь использовать Ansible прямо на виртуалке:

bash
# Установить Ansible
sudo pacman -S ansible  # Arch
# или
sudo apt install ansible  # Ubuntu

# Запустить плейбук локально
cd /home/denismatveev/event-horizon/delivery
ansible-playbook -i inventory/dev.ini ansible/site.yml