# Frontend auth — status (30.08.2026)

## Symptom

Console: `POST /api/auth/register` → **500**, then «can't even log in».

## Cause

Register form required **min 6** characters; Auth requires **8–128**. Short password → validation error; Gateway returned 500; UI showed «Ошибка соединения».

## Fix (in SPA)

- `Register.tsx`: `minLength={8}`, clearer error from API
- `Login.tsx`: show password validation / API message (not only «connection error»)

## Try now

1. Hard-refresh the Vite page (or restart `npm run dev` if needed).
2. Register with password **≥ 8** chars, e.g. `secret123`.
3. Login with the same credentials.

API smoke (works without frontend):

```bash
curl -sS -X POST http://localhost:8079/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"diag1788093613@example.com","password":"secret123"}'
```
