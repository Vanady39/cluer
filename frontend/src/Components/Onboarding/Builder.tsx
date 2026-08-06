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

      console.log("Выбран элемент:", element);
      console.log("Селектор:", selector);

      if (window.opener) {
        window.opener.postMessage(
          { type: "SELECTOR_SELECTED", selector },
          "*"
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

  return (
    <div
      style={{
        position: "fixed",
        top: 20,
        left: "50%",
        transform: "translateX(-50%)",
        background: "#1677ff",
        color: "#fff",
        padding: "12px 20px",
        borderRadius: 12,
        zIndex: 99999,
        boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
        pointerEvents: "none",
      }}
    >
      Нажмите на любой элемент на этой странице, чтобы выбрать его
    </div>
  );
}

function createSelector(element: HTMLElement) {
  if (element.dataset.onboardingId) {
    return `[data-onboarding-id="${element.dataset.onboardingId}"]`;
  }

  const id = `element-${Date.now()}`;
  element.dataset.onboardingId = id;
  return `[data-onboarding-id="${id}"]`;
}