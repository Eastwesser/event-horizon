. ValidateToken — не кэшируется
middleware/auth.go:

go
func RequireAuth(authClient AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := ExtractBearerToken(c.GetHeader("Authorization"))
		resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
		...
Нет кэша. Каждый закрытый запрос = gRPC-вызов в Auth. 140 browse-VU × 4 ручки = 560 ValidateToken/итерация. Auth делает HMAC + Redis lookup на каждый. Под 140 VU — 560 rps в Auth только на валидацию. Auth не тянет.

Фикс: кэш в Gateway (in-memory, TTL 5s, ключ = token). authReadCache уже есть (5s), но только для whoami/user/profile, не для ValidateToken. Нужен отдельный кэш ValidateToken с TTL 5s.

Приоритет: P1. Это самый большой избыток коллов.