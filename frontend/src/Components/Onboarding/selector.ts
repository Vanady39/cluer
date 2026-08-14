// Генератор селекторов для билдера. Главное требование к результату — пережить
// пересборку фронтенда: классы CSS-модулей несут в себе хеш сборки
// (_home__cards_14skw_4), и селектор, склеенный из них, перестаёт находить
// элемент после первого же `vite build`. Поэтому хешированные классы
// отбрасываются, а путь обрывается на ближайшем стабильном якоре.

const MAX_DEPTH = 5;

// _local_14skw_4 — формат CSS-модулей vite: ведущее подчёркивание либо
// суффикс из хеша и порядкового номера.
const HASHED_CLASS = /^_|_[a-z0-9]{4,}_\d+$/i;

function escapeToken(value: string): string {
  return typeof CSS !== "undefined" && typeof CSS.escape === "function"
    ? CSS.escape(value)
    : value.replace(/["\\\]]/g, "\\$&");
}

function escapeValue(value: string): string {
  return value.replace(/["\\]/g, "\\$&");
}

function stableClasses(element: Element): string[] {
  if (typeof element.className !== "string") return [];

  return element.className
    .split(/\s+/)
    .filter(Boolean)
    .filter((name) => !HASHED_CLASS.test(name));
}

// Атрибуты, которые ставит человек, а не сборщик: они переживают ребилд и
// переезд вёрстки.
function anchorSelector(element: HTMLElement): string | null {
  if (element.dataset.tour) {
    return `[data-tour="${escapeValue(element.dataset.tour)}"]`;
  }
  if (element.dataset.testid) {
    return `[data-testid="${escapeValue(element.dataset.testid)}"]`;
  }
  if (element.id && !HASHED_CLASS.test(element.id)) {
    return `#${escapeToken(element.id)}`;
  }

  const tag = element.tagName.toLowerCase();
  const name = element.getAttribute("name");
  if (name) return `${tag}[name="${escapeValue(name)}"]`;

  const label = element.getAttribute("aria-label");
  if (label) return `${tag}[aria-label="${escapeValue(label)}"]`;

  return null;
}

// Шаг пути: тег плюс уцелевшие классы, а :nth-of-type добавляется только если
// без него шаг не отличает элемент от соседей.
function stepSelector(element: HTMLElement): string {
  const tag = element.tagName.toLowerCase();
  const classes = stableClasses(element).slice(0, 2).map(escapeToken);
  let step = classes.length ? `${tag}.${classes.join(".")}` : tag;

  const parent = element.parentElement;
  if (!parent) return step;

  const matching = Array.from(parent.children).filter((child) =>
    child.matches(step),
  );
  if (matching.length > 1) {
    const sameTag = Array.from(parent.children).filter(
      (child) => child.tagName === element.tagName,
    );
    step += `:nth-of-type(${sameTag.indexOf(element) + 1})`;
  }
  return step;
}

// Селектор из одних тегов (`main > div > img`) может быть уникален прямо
// сейчас и развалиться от второй карточки в выдаче. Останавливаемся только на
// том, что содержит якорь, id или уцелевший класс.
function isSpecific(selector: string): boolean {
  return /[[#.]/.test(selector);
}

function matchesOnly(selector: string, element: HTMLElement): boolean {
  try {
    const found = document.querySelectorAll(selector);
    return found.length === 1 && found[0] === element;
  } catch {
    return false;
  }
}

export function createSelector(element: HTMLElement): string {
  const direct = anchorSelector(element);
  if (direct && matchesOnly(direct, element)) return direct;

  const path: string[] = [];
  let current: HTMLElement | null = element;
  let depth = 0;

  while (current && current !== document.body && depth < MAX_DEPTH) {
    if (current !== element) {
      const anchor = anchorSelector(current);
      // Якорь у предка обрывает путь: всё, что выше, для устойчивости уже
      // не нужно и только добавляет хрупкости.
      if (anchor) {
        path.unshift(anchor);
        const anchored = path.join(" > ");
        if (matchesOnly(anchored, element)) return anchored;
        path.shift();
      }
    }

    path.unshift(stepSelector(current));

    const candidate = path.join(" > ");
    if (isSpecific(candidate) && matchesOnly(candidate, element)) {
      return candidate;
    }

    current = current.parentElement;
    depth++;
  }

  return path.join(" > ");
}
