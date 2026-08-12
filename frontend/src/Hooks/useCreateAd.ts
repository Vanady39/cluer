import { useForm } from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import type { InferType } from "yup";
import { createAdSchema } from "../Components/Classifier/Pages/AddItem/schema";
import { trackOnboardingGoal } from "../Components/Onboarding/goal";

export function useCreateAd() {
  const createForm = useForm({
    resolver: yupResolver(createAdSchema),
    defaultValues: {
      category: "",
      title: "",
      description: "",
      city: "",
    },
  });

  const submitAd = async (
    data: InferType<typeof createAdSchema>,
    files: File[],
    onSuccess?: () => void,
    onError?: () => void,
  ) => {
    try {
      const formData = new FormData();
      Object.entries(data).forEach(([key, value]) => {formData.append(key, String(value));});
      files.forEach((file) => {formData.append("images", file);});

      console.log("Отправка:", Object.fromEntries(formData));

      trackOnboardingGoal("listing_created", {
        title: data.title,
        category: data.category,
      });

      onSuccess?.();
    } catch (error) {
      console.error("Ошибка при создании", error);
      onError?.();
    }
  };

  return { createForm, submitAd };
}