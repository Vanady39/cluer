# Платформа интерактивного онбординга — спецификация БД и контрактов

Документ предназначен для реализации бэкенда (Go + PostgreSQL) и контрактов, по которым с ним
работают SDK и админка. Каждое решение снабжено обоснованием — при реализации важно понимать
*почему* сделано так, чтобы не «упростить» то, на чём держатся инварианты.

---

## 1. Что мы строим

Система из трёх частей:

- **Админка** — продуктовая команда создаёт сценарий онбординга, настраивает шаги, публикует, смотрит аналитику.
- **SDK/виджет** — подключается тегом `<script>` к любому веб-приложению, получает опубликованный
  сценарий и показывает подсказки поверх чужого DOM.
- **Бэкенд** — хранит сценарии, решает кому и что показывать, принимает события, считает воронку, генерирует PDF.

Цикл замкнутый: админ публикует → бэк резолвит → SDK показывает → события возвращаются →
админ видит воронку и правит сценарий → новая версия.

---

## 2. Глоссарий

| Термин | Значение |
|---|---|
| **Flow** (сценарий) | Стабильная сущность: «онбординг подачи объявления». Живёт вечно, имеет неизменный `key`. |
| **Flow Version** | Конкретная редакция сценария. Иммутабельна. Ровно одна из них может быть `published`. |
| **Definition** | Тело сценария: `trigger`, `goal`, `steps[]`. Хранится как JSONB внутри версии. |
| **Step** (шаг) | Одна подсказка. **Не имеет отдельного жизненного цикла** — существует только внутри definition. |
| **Anchor** (якорь) | Набор CSS-селекторов, по которым SDK находит DOM-элемент для привязки подсказки. |
| **Subject** | Тот, кому показываем: авторизованный `userId` либо анонимный `anonId` из localStorage. |
| **Goal** | Бизнес-результат (например `ad_published`). Намеренно отделён от прохождения последнего шага. |
| **Resolve** | Запрос SDK на старте: «что показать этому пользователю здесь и сейчас». |

---

## 3. Архитектурные принципы

Эти четыре правила определяют всю остальную структуру. Нарушение любого ломает систему.

### 3.1. Единица хранения — версия, а не шаг

Шаг не имеет самостоятельного существования: его нельзя запросить, изменить или удалить отдельно.
Мы всегда читаем и пишем весь `definition` целиком.

**Почему:** SDK получает сценарий целиком одним запросом, админ сохраняет отредактированный сценарий
целиком одним действием. Дробление на строки создало бы сборку джойном при чтении и diff-логику
(что вставить / удалить / обновить / как пересчитать `order_index`) при записи — чистые накладные
расходы на несуществующую потребность.

### 3.2. Опубликованная версия иммутабельна

Не «мы стараемся её не менять» — в коде **нет операции**, которая её меняет. Любая правка создаёт новую версию.

**Почему:** каждое событие аналитики несёт `flow_version_id`. Иммутабельность делает эту ссылку
точным снимком того, что пользователь видел на экране. Если бы версию можно было править на месте,
один `UPDATE` текста задним числом обесценил бы всю накопленную по ней статистику: цифры относятся
к одному сценарию, а в базе лежит другой, и различить их нечем. Данные тихо становятся ложью.

Следствие: откат = скопировать `definition` старой версии в новую и опубликовать. История не
переписывается, а дополняется.

### 3.3. SDK реактивен, а не императивен

SDK **не ведёт** пользователя навигацией. Он ждёт, пока условия шага станут истинными (URL совпал,
элемент появился в DOM), и рисует подсказку. Ушёл пользователь в сторону — SDK молчит; вернулся —
подсказка появилась снова.

**Почему:** пользователь непредсказуем, а SDK работает поверх чужого приложения и не имеет права
дёргать его роутер.

### 3.4. Сценарий отдаётся целиком за один resolve

Бэк не диктует переходы между шагами — он их только регистрирует. Логика «какой шаг следующий»
уже в `definition`, полученном на старте.

**Почему:** запрос за каждым шагом = раундтрип 200–400 мс на каждый клик (визуально — лаг),
и обрыв сети рвёт онбординг посередине. При отдаче целиком отвалившаяся сеть ломает аналитику,
но не сам онбординг.

---

## 4. Схема базы данных

### 4.1. Почему PostgreSQL

- Нужны транзакции и уникальные ограничения для инвариантов публикации (см. 4.4) — key-value БД этого не даст.
- Нужен гибкий, часто меняющийся формат шага — `JSONB` c GIN-индексом снимает необходимость
  в миграции на каждое новое поле шага.
- Аналитика на объёмах MVP отлично считается обычными SQL-агрегатами.
- Одна БД вместо двух — меньше движущихся частей в Docker Compose.

В README честно указать post-MVP развитие: при росте потока событий — вынос `flow_events` в ClickHouse,
`flows`/`flow_versions` остаются в Postgres.

### 4.2. Расширения и общие типы

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- gen_random_uuid()
```

Все `id` — `UUID`, генерируются БД (`DEFAULT gen_random_uuid()`), кроме `step.id` внутри definition —
там короткие человекочитаемые строки (`"s1"`, `"pick-category"`), задаваемые админкой.

### 4.3. `apps` — приложение-потребитель

```sql
CREATE TABLE apps (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  public_key  TEXT NOT NULL UNIQUE,       -- уходит в <script data-app-id="...">
  allowed_origins TEXT[] NOT NULL DEFAULT '{}',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Зачем эта таблица.** Она делает решение «механизмом, подключаемым к разным веб-приложениям» —
прямое требование кейса. Без неё система жёстко привязана к одному сайту.

- `public_key` — не секрет. Он виден в исходниках страницы, поэтому даёт **только** право читать
  опубликованные сценарии и слать события. Ничего писать им нельзя.
- `allowed_origins` — белый список доменов для CORS на runtime-эндпоинтах. Защищает от того,
  что чужой сайт начнёт слать мусорные события с вашим ключом.

### 4.4. `flows` — стабильная сущность сценария

```sql
CREATE TABLE flows (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  key         TEXT NOT NULL,              -- 'post-ad-flow', стабилен между версиями
  name        TEXT NOT NULL,              -- человекочитаемое, для админки
  description TEXT NOT NULL DEFAULT '',
  enabled     BOOLEAN NOT NULL DEFAULT false,
  priority    INT NOT NULL DEFAULT 0,     -- больше = важнее при конкуренции
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at TIMESTAMPTZ,
  UNIQUE (app_id, key)
);

CREATE INDEX flows_app_enabled_idx ON flows (app_id) WHERE enabled AND archived_at IS NULL;
```

Разбор колонок:

- **`key`** — стабильный идентификатор для аналитики и внешних ссылок. `id` тоже стабилен, но `key`
  читаем человеком и переживает пересоздание БД при демо.
- **`enabled`** — рубильник, **независимый от версий**. Это и есть тот самый «feature-флаг», только
  на верном уровне абстракции. Он нужен, чтобы отличать «онбординг остановлен» от «никогда не публиковался»:
  без него единственный способ выключить флоу — заархивировать опубликованную версию, и тогда
  эти два состояния становятся неразличимы.
- **`priority`** — при `resolve` под контекст может подойти несколько флоу. Показывать надо **один**
  (два оверлея одновременно — сломанный UX). Разрешаем по `priority DESC, created_at ASC` —
  детерминированно и настраиваемо из админки.
- **`archived_at`** — мягкое удаление. Жёсткий `DELETE` уничтожил бы историю аналитики.

### 4.5. `flow_versions` — иммутабельный лог редакций

```sql
CREATE TABLE flow_versions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  flow_id       UUID NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  version       INT  NOT NULL,
  status        TEXT NOT NULL CHECK (status IN ('draft','published','archived')),
  definition    JSONB NOT NULL,
  created_by    TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at  TIMESTAMPTZ,
  archived_at   TIMESTAMPTZ,
  UNIQUE (flow_id, version),
  CHECK (status <> 'published' OR published_at IS NOT NULL)
);

-- ИНВАРИАНТ 1: максимум одна опубликованная версия на флоу
CREATE UNIQUE INDEX flow_versions_one_published
  ON flow_versions (flow_id) WHERE status = 'published';

-- ИНВАРИАНТ 2: максимум один черновик на флоу
CREATE UNIQUE INDEX flow_versions_one_draft
  ON flow_versions (flow_id) WHERE status = 'draft';

-- поиск по содержимому: «в каких сценариях используется селектор X»
CREATE INDEX flow_versions_definition_gin
  ON flow_versions USING GIN (definition jsonb_path_ops);
```

**Ключевое решение: в `flows` НЕТ колонки `current_version_id`.**

Актуальность не хранится указателем, а выводится: опубликованная версия флоу X — это строка
с `flow_id = X AND status = 'published'`, а единственность гарантирует частичный уникальный индекс.

Почему не указателем: тогда правда о состоянии лежит в двух местах (`flows.current_version_id`
и `flow_versions.status`), их надо синхронизировать при каждой публикации, и рано или поздно они
разъедутся — указатель будет вести на версию со статусом `archived`. БД такое рассогласование
не поймает. С частичным индексом оно невозможно по построению.

`flow_versions_one_draft` нужен, чтобы админка не получила неоднозначность «какой черновик открывать».

**Что внутри JSONB, а что снаружи.** Критерий: если поле участвует в `WHERE` каждого запроса
или в ограничении целостности — оно колонка. Поэтому `status` не внутри `definition`: на нём висит
уникальный индекс, а JSONB такой гарантии дать не может. А тексты, селекторы и `advanceOn` нужны
только для отрисовки — они внутри.

Пример поиска по GIN-индексу (для функции «где используется этот селектор» в админке):

```sql
SELECT fv.id, f.key
FROM flow_versions fv JOIN flows f ON f.id = fv.flow_id
WHERE fv.definition @> '{"steps":[{"anchor":{"primary":"#category"}}]}';
```

### 4.6. `flow_progress` — состояние прохождения

```sql
CREATE TABLE flow_progress (
  app_id          UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  subject_id      TEXT NOT NULL,
  flow_id         UUID NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  flow_version_id UUID NOT NULL REFERENCES flow_versions(id),
  current_step_id TEXT,
  status          TEXT NOT NULL CHECK (status IN ('in_progress','completed','dismissed')),
  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at     TIMESTAMPTZ,
  PRIMARY KEY (app_id, subject_id, flow_id)
);

CREATE INDEX flow_progress_flow_status_idx ON flow_progress (flow_id, status);
```

**Это единственная мутабельная таблица в системе, и так задумано.**

- **Почему не выводим прогресс из событий на лету.** `resolve` должен ответить за единицы миллисекунд;
  реконструкция состояния агрегацией лога на каждый запрос — лишняя работа. Держим денормализованную
  проекцию.
- **Почему прогресс на бэке, а не в localStorage.** localStorage чистится, не переживает смену
  устройства, и ему нельзя доверять (пользователь может подделать «я уже прошёл»). На фронте
  localStorage допустим только как кэш для мгновенной отрисовки до ответа сети.
- **Почему PK без `flow_version_id`.** Один subject проходит один флоу один раз. Если во время
  прохождения вышла новая версия — см. правило миграции в 6.2.
- **`status`** различает «идёт», «дошёл до конца шагов» и «закрыл крестиком». Третье — важный
  негативный сигнал для продуктовой команды.

### 4.7. `flow_events` — сырой лог аналитики

```sql
CREATE TABLE flow_events (
  id              BIGSERIAL PRIMARY KEY,
  app_id          UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  event_key       TEXT NOT NULL,           -- UUID, сгенерированный клиентом
  flow_id         UUID NOT NULL,
  flow_version_id UUID NOT NULL REFERENCES flow_versions(id),
  step_id         TEXT,                    -- 's2'; NULL для событий уровня флоу
  session_id      TEXT NOT NULL,
  subject_id      TEXT NOT NULL,
  type            TEXT NOT NULL,
  occurred_at     TIMESTAMPTZ NOT NULL,    -- время на клиенте
  received_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  payload         JSONB NOT NULL DEFAULT '{}',
  UNIQUE (app_id, event_key)
);

CREATE INDEX flow_events_version_type_idx ON flow_events (flow_version_id, type);
CREATE INDEX flow_events_subject_idx ON flow_events (app_id, subject_id, occurred_at);
```

**Здесь строки, а не JSONB — и это принципиально противоположно решению по `flow_versions`.**
Причина: события мы фильтруем, группируем и агрегируем по каждому полю. Всё, что участвует
в `GROUP BY` и `WHERE`, обязано быть колонкой.

Разбор:

- **`step_id` — TEXT, а НЕ внешний ключ.** Шаг не существует как строка в БД, он живёт внутри
  документа версии. Целостность обеспечивается тем, что `flow_version_id` жёстко фиксирует,
  в каком именно документе искать этот `s2`.
- **`UNIQUE (app_id, event_key)`** — идемпотентность. SDK шлёт события батчами с retry при обрыве
  сети; без этого ограничения ретрай удвоит метрики. Вставка через `ON CONFLICT (app_id, event_key) DO NOTHING`.
- **`occurred_at` vs `received_at`** — часы клиента врут и могут быть сдвинуты. Воронку считаем
  по `occurred_at` (порядок действий пользователя), но при расхождении больше суток доверяем
  `received_at`. В отчёте использовать `received_at` для фильтра по периоду.
- **`flow_version_id` обязателен во всех событиях.** Без него воронка бессмысленна: показы
  до и после правки текстов смешаются, и сравнивать будет нечего. Это главное правило аналитики.

**Словарь `type`:**

| type | step_id | Смысл |
|---|---|---|
| `flow_started` | NULL | Сценарий начат (показан первый шаг) |
| `step_shown` | есть | Подсказка отрисована на экране |
| `step_completed` | есть | Сработал `advanceOn` — пользователь сделал нужное действие |
| `step_skipped` | есть | Шаг пропущен (`onMissing: "skip"`) |
| `anchor_missing` | есть | Элемент не найден за `waitForAnchorMs`. **Сигнал «сценарий сломан»** |
| `flow_completed` | NULL | Пройден последний шаг |
| `flow_dismissed` | NULL | Пользователь закрыл онбординг крестиком |
| `goal_reached` | NULL | Достигнут бизнес-результат |

**`anchor_missing` — не служебный лог, а продуктовая фича.** По нему в админке сценарий
подсвечивается как сломанный: вёрстка хоста изменилась, селектор не находится. Без этого события
онбординг молча перестаёт работать, и никто не узнает.

**`goal_reached` ≠ `flow_completed`.** Пользователь может пройти все подсказки и не опубликовать
объявление, или опубликовать, закрыв онбординг на середине. Разница между этими метриками —
и есть ответ на вопрос «помогает ли сценарий или его просто вежливо кликают насквозь».
Если их слить, метрика начнёт измерять сама себя.

### 4.8. `reports` — сгенерированные PDF (опционально)

```sql
CREATE TABLE reports (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  flow_id     UUID NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  period_from TIMESTAMPTZ NOT NULL,
  period_to   TIMESTAMPTZ NOT NULL,
  status      TEXT NOT NULL CHECK (status IN ('pending','ready','failed')),
  file_path   TEXT,
  error       TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Для MVP допустима синхронная генерация с отдачей файла прямо в ответе — данных мало. Таблицу
завести, если генерация уйдёт в фон; тогда фронт поллит `status`.

---

## 5. Формат `definition`

### 5.1. Пример

```json
{
  "schemaVersion": 1,
  "trigger": {
    "urlPattern": "/*",
    "audience": {
      "allOf": [
        { "prop": "isNewUser", "op": "eq", "value": true },
        { "prop": "adsCount", "op": "lt", "value": 1 }
      ]
    },
    "showOnce": true
  },
  "goal": { "event": "ad_published", "windowMinutes": 60 },
  "settings": {
    "allowDismiss": true,
    "showProgress": true,
    "theme": { "accentColor": "#00AAFF" }
  },
  "steps": [
    {
      "id": "s1",
      "match": { "urlPattern": "/*", "waitForAnchorMs": 5000 },
      "anchor": {
        "primary": "[data-onboarding-id='post-ad-btn']",
        "fallbacks": ["#postAdBtn", "header .btn--primary"]
      },
      "ui": { "type": "tooltip", "placement": "bottom", "spotlight": true },
      "content": { "title": "Начните здесь", "body": "Нажмите, чтобы разместить объявление." },
      "advanceOn": { "type": "click", "selector": "[data-onboarding-id='post-ad-btn']" },
      "onMissing": "skip"
    },
    {
      "id": "s2",
      "match": { "urlPattern": "/additem*", "waitForAnchorMs": 8000 },
      "anchor": { "primary": "[data-onboarding-id='category-select']", "fallbacks": ["#category"] },
      "ui": { "type": "tooltip", "placement": "right", "spotlight": true },
      "content": { "title": "Выберите категорию", "body": "От неё зависит набор полей." },
      "advanceOn": { "type": "change", "selector": "#category" },
      "onMissing": "abort"
    }
  ]
}
```

### 5.2. Разбор полей

**`schemaVersion`** — дешёвая страховка. Когда формат шага изменится, старые версии останутся
читаемыми, а код будет знать, по каким правилам их разбирать.

**`trigger`** — условие старта **всего сценария**:
- `urlPattern` — glob-подобный паттерн (`*` = любые символы). Не regex: админ пишет его руками,
  regex он напишет с ошибкой. Компилировать в regex на бэке при валидации.
- `audience` — правила таргетинга по `props`, которые SDK передаёт в `resolve`. Операторы:
  `eq`, `neq`, `lt`, `gt`, `in`, `exists`. Комбинаторы `allOf` / `anyOf`.
  **Вычисляется на бэке, а не на фронте** — логику таргетинга нельзя отдавать в браузер, её там подменят.
- `showOnce` — не показывать повторно тем, кто уже завершил или закрыл.

**`goal`** — бизнес-результат. `windowMinutes` ограничивает окно атрибуции: публикация объявления
через три дня после онбординга вряд ли его заслуга.

**`steps[].match`** — где шаг активен. `waitForAnchorMs` — сколько ждать появления элемента
(в SPA DOM приходит асинхронно), прежде чем считать якорь ненайденным.

**`steps[].anchor`** — `primary` и `fallbacks` это **альтернативные селекторы одного и того же
элемента**, а не начало и конец перехода. Сценарий живёт дольше вёрстки: сегодня у кнопки есть
`data-onboarding-id`, завтра фронтендер хоста переписал компонент. SDK пробует по очереди;
не нашёл ни одного — шлёт `anchor_missing`.

Приоритет при генерации селектора визуальным пикером:
`[data-onboarding-id]` → `[data-testid]` → `#id` → короткий структурный путь.

**`steps[].advanceOn`** — условие выхода из шага. Типы:

| type | Поля | Когда |
|---|---|---|
| `click` | `selector` | Клик по элементу |
| `change` | `selector` | Изменение значения поля |
| `urlChange` | `urlPattern` | Сам переход есть событие |
| `customEvent` | `name` | Хост дёрнул `window.onboarding.track('photo_uploaded')` |
| `nextButton` | — | Информационный шаг, кнопка «Далее» в тултипе |

**`steps[].onMissing`** — `skip` (пропустить, идти дальше) или `abort` (прервать флоу).
Конфигурируемо, потому что для необязательной подсказки пропуск нормален, а для критического шага — нет.

### 5.3. Где кодируются переходы между шагами

Отдельной сущности «переход» **нет**. Ребро между шагом N и N+1 выводится из двух полей:
`advanceOn` шага N говорит «когда уходим», `match` шага N+1 говорит «где мы окажемся».
Путь по сайту закодирован как последовательность пар (URL-паттерн, элемент).

SDK **не знает граф сайта** и не должен: иначе знание о конкретном приложении зашивается
в универсальный механизм и теряется переносимость.

**Post-MVP (в README как план развития):** при появлении ветвления шаги становятся узлами
конечного автомата с явным ребром:

```json
"next": [
  { "if": { "prop": "hasAds", "op": "eq", "value": true }, "goTo": "s5" },
  { "default": true, "goTo": "s3" }
]
```

Формат расширяется без ломки, потому что линейный список — частный случай графа, где у каждого
узла ровно одно безусловное ребро. **В MVP не реализовывать.**

### 5.4. Валидация при публикации

Бэк **не принимает definition вслепую**. При публикации Go парсит его в типизированную структуру
и проверяет:

1. `steps` непуст; все `step.id` уникальны и соответствуют `^[a-z0-9-]{1,32}$`.
2. `anchor.primary` непустой и является синтаксически валидным CSS-селектором.
3. `advanceOn.type` из допустимого набора; для `click`/`change` задан `selector`,
   для `customEvent` — `name`.
4. Все `urlPattern` успешно компилируются.
5. `goal.event` задан и непустой.
6. `ui.type` ∈ {`tooltip`, `modal`, `spotlight`}; `placement` ∈ {`top`,`bottom`,`left`,`right`,`auto`}.
7. `schemaVersion` поддерживается текущим кодом.

Иначе одна опечатка в админке кладёт онбординг в проде, и узнаете вы об этом от пользователей.
Это же — готовая цель для юнит-тестов (отдельный критерий оценки в кейсе).

Держать JSON Schema в `/schemas/flow-definition.v1.json`, проверять через
`santhosh-tekuri/jsonschema` **в дополнение** к парсингу в структуру.

### 5.5. Go-типы

```go
type Definition struct {
    SchemaVersion int      `json:"schemaVersion"`
    Trigger       Trigger  `json:"trigger"`
    Goal          Goal     `json:"goal"`
    Settings      Settings `json:"settings"`
    Steps         []Step   `json:"steps"`
}

type Step struct {
    ID        string    `json:"id"`
    Match     Match     `json:"match"`
    Anchor    Anchor    `json:"anchor"`
    UI        StepUI    `json:"ui"`
    Content   Content   `json:"content"`
    AdvanceOn AdvanceOn `json:"advanceOn"`
    OnMissing string    `json:"onMissing"` // skip | abort
}

type Anchor struct {
    Primary   string   `json:"primary"`
    Fallbacks []string `json:"fallbacks,omitempty"`
}

// Работаем с типизированной структурой, в БД лежит JSONB.
func (d Definition) Value() (driver.Value, error) { return json.Marshal(d) }

func (d *Definition) Scan(src any) error {
    b, ok := src.([]byte)
    if !ok {
        return fmt.Errorf("definition: expected []byte, got %T", src)
    }
    return json.Unmarshal(b, d)
}
```

Никаких `map[string]any` в бизнес-логике.

---

## 6. API: Runtime (для SDK)

Аутентификация: заголовок `X-App-Key: <apps.public_key>`.
CORS: `Access-Control-Allow-Origin` только из `apps.allowed_origins`.
Формат: JSON, `Content-Type: application/json`.

### 6.1. `POST /v1/resolve`

Единственный запрос, который SDK делает при инициализации. Отвечает на вопрос кейса
«когда и кому показать конкретный онбординг».

**Request**

```json
{
  "url": "https://demo.local/additem?cat=1",
  "subjectId": "anon_8f3d...",
  "sessionId": "sess_2b91...",
  "props": { "isNewUser": true, "adsCount": 0, "locale": "ru" }
}
```

- `subjectId` — `userId` хоста, если есть; иначе `anonId`, сгенерированный SDK и сохранённый
  в localStorage. Формат — произвольная строка до 128 символов.
- `props` — произвольные атрибуты пользователя от хост-приложения, по ним работает `trigger.audience`.

**Response 200**

```json
{
  "flowId": "3f2a...",
  "flowKey": "post-ad-flow",
  "flowVersionId": "9c1e...",
  "version": 7,
  "currentStepId": "s2",
  "definition": { "...весь документ..." }
}
```

**Response 204 No Content** — показывать нечего. Это нормальный, ожидаемый ответ, не ошибка.

**Алгоритм на бэке:**

1. Найти `apps` по `public_key`, проверить `Origin`.
2. Выбрать флоу приложения с `enabled = true AND archived_at IS NULL`, имеющие `published`-версию.
3. Отфильтровать по `trigger.urlPattern` (матчинг `path + search` из `url`).
4. Отфильтровать по `trigger.audience` относительно `props`.
5. Отбросить те, где `showOnce = true` и в `flow_progress` есть `completed`/`dismissed`.
6. Из оставшихся взять первый по `priority DESC, created_at ASC`.
7. Прочитать/создать `flow_progress`, вернуть `currentStepId`.

**Кэширование.** Версии иммутабельны, поэтому `definition` кэшируется в памяти по `flow_version_id`
навсегда — инвалидация не нужна вовсе. Кэшировать надо именно документ, а не результат резолва
(он зависит от пользователя).

### 6.2. Правило миграции при смене версии

Если у subject есть `in_progress` прогресс по версии, которая больше не `published`:

- прогресс завершается как `dismissed` c `payload.reason = "version_superseded"`;
- заводится новый прогресс по актуальной версии с первого шага.

**Почему не переносить позицию:** `step_id` в новой версии может отсутствовать или означать
другое. Смешивание сделало бы воронку недостоверной. Лучше честно начать заново, чем получить
события, не соответствующие ни одной реальной версии.

### 6.3. `POST /v1/events`

Батч событий. Вызывается асинхронно, **не блокирует UI**.

**Request**

```json
{
  "subjectId": "anon_8f3d...",
  "sessionId": "sess_2b91...",
  "events": [
    {
      "eventKey": "b1e7...",
      "type": "step_shown",
      "flowId": "3f2a...",
      "flowVersionId": "9c1e...",
      "stepId": "s2",
      "occurredAt": "2026-08-07T10:15:03.412Z",
      "payload": {}
    }
  ]
}
```

**Response 202 Accepted**

```json
{ "accepted": 4, "duplicates": 1 }
```

- `eventKey` — UUID, генерируемый **клиентом**. Основа идемпотентности: retry после обрыва сети
  не удвоит метрики. Вставка `ON CONFLICT (app_id, event_key) DO NOTHING`.
- Лимит: до 50 событий в батче, тело до 256 КБ.
- Побочный эффект: события `step_completed`, `flow_completed`, `flow_dismissed`, `goal_reached`
  обновляют `flow_progress` в той же транзакции.
- SDK отправляет батч по таймеру (~2 с) и через `navigator.sendBeacon` на `visibilitychange`,
  чтобы не терять события при закрытии вкладки.

### 6.4. `GET /v1/health`

`200 {"status":"ok","db":"ok"}` — для healthcheck в Docker Compose.

---

## 7. API: Admin

Аутентификация: `Authorization: Bearer <jwt>`. Для MVP допустим один захардкоженный
администратор с логином/паролем из env — но эндпоинты должны быть закрыты, а не открыты миру.

### 7.1. Управление сценариями

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/v1/admin/flows?appId=` | Список сценариев со статусом текущей версии |
| `POST` | `/v1/admin/flows` | Создать сценарий (сразу с пустым draft v1) |
| `GET` | `/v1/admin/flows/{id}` | Карточка: флоу + published + draft |
| `PATCH` | `/v1/admin/flows/{id}` | Изменить `name`, `enabled`, `priority` |
| `DELETE` | `/v1/admin/flows/{id}` | Мягкое удаление (`archived_at`) |
| `GET` | `/v1/admin/flows/{id}/versions` | История версий (без тел, только метаданные) |
| `GET` | `/v1/admin/versions/{versionId}` | Конкретная версия с `definition` |

### 7.2. `PUT /v1/admin/flows/{id}/draft`

Создать или перезаписать черновик. Черновик — **единственная мутабельная версия**.

```json
{ "definition": { "...": "..." } }
```

Логика: если `draft` есть — `UPDATE definition`; если нет — `INSERT` с
`version = COALESCE(MAX(version),0) + 1, status = 'draft'`.

Валидация по 5.4 выполняется и здесь, но **не блокирует сохранение** — возвращается
в поле `warnings`, чтобы админ мог сохранить незаконченную работу. Блокирует только публикация.

### 7.3. `POST /v1/admin/flows/{id}/publish`

Главная операция. **Транзакция обязательна.**

```sql
BEGIN;

-- 1. Блокируем флоу: два одновременных нажатия «Опубликовать» иначе упрутся
--    в уникальный индекс, и одно упадёт с сырой ошибкой БД вместо внятного ответа.
SELECT id FROM flows WHERE id = $1 FOR UPDATE;

-- 2. Архивируем текущую опубликованную (если есть)
UPDATE flow_versions
   SET status = 'archived', archived_at = now()
 WHERE flow_id = $1 AND status = 'published';

-- 3. Публикуем черновик
UPDATE flow_versions
   SET status = 'published', published_at = now()
 WHERE flow_id = $1 AND status = 'draft'
RETURNING id, version;

COMMIT;
```

Перед транзакцией — полная валидация definition (5.4). Невалидный definition → `422` со списком ошибок.

**Ответ 200:** `{ "flowVersionId": "...", "version": 8, "publishedAt": "..." }`

После публикации черновика больше нет — следующая правка создаст новый draft (v9)
копированием definition из опубликованной версии.

### 7.4. `POST /v1/admin/flows/{id}/rollback`

```json
{ "toVersionId": "..." }
```

**Реализация: не переписываем историю, а добавляем к ней.** Копируем `definition` указанной версии
в новую версию с максимальным номером и публикуем её тем же алгоритмом, что и 7.3.

Почему так: статистика версии 5 остаётся принадлежать версии 5 и не смешивается с новым трафиком.
Если бы мы «вернули статус» старой строке, события до и после отката слились бы в одну выборку.

### 7.5. `GET /v1/admin/flows/{id}/analytics`

Параметры: `versionId` (опционально, по умолчанию — published), `from`, `to`.

```json
{
  "flowVersionId": "9c1e...",
  "version": 7,
  "period": { "from": "...", "to": "..." },
  "totals": {
    "started": 1240,
    "completed": 612,
    "dismissed": 388,
    "goalReached": 501,
    "completionRate": 0.494,
    "goalRate": 0.404,
    "medianDurationSec": 74
  },
  "funnel": [
    { "order": 1, "stepId": "s1", "title": "Начните здесь",
      "shown": 1240, "completed": 1105, "skipped": 0, "anchorMissing": 0, "dropoff": 0.109 },
    { "order": 2, "stepId": "s2", "title": "Выберите категорию",
      "shown": 1105, "completed": 780, "skipped": 12, "anchorMissing": 41, "dropoff": 0.294 }
  ],
  "health": { "brokenAnchors": ["s2"] }
}
```

**Как достать порядок шагов из JSONB** (нужен, чтобы воронка шла в правильной последовательности):

```sql
SELECT ord, step->>'id' AS step_id, step->'content'->>'title' AS title
FROM flow_versions fv,
     LATERAL jsonb_array_elements(fv.definition->'steps') WITH ORDINALITY AS t(step, ord)
WHERE fv.id = $1
ORDER BY ord;
```

**Агрегация событий:**

```sql
SELECT step_id,
       COUNT(DISTINCT subject_id) FILTER (WHERE type = 'step_shown')     AS shown,
       COUNT(DISTINCT subject_id) FILTER (WHERE type = 'step_completed') AS completed,
       COUNT(DISTINCT subject_id) FILTER (WHERE type = 'step_skipped')   AS skipped,
       COUNT(*)                   FILTER (WHERE type = 'anchor_missing') AS anchor_missing
FROM flow_events
WHERE flow_version_id = $1
  AND received_at BETWEEN $2 AND $3
  AND step_id IS NOT NULL
GROUP BY step_id;
```

Join этих двух результатов делать в Go — так проще, чем городить один большой запрос,
и порядок шагов гарантированно берётся из definition, а не из данных.

**`COUNT(DISTINCT subject_id)`, а не `COUNT(*)`** — иначе пользователь, вернувшийся на шаг
дважды, посчитается как два. Воронка меряет людей, а не показы.

### 7.6. `POST /v1/admin/flows/{id}/report`

Генерация PDF (прямое требование кейса).

Request: `{ "versionId": "...", "from": "...", "to": "..." }`

Response: `200` с `Content-Type: application/pdf` и
`Content-Disposition: attachment; filename="onboarding-post-ad-flow-v7.pdf"`.

Содержимое отчёта: название и версия сценария, период, сводные метрики, таблица воронки,
столбчатая диаграмма конверсии по шагам, блок «проблемы» (шаги с `anchor_missing`).

Библиотека: `github.com/johnfercher/maroto/v2` (чистый Go, без внешних зависимостей —
проще для Docker Compose). Альтернатива, если нужна сложная вёрстка: HTML-шаблон → `chromedp`,
но тянет headless Chrome в образ.

### 7.7. Формат ошибок

Единый для всех эндпоинтов:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Definition is invalid",
    "details": [
      { "path": "steps[1].anchor.primary", "message": "must not be empty" }
    ]
  }
}
```

Коды: `UNAUTHORIZED` (401), `FORBIDDEN` (403), `NOT_FOUND` (404), `CONFLICT` (409,
например публикация при отсутствующем черновике), `VALIDATION_FAILED` (422),
`RATE_LIMITED` (429), `INTERNAL` (500).

---

## 8. Контракт визуального пикера (админка ↔ тестовый сайт)

Это то, что превращает админку из «поля для ввода CSS-селектора» в инструмент для непрограммиста.

Админка открывает тестовый сайт в `<iframe>` с `?onboarding_editor=1`. SDK, увидев параметр,
переходит в режим редактора: подсвечивает элементы под курсором, по клику вычисляет селектор
и отдаёт наверх через `postMessage`.

```js
// iframe → админка
{
  source: "onboarding-sdk",
  type: "element-picked",
  payload: {
    primary: "[data-onboarding-id='category-select']",
    fallbacks: ["#category", "form.add-item select:nth-of-type(2)"],
    preview: { text: "Категория", tagName: "SELECT", rect: { x: 320, y: 210, w: 280, h: 40 } }
  }
}

// админка → iframe
{ source: "onboarding-admin", type: "enter-picker-mode" }
{ source: "onboarding-admin", type: "exit-picker-mode" }
{ source: "onboarding-admin", type: "preview-flow", payload: { definition: {...} } }
```

Обязательно проверять `event.origin` на обеих сторонах.

Работает и кросс-доменно — потому что SDK на той стороне «свой».

---

## 9. Границы MVP

**Реализуем:**
- Один тестовый сайт-классифайд с 2–3 экранами и разметкой `data-onboarding-id`.
- Линейные сценарии, типы шагов `tooltip` и `modal`, спотлайт.
- Админка: список, редактор шагов, визуальный пикер, публикация, откат, аналитика, PDF.
- Резолв с таргетингом по `urlPattern` + `audience`.
- Аналитика: воронка по шагам, completion rate, goal rate, детект сломанных якорей.

**Явно вне MVP (перечислить в README как план развития):**
- Ветвление (`next[]`) и A/B-тест двух версий одного флоу.
- Процентная раскатка и frequency capping.
- Локализация контента шагов.
- Вынос событий в ClickHouse.
- AI-ассистент: генерация черновика сценария по текстовому описанию цели и реестру якорей
  страницы; предложение гипотез по шагу с наибольшим отвалом.

---

## 10. Чек-лист инвариантов

Проверить юнит-тестами — это самые важные участки кода:

1. Опубликованная версия не изменяется ни одной операцией. Правка → новая версия.
2. У флоу максимум одна `published` и максимум один `draft` (гарантия — частичные уникальные индексы).
3. Публикация невалидного definition невозможна.
4. Публикация атомарна и защищена `SELECT ... FOR UPDATE`.
5. Повторная отправка одного `eventKey` не создаёт второй записи.
6. Каждое событие несёт `flow_version_id`.
7. `resolve` возвращает не более одного флоу.
8. `resolve` возвращает 204, а не ошибку, когда показывать нечего.
9. `goal_reached` учитывается отдельно от `flow_completed`.
10. Прогресс по устаревшей версии не переносится по номеру шага, а начинается заново.
