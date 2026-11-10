# Все k6-скрипты в репо (по расширению .js с k6-импортом)
grep -rl 'from ["\x27]k6' --include='*.js' . 2>/dev/null

# Все каталоги, где что-то похоже на нагрузочные тесты
find . -type d \( -name 'k6' -o -name 'load_test*' -o -name 'loadtest*' \) 2>/dev/null

# Все упоминания k6 в Makefile / Taskfile / package.json / CI
grep -rniE 'k6|K6_VUS|browse\.js|purchase\.js|loadtest' \
  Makefile Taskfile.yml package.json .github/workflows 2>/dev/null

# Быстрый просмотр содержимого всех *.js под k6-каталогами
for f in $(find . -type f -name '*.js' -path '*k6*' 2>/dev/null); do
  echo "=== $f ==="
  head -20 "$f"
done
