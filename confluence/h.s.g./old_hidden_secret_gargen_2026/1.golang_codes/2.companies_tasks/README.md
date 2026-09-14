# Company packs — индекс haven

Корень: `1.golang_codes/2.companies_tasks/`

Теория/плейбуки → при дырах якорь в `prydwen_knowledge/`.  
Runnable — таймер 20–40 мин. Всегда: `export GOWORK=off`.

| Компания | Что есть | Runnable / старт |
|----------|----------|------------------|
| **Yandex** | LB least-conn + FinTech limits | `yandex/load_balancer/main.go`, `yandex/fin_tech/main.go` |
| **Avito** | Champions | `avito/champions/main.go` · `CHAMPIONS.md` |
| **T-Bank** | haven + 2 demo · полный `tasks.md` | `t-bank/code/01_min_cycle`, `02_tbank_study` · [README](t-bank/README.md) |
| **Ozon** | playbook (теория) + screening packs | `ozon/code/01..04_*` · [README](ozon/README.md) |
| **WB** | TechnoSchool L1–L3 | priority в [wb/README.md](wb/README.md) · `go run sol.go` |
| **HFLabs** | pointer → урок 24 (RLE, LRU, unique, Java→Go) | [hflabs/README.md](hflabs/README.md) → `../../../../24_unknown_date/` |
| **Sber** | pointer → урок 24 (payments, ATM, employees) | [sber/README.md](sber/README.md) → `../../../../24_unknown_date/` |
| **Zero** | Evrone theory checklist | [zero/README.md](zero/README.md) · `evrone.md` |
| kuper, lamoda, rambler, samokat, vk | заготовки | README-stub → PRINTABLE + strikes + dojo |

```bash
export GOWORK=off
cd yandex/load_balancer && go run main.go
cd ../../avito/champions && go run main.go
cd ../../t-bank/code/01_min_cycle && go run main.go
cd ../../ozon/code/01_uniq_randn && go run main.go
cd ../../../wb/wb_technoschool-main/level_1/task_7 && go run sol.go
```

**Порядок перед собесом в компанию:** twenty-strikes 1–10 → dojo паттерн → этот pack.
