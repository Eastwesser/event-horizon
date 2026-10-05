# C2 — Admin approval (local verify)

Smoke (API) — all OK:

1. User submits application (C1 path) — OK  
2. Admin `GET /api/authors/applications?status=pending` — OK  
3. Approve → `status=approved`, Auth `role=author`, profile upserted — OK  
4. Re-login as user → `role=author`; apply again → 400 already author — OK  
5. Profile has `display_name` / `portfolio` / `bio=motivation` / `verified_at_unix` — OK  
6. Reject + `reviewer_note` — OK  
7. Double-approve → **400** `application already reviewed` — OK  
8. Admin refresh approved/rejected filters — OK  

Screenshots:

- `c2-admin-applications-pending-1920x1080.png` — tab «Заявки» / Pending  
- `c2-admin-approve-success-pending-refresh-1920x1080.png` — flash «Одобрено: …»  
- `c2-admin-applications-approved-1920x1080.png` — Approved filter  
- `c2-author-profile-upsert-1920x1080.png` — authors profile after approve  

Follow-up (map): JWT role refresh UX-gap — re-login needed until refresh reloads role from Auth DB. Notifications deferred.

No push until Emma OK.
