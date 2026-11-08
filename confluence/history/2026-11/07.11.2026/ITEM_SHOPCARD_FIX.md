1. ShopItemCard.tsx — выравниваем и убираем текст
Замени весь блок заголовка и бейджа на этот:

tsx
          <div className="mb-2 flex items-start justify-between gap-3">
            <h3 className="font-display text-lg font-semibold leading-snug text-text-primary line-clamp-2 min-h-[2.8rem]">
              {item.name}
            </h3>
            <div className="shrink-0 pt-1">
              <Badge tone={isOwned ? 'success' : 'indigo'}>
                {isOwned ? (
                  <Icon name="check" className="h-3 w-3" title="В инвентаре" />
                ) : (
                  categoryLabel
                )}
              </Badge>
            </div>
          </div>
Что изменилось:

items-start — бейдж прижат к верху заголовка (а не к центру).

line-clamp-2 — заголовок обрезается до 2 строк, если длинный.

min-h-[2.8rem] — резервирует высоту под 2 строки, чтобы кнопки под ними не прыгали.

Для «В инвентаре» — только иконка check, без текста.