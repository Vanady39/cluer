import { afterEach, describe, expect, it } from "vitest";
import { createSelector } from "./selector";

function mount(html: string) {
  document.body.innerHTML = html;
}

afterEach(() => {
  document.body.innerHTML = "";
});

describe("createSelector", () => {
  it("берёт data-tour с самого элемента", () => {
    mount(`<div><button data-tour="add-item">Разместить</button></div>`);
    const el = document.querySelector("button")!;

    expect(createSelector(el as HTMLElement)).toBe('[data-tour="add-item"]');
  });

  // Классы CSS-модулей содержат хеш сборки: следующий build меняет его и
  // обесточивает все сохранённые сценарии.
  it("не вставляет хешированные классы CSS-модулей", () => {
    mount(`
      <main class="_home_14skw_1">
        <div class="_home__cards_14skw_4">
          <article class="_card_1gxno_1"><img class="_card__image_1gxno_10" /></article>
        </div>
      </main>`);
    const el = document.querySelector("img")!;

    const selector = createSelector(el as HTMLElement);

    expect(selector).not.toMatch(/14skw|1gxno/);
  });

  it("цепляется за ближайший стабильный якорь у предка", () => {
    mount(`
      <section data-tour="listings">
        <div class="_cards_abcde_1">
          <article><img /></article>
        </div>
      </section>`);
    const el = document.querySelector("img")!;

    expect(createSelector(el as HTMLElement)).toMatch(/^\[data-tour="listings"\]/);
  });

  it("сохраняет осмысленные классы", () => {
    mount(`<div><span class="price">100</span></div>`);
    const el = document.querySelector("span")!;

    expect(createSelector(el as HTMLElement)).toContain("price");
  });

  // Главный инвариант: чем бы селектор ни оказался, он обязан находить ровно
  // тот элемент, по которому кликнули.
  describe("находит ровно исходный элемент", () => {
    const shapes: Array<[string, string]> = [
      [
        "карточка среди одинаковых соседей",
        `<main class="_home_14skw_1"><div class="_cards_14skw_4">
           <article class="_card_1gxno_1"><img class="_img_1gxno_2" /></article>
           <article class="_card_1gxno_1"><img class="_img_1gxno_2" id="target" /></article>
           <article class="_card_1gxno_1"><img class="_img_1gxno_2" /></article>
         </div></main>`,
      ],
      [
        "поле формы",
        `<form><label>Цена</label><input name="price" id="target" /></form>`,
      ],
      [
        "глубоко вложенный элемент без атрибутов",
        `<div><div><div><section><ul><li></li><li><b id="target">x</b></li></ul></section></div></div></div>`,
      ],
      [
        "элемент с data-testid у предка",
        `<div data-testid="panel"><div class="_x_ab12c_3"><span id="target">y</span></div></div>`,
      ],
    ];

    for (const [name, html] of shapes) {
      it(name, () => {
        mount(html);
        const target = document.getElementById("target") as HTMLElement;
        target.removeAttribute("id");

        const selector = createSelector(target);
        const found = document.querySelectorAll(selector);

        expect({ name, count: found.length, isTarget: found[0] === target }).toEqual({
          name,
          count: 1,
          isTarget: true,
        });
      });
    }
  });
});
