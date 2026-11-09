Что изменилось:

lb вынесен в App (нужен для health-check и /ready).

healthCheckLoop — горутина, пингует /ready каждого бэкенда раз в 5s (timeout 2s), флипает Healthy.

Первый check() — сразу на старте (не ждём 5s).

Лог только при смене состояния (Healthy.Swap возвращает старое значение).

metricsHandler(a.lb) — передаём lb для осмысленного /ready.

