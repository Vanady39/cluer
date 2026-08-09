import { type RefObject, useEffect } from "react";

const useOutsideClick = (
  target:
    | React.RefObject<HTMLElement>
    | RefObject<HTMLDivElement | null>
    | undefined,
  handler: (event: MouseEvent) => void,
  options = ["button", "a[href]"],
  isActive?: boolean | undefined,
) => {
  useEffect(() => {
    if (!target || !handler || isActive === false) return;

    const handleClickOutside = (event: MouseEvent) => {
      const element = target.current;
      if (!element) return;
      if (element.contains(event.target as Node)) return;

      const isExcluded = options.some((selector) =>
        (event.target as Element)?.closest(selector),
      );
      if (isExcluded) return;
      handler(event);
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [target, handler, options, isActive]);
};

export default useOutsideClick;