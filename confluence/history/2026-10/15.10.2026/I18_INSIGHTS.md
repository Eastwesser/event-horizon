# i18n — Insights & Plan

> Дата: 15.10.2026
> Статус: **план** (не начато).
> Связанные документы:
> - `FUTURE_EVENT_HORIZON.md` — общий roadmap.
> - `EVENT_HORIZON_PATTERNS_15.10.2026.md` — паттерны бэкенда.

## Зачем

- Дать пользователям выбор языка (RU / EN / DE / ES).
- Показать проект немецкому другу — чтобы он сам переключился и потыкал.
- Подготовиться к расширению аудитории.
- Самое дешёвое улучшение с максимальным охватом.

## Что **не** делаем

- Не переводим **backend**-строки (gRPC-ошибки) — оставляем английский.
- Не переводим **email/уведомления** — только in-app.
- Не используем **сложные форматы** (валюта, даты) — `Intl` из коробки.
- Не тащим **полную локализацию контента** (карточки Берсерк, описания игр от авторов) — это отдельная задача.

## Стек

| Пакет | Зачем |
|-------|-------|
| `i18next` | Ядро. |
| `react-i18next` | React-биндинги. |
| `i18next-browser-languagedetector` | Автоопределение языка браузера. |

**Почему не `react-intl`:** `i18next` проще, лучше работает с React-хуками, легче масштабируется.

**Почему не `Lingui`:** `Lingui` требует Babel-плагин, избыточно для нашего размера.

## Архитектура

### Структура файлов
frontend/src/i18n/
├── index.ts # конфиг i18next + I18nextProvider
├── locales/
│ ├── ru.json # базовый (все ключи)
│ ├── en.json
│ ├── de.json
│ └── es.json
└── types.ts # (опционально) типы для ключей

text

### Подключение

В `main.tsx`:
```tsx
import './i18n';  // должен быть до <App />
В i18n/index.ts:

ts
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import ru from './locales/ru.json';
import en from './locales/en.json';
import de from './locales/de.json';
import es from './locales/es.json';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      ru: { translation: ru },
      en: { translation: en },
      de: { translation: de },
      es: { translation: es },
    },
    fallbackLng: 'ru',
    supportedLngs: ['ru', 'en', 'de', 'es'],
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: 'eh_lang',
    },
  });

export default i18n;
UI
Где показывать
Рядом с бургером в AppNavbar.tsx.

Иконка флага + dropdown.

На мобильных — внутри бургер-меню (первым пунктом).

Флаги
Код	Язык	Флаг
ru	Русский	🇷🇺
en	English	🇬🇧
de	Deutsch	🇩🇪
es	Español	🇪🇸
Поведение
Клик → i18n.changeLanguage(lang).

Сохраняется в localStorage['eh_lang'].

При следующем заходе — используется сохранённый язык.

Если язык не выбран — детектится из браузера.

Ключи переводов
Нейминг
Формат: <namespace>.<section>.<key>

Примеры:

json
{
  "common": {
    "back": "На главную",
    "cancel": "Отмена",
    "save": "Сохранить",
    "loading": "Загрузка..."
  },
  "auth": {
    "login": "Войти",
    "register": "Зарегистрироваться",
    "email": "Email",
    "password": "Пароль"
  },
  "games": {
    "play": "Играть",
    "boost": "Boost",
    "how_to_play": "Как играть?"
  },
  "shop": {
    "buy": "Купить",
    "owned": "В инвентаре",
    "stock": "В наличии: {{count}}"
  }
}
Плюрализация
json
{
  "shop": {
    "items_count": "{{count}} товар",
    "items_count_plural": "{{count}} товара",
    "items_count_plural_2": "{{count}} товаров"
  }
}
Или через i18next-формат:

json
{
  "shop": {
    "items_count": "{{count, plural, one {# товар} few {# товара} many {# товаров} other {# товара}}}"
  }
}
Волны перевода
Волна 1 (Must-Have)
Файл	Что
AppNavbar.tsx	Все пункты меню
Login.tsx, Register.tsx	Auth-формы
Home.tsx	Главная, список игр
PageHeader.tsx	«На главную»
Button.tsx	Нет (компонент не переводим)
GameShell.tsx	«Топ-10», «Как играть?»
Оценка: ~50 ключей.

Волна 2 (Игры)
Файл	Что
GameHelp.tsx	«Как играть» тексты
BoostCheckbox.tsx	«Boost за N ламп», эффекты
GameOverActions.tsx	«Новая игра», «Сохранить рекорд»
gameBoostCopy.ts	Все эффекты бустов
Оценка: ~80 ключей.

Волна 3 (Магазин, инвентарь, лидерборд)
Файл	Что
Shop.tsx, ShopItemCard.tsx	«Купить», «В инвентаре»
InventoryPage.tsx	«Мой инвентарь»
Leaderboard.tsx	«Топ-10»
Profile.tsx	Профиль
Оценка: ~60 ключей.

Волна 4 (Admin, Analytics, Authors)
Файл	Что
AdminPage.tsx	Админ-панель
AnalyticsDashboard.tsx	DAU / MAU / Retention
AuthorsPage.tsx	Авторы
Оценка: ~40 ключей.

Итого: ~230 ключей.

Что может пойти не так
Проблема	Решение
Ключи забыли перевести в en / de / es	Fallback на ru через fallbackLng
Плюрализация для 4 языков	Используем i18next-формат с plural
Длинные немецкие слова ломают UI	Дизайн с min-width / flex-shrink
Даты и числа	Intl.DateTimeFormat, Intl.NumberFormat
Backend-ошибки на английском	Оставляем как есть — это осознанно
Динамические ключи (boost.flappy)	Используем t(\boost.${gameId}`)`
План работ (завтра)
День 1 — инфраструктура
Установить пакеты:

bash
cd frontend
npm install i18next react-i18next i18next-browser-languagedetector
Создать src/i18n/index.ts + locales/ru.json + locales/en.json.

Подключить в main.tsx.

Проверить: useTranslation() работает в одном компоненте.

День 2 — UI
Добавить иконку флага в AppNavbar.tsx.

Dropdown с 4 языками.

Сохранение в localStorage['eh_lang'].

Тест: переключение языка не роняет приложение.

День 3–4 — Волна 1 (Must-Have)
Перевести AppNavbar, Login, Register, Home, PageHeader, GameShell.

Проверить на русском, английском, немецком, испанском.

Скриншоты — в confluence/history/2026-10/16.10.2026/i18n_screenshots/.

День 5+ — Волны 2–4
По мере сил. Не пытаться сделать всё за раз.

Что записать в FUTURE_EVENT_HORIZON.md
□ i18n: инфраструктура + 4 языка
□ i18n: волна 1 (Must-Have)
□ i18n: волна 2 (игры)
□ i18n: волна 3 (магазин)
□ i18n: волна 4 (admin)
Инсайты
Что уже готово к i18n
useSkins.ts — уже возвращает emptyMessage.

gameBoostCopy.ts — все тексты бустов в одном месте. Идеально для перевода.

gameHowTo.ts — все «Как играть» в одном файле.

pluralize.ts — уже есть plural-логика.

Что придётся тронуть
Каждый компонент с текстом.

constants — если есть.

Ошибки — catch блоки.

Что можно оставить
Backend — не трогаем.

Email — не трогаем.

Комментарии в коде — не трогаем.

Логи — не трогаем.

Рекомендации
Начать с ru.json — самый полный. Остальные языки добавлять по мере.

Не переводить всё сразу. Лучше 4 языка на 30 ключей, чем 1 язык на 300.

Использовать t('key', { count }) для плюрализации.

Проверять на мобильном — немецкие слова длиннее.

Скриншоты — обязательно. Иначе не вспомнишь, как выглядело.