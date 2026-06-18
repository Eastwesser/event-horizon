📦 Установка k6 на Arch Linux

Вариант 1: Через AUR (yay)

bash
# Если нет yay — установим
sudo pacman -S --needed git base-devel
git clone https://aur.archlinux.org/yay.git
cd yay
makepkg -si

# Устанавливаем k6
yay -S k6
Вариант 2: Через AUR (paru)

bash
# Если нет paru
sudo pacman -S --needed git base-devel
git clone https://aur.archlinux.org/paru.git
cd paru
makepkg -si

# Устанавливаем k6
paru -S k6
Вариант 3: Скачать бинарник вручную

bash
# Скачать последнюю версию с GitHub
cd /tmp
wget https://github.com/grafana/k6/releases/download/v0.57.0/k6-v0.57.0-linux-amd64.tar.gz
tar -xzf k6-v0.57.0-linux-amd64.tar.gz
sudo cp k6-v0.57.0-linux-amd64/k6 /usr/local/bin/

# Проверить
k6 --version