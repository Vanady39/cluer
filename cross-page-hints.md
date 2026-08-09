# Кросс-страничные подсказки: что починено на бэке и что нужно на фронте

Статус: бэкенд готов и в ветке `feature/frontendcompose`. Фронт **не тронут** —
без правок ниже баг остаётся воспроизводимым.

## В чём был баг

Фронт даёт задать для каждой подсказки страницу — поле `page_path`. Этого поля
не существовало нигде в бэкенде: ни в Go-структуре, ни в таблице `hints`.

При сохранении Go биндил JSON в структуру без такого поля, и `encoding/json`
молча выбрасывал неизвестный ключ — без ошибки, без лога, с ответом `201`.
Обратно `page_path` приходил пустым, и переход не срабатывал:

```ts
if (nextHint.page_path && nextHint.page_path !== window.location.pathname)
    navigate(nextHint.page_path)   // пусто -> условие ложно -> перехода нет
```

Дальше тур не падал, а зависал: счётчик шага уже сдвинут, а `setInterval` каждые
200 мс искал селектор, которого на текущей странице нет. Подсказка не рендерилась
(`if (!hint || !element) return null`). Выглядело как оборвавшийся онбординг.

Важно: **степы были и остаются отдельными сущностями** — таблица `hints`, своя
строка на подсказку, поле `step`, `UNIQUE (tour_version_id, step)`, ручка reorder.
Первоначальный диагноз «степы не хранятся отдельно» неверен. Терялось ровно одно
поле — привязка подсказки к странице.

## Что теперь даёт бэкенд

Колонка `hints.page_path TEXT NOT NULL DEFAULT ''` (миграция `000002`, применяется
автоматически при старте сервиса).

**Семантика:** пустая строка = «та же страница, где тур стартовал», то есть
наследует `target_path` версии. Поле опциональное, старые туры ведут себя как
раньше.

### Ручки

Изменение аддитивное, ничего не сломано.

| Ручка | Что изменилось |
|---|---|
| `POST /v1/tours/{tourId}/hints` | принимает `page_path` в теле, возвращает в `201` |
| `PATCH /v1/tours/{tourId}/hints/{hintId}` | принимает `page_path` |
| `GET /v1/tours/{tourId}/hints` | отдаёт `page_path` у каждой подсказки |
| `GET /v1/tours/{tourId}` | то же, внутри карточки тура |
| `GET /v1/versions/{versionId}` | то же |
| `POST /v1/resolve` | то же — это ручка, которая чинит баг для рантайма |
| `POST /v1/tours/{tourId}/publish` | новый повод для `422` |

Тело создания подсказки:

```json
{
  "title": "Откройте профиль",
  "content": "Здесь настройки аккаунта",
  "selector": "#profile-btn",
  "placement": "bottom",
  "page_path": "/profile",
  "spotlight": true,
  "wait_for_selector": true
}
```

### Две ловушки в контракте

**1. `PATCH` подсказки не частичный.** Контроллер биндит тело в целую структуру,
поэтому отсутствующее поле читается как пустая строка, а не «не менять».
`page_path` нужно слать при каждом обновлении, иначе он затрётся.

**2. Публикация валидирует формат пути.** Непустой `page_path` обязан начинаться
с `/`:

```json
{"details": [{"path": "hints[0].page_path", "message": "must start with /"}]}
```

Код `422`. Относительный путь никогда не совпал бы с `location.pathname` — тур
снова завис бы молча, поэтому ловим до публикации. Пустая строка проходит, это
законное наследование `target_path`.

## Что нужно сделать на фронте

### 1. Перестать хардкодить `target_path`

[`frontend/src/Hooks/useSaveScenario.ts:23`](frontend/src/Hooks/useSaveScenario.ts)

```ts
const tourId = await onboardingAPI.createTour({
  title: data.title,
  description: data.description,
  target_path: "/",   // <- игнорирует форму
  priority: 1,
  ...
});
```

Из-за этого любой тур, созданный через админку, прибит к корню: на `/dashboard`
он не стартует вовсе. Нужно брать путь из формы. Тип `SaveScenarioInput` сейчас
исключает `target_path` через `Omit` — его надо вернуть в форму и в тип.

### 2. Фолбэк на `target_path` в рантайме

[`frontend/src/Components/Onboarding/TourRunner.tsx:40`](frontend/src/Components/Onboarding/TourRunner.tsx)
и `:91` — два одинаковых блока перехода. Без фолбэка шаги с пустым `page_path`
(а это все существующие туры) поведут себя как до фикса.

Логика дублируется в `next()` и в `handleAction`, её стоит свести в одну функцию:

```ts
const pageOf = (h: TourHint) => h.page_path || tour.target_path;

const goToStep = (nextStep: number) => {
  if (nextStep >= tour.hints.length) {
    sessionStorage.removeItem(storageKey);
    setElement(null);
    onClose();
    return;
  }

  const nextHint = tour.hints[nextStep];
  changeStep(nextStep);

  const page = pageOf(nextHint);
  if (page && page !== window.location.pathname) {
    setElement(null);   // иначе на новой странице секунду висит старый якорь
    navigate(page);
  }
};
```

Обратите внимание на `setElement(null)`: сейчас в `next()` он вызывается только
на последнем шаге. При переходе между страницами старый DOM-узел остаётся в
стейте, и подсказка успевает мигнуть на неправильном элементе, пока поллинг не
найдёт новый.

Про сам поллинг: он ищет селектор бесконечно. Если элемент не появится
(опечатка в селекторе, изменилась вёрстка), тур зависнет так же тихо, как в
исходном баге. Стоит добавить таймаут и заметное поведение при его истечении —
пропустить шаг или завершить тур с ошибкой в консоль.

### 3. Превью сценария

[`frontend/src/Components/Admin/Pages/Scenarios/Scenarios.tsx:84`](frontend/src/Components/Admin/Pages/Scenarios/Scenarios.tsx)

```ts
`${item.hints[0]?.page_path || "/"}?tour=${item.id}&preview=true`
```

Здесь тоже фолбэк должен быть на `item.target_path`, а не на `/`.

### 4. Типы

[`frontend/src/types/sdk.ts:17`](frontend/src/types/sdk.ts) объявляет
`page_path: string` обязательным в `TourHint`. Формально это верно — бэк всегда
отдаёт строку. Но раньше поле приходило `undefined`, и TS этого не заметил,
потому что ответ приводится по типу без валидации. Именно это скрыло баг.

Стоит подумать о рантайм-валидации ответов (zod или руками на границе API) —
иначе следующее разъехавшееся поле опять будет искаться в проде.

## Отдельный баг, не связанный с этим фиксом

Матчинг стартовой страницы на фронте и бэке разъехался.

Бэкенд: `MatchPath` в `internal/domains/match.go` умеет wildcard'ы —
`/dashboard`, `/additem*`, `/items/*/edit`, `/*`, игнорирует хвостовой слеш и
query, если он не указан в паттерне.

Фронт: [`frontend/src/Components/Onboarding/storage.ts:31`](frontend/src/Components/Onboarding/storage.ts)

```ts
return !tour.target_path || tour.target_path === path;
```

Строгое сравнение. Тур с `target_path: "/items/*"` бэк считает подходящим для
`/items/42`, а фронт — нет. Нужно решить, где живёт правда, и свести к одной
реализации.

К `page_path` это не относится: туда wildcard не годится в принципе — по
`/items/*` нельзя сделать `navigate`, нужен конкретный путь. Поэтому валидация
`page_path` проверяет только ведущий `/`.
