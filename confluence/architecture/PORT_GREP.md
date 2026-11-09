# Занятые порты на хосте (только 0.0.0.0)
docker ps -a --format '{{.Ports}}' \
  | tr ',' '\n' \
  | grep -oE '0\.0\.0\.0:[0-9]+' \
  | cut -d: -f2 \
  | sort -n \
  | uniq

# Свободные в диапазоне 9100-9999 (для observability)
for p in $(seq 9100 9999); do
  ss -tln | grep -q ":$p " || echo "$p"
done | head -20