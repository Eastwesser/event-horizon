1. deployments/nginx/nginx.conf
Место: весь файл (правки: ; в location /, proxy_next_upstream, resolver не нужен, depends_on — в compose отдельно).

Почему max_fails=0: с дефолтным max_fails=1 Nginx пометит балансер dead на 10s после первого connection error. Если балансер жив, но один раз моргнул — Nginx 10s не будет к нему ходить. Нам это не надо: proxy_next_upstream и так ретраит. max_fails=0 отключает passive health-check, оставляем только активный ретрай.

Почему proxy_next_upstream http_502: 502 от балансера — это его ErrorHandler, т.е. connection error от балансера к gateway. Gateway не обработал запрос. Ретрай безопасен даже для POST.