# C3 — Author dashboard (local verify)

Smoke:

1. Login as author → `/author/dashboard` — OK  
2. «Мои карты» lists own items only (`author_id=me`) — OK  
3. Create card → `author_id` = me — OK  
4. Soft-delete → visible with `deleted=true` + restore — OK  
5. «Продажи» — buy → `sales_count`/`tickets_earned` match; cancel → aggregates back to 0; history rows keep `refunded` — OK  
6. «Профиль» — GET/PUT portfolio — OK  
7. Burger «Автор» for author; plain user → sales **403** — OK  
8. Non-author `/author/dashboard` → redirect `/register-author` — OK  

Screenshots in this folder:

- `c3-dashboard-cards-1920x1080.png`
- `c3-dashboard-sales-1920x1080.png`
- `c3-dashboard-profile-1920x1080.png`

No push until Emma OK.
