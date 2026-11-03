# FIXED event_horizon/frontend/src/components/ui/Modal.tsx

1. Ширина (max-w-lg vs max-w-3xl)
Раньше max-w-lg стоял жестко в базовых классах. Если ты передавал className="max-w-3xl", они конфликтовали, и побеждал тот, что был "сильнее" в CSS (обычно max-w-3xl, но tailwind-merge в cn должен был это чинить). Теперь базовый max-w-lg остался, а твой className (с max-w-3xl) подмешивается в конец через cn(). tailwind-merge увидит два max-w-* и оставит последний.

Результат: Маленькие модалки останутся узкими (max-w-lg), а CreateModal и EditModal станут широкими (max-w-3xl).

2. Скролл (Шапка теперь прибита)
Раньше скроллилась вся белая плашка целиком (включая заголовок). Теперь мы сделали flex flex-col и обернули контент в flex-1 overflow-y-auto.

Результат: Если контента много (как в форме создания карточки), скроллиться будет только форма, а заголовок "Создать товар" всегда останется на экране. Это сильно улучшает UX.

3. Паддинги (Отступы)
Раньше p-6 sm:p-8 применялся ко всей модалке разом. Из-за этого контент при скролле "прилипал" к краям.
Теперь паддинги разнесены: px-6 pt-6 для шапки и px-6 py-4 для контента.

Результат: Визуально стало аккуратнее, текст не бьется об края при скролле.

## !!! InventoryCreateModal.tsx и InventoryEditModal.tsx найди строку:

tsx
<form onSubmit={handleSubmit} className="flex max-h-[80vh] flex-col gap-5 overflow-y-auto">

И замени на:

tsx
<form onSubmit={handleSubmit} className="flex flex-col gap-5">