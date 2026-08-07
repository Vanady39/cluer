import { useEffect } from "react";

interface Props {
  onSelect?: (selector: string) => void;
}

export function Builder({ onSelect }: Props) {
  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();

      const element = event.target as HTMLElement;
      const selector = createSelector(element);

      if (window.opener) {
        window.opener.postMessage(
          {
            type: "SELECTOR_SELECTED",
            selector,
          },
          "*",
        );

        window.close();
        return;
      }

      onSelect?.(selector);
    };

    document.addEventListener("click", handleClick, true);

    return () => {
      document.removeEventListener("click", handleClick, true);
    };
  }, [onSelect]);

  return null;
}

function createSelector(element: HTMLElement) {
  if (element.dataset.tour) {
    return `[data-tour="${element.dataset.tour}"]`;
  }

  if (element.id) {
    return `#${element.id}`;
  }

  if (element.dataset.testid) {
    return `[data-testid="${element.dataset.testid}"]`;
  }

  const path: string[] = [];

  let current: HTMLElement | null = element;

  while (current && current !== document.body) {
    let selector = current.tagName.toLowerCase();

    // классы
    if (typeof current.className === "string" && current.className.trim()) {
      const classes = current.className.split(" ").filter(Boolean).slice(0, 2);

      if (classes.length) {
        selector += "." + classes.join(".");
      }
    }

    // если одинаковые элементы среди братьев
    const parent = current.parentElement;

    if (parent) {
      const siblings = Array.from(parent.children).filter(
        (child) => child.tagName === current!.tagName,
      );

      if (siblings.length > 1) {
        const index = siblings.indexOf(current) + 1;

        selector += `:nth-of-type(${index})`;
      }
    }

    path.unshift(selector);

    current = current.parentElement;
  }

  return path.join(" > ");
}