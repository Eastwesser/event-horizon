Шаг 2: Админские права (вседозволенность).

Что нужно:

Разрешить админу удалять/редактировать чужие карточки. Сейчас checkAuthorAccess проверяет author_id. Для админа — разрешить.

Разрешить админу отменять чужие покупки. Сейчас CancelPurchase работает только для своих. Нужно добавить проверку role = admin.

В services/inventory/internal/handler/grpc_handler.go:

Найди checkAuthorAccess. Сейчас там что-то вроде:

go
func checkAuthorAccess(ctx context.Context, authorID string) error {
    userID := md.Get("x-user-id")[0]
    if authorID != userID {
        return status.Error(codes.PermissionDenied, "you can only edit your own items")
    }
    return nil
}
Замени на:

go
func checkAuthorAccess(ctx context.Context, authorID string) error {
    userID := md.Get("x-user-id")[0]
    role := md.Get("x-user-role")[0] // <--- ДОБАВИТЬ
    if role == "admin" {
        return nil // <--- АДМИН МОЖЕТ ВСЁ
    }
    if authorID != userID {
        return status.Error(codes.PermissionDenied, "you can only edit your own items")
    }
    return nil
}
В services/shop/internal/service/shop_service.go:

Найди CancelPurchase. Сейчас там:

go
func (s *shopService) CancelPurchase(ctx context.Context, userID, itemID string) (*CancelResult, error) {
    // ...
}
Нужно добавить проверку роли. Но CancelPurchase не знает про роль. Нужно либо:

Вариант А: Добавить в CancelPurchase параметр role string и проверять.

Вариант Б: Создать отдельный метод AdminCancelPurchase, который не проверяет userID (отменяет любую покупку).

Вариант Б — чище. Но это новый gRPC-метод + Gateway-ручка.

Пока начни с Шага 2.1 (админ правит чужие карточки). Это проще (одна проверка в checkAuthorAccess). Потом сделаем Шаг 2.2 (отмена чужих покупок).