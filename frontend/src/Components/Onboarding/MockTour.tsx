import type { Tour } from "../../types/sdk";

export const mockTour: Tour = {
    id: "1",
    title: "Первое объявление",
    hints: [
        {
            id: "1",
            title: "Создайте объявление",
            content: "Нажмите сюда, чтобы разместить первое объявление",
            selector: "[data-onboarding='create-ad']",
            placement: "bottom",
            spotlight: true,
        },

        {
            id: "2",
            title: "Добавьте фото",
            content: "Добавьте фотографии товара",
            selector: "[data-onboarding='photo-upload']",
            placement: "right",
            spotlight: true,
        },
    ],
    description: "",
    status: "published",
    target_path: ""
};
