import { useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import type { InferType } from "yup";
import { useTourLoader } from "./useTourLoader";
import { useSaveScenario } from "./useSaveScenario";
import type { Audience, TourHint, TriggerType } from "../types/sdk";
import { scenarioSchema } from "../Components/Admin/Pages/AddScenarios/schema";

type ScenarioHint = InferType<typeof scenarioSchema>["hints"][number];

const createEmptyHint = (step: number): ScenarioHint => ({
  id: `new-${Date.now()}-${step}`,
  step,
  title: step === 1 ? "Первый шаг" : "",
  content: "",
  selector: "",
  placement: "bottom",
  page_path: "/",
  spotlight: true,
  wait_for_selector: false,
  media_url: "",
});

export function useCreateScenario(editId: string | null) {
  const { data: loadedTour } = useTourLoader(editId);
  const [hasLoaded, setHasLoaded] = useState(false);
  const selectedHintIndex = useRef<number | null>(null);
  const saveMutation = useSaveScenario(editId);

  const createForm = useForm<InferType<typeof scenarioSchema>>({
    resolver: yupResolver(scenarioSchema),
    defaultValues: {
      title: "",
      description: "",
      trigger_type: "on_load",
      audience: {
        show_once: true,
        max_shows: 1,
        only_new: false,
      },
      hints: [createEmptyHint(1)],
    },
  });

  useEffect(() => {
    if (!loadedTour || hasLoaded) return;

    const loadedHints: ScenarioHint[] = loadedTour.hints?.length
      ? loadedTour.hints.map((hint: TourHint) => ({
          id: hint.id,
          step: hint.step,
          title: hint.title ?? "",
          content: hint.content ?? "",
          selector: hint.selector ?? "",
          placement: hint.placement ?? "bottom",
          page_path: loadedTour.target_path || "/",
          spotlight: hint.spotlight ?? true,
          wait_for_selector: hint.wait_for_selector ?? false,
          media_url: hint.media_url ?? "",
        }))
      : [createEmptyHint(1)];

    createForm.reset({
      title: loadedTour.title,
      description: loadedTour.description || "",
      trigger_type: loadedTour.trigger_type || "on_load",
      audience: loadedTour.audience || {
        show_once: true,
        max_shows: 1,
        only_new: false,
      },
      hints: loadedHints,
    });
    setHasLoaded(true);
  }, [loadedTour, hasLoaded, createForm]);

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return;
      if (event.data?.type !== "SELECTOR_SELECTED") return;

      const selector = event.data.selector;
      if (!selector) return;

      const index = selectedHintIndex.current;
      if (index === null) return;

      const hints = createForm.getValues("hints");
      if (!hints[index]) return;

      createForm.setValue(`hints.${index}.selector`, selector, {
        shouldDirty: true,
        shouldValidate: true,
      });
      selectedHintIndex.current = null;
    };

    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [createForm]);

  const saveScenario = (
    data: InferType<typeof scenarioSchema>,
    scenarioStatus: "draft" | "published",
  ) => {
    const deletedHints =
      loadedTour?.hints?.filter(
        (loadedHint) => !data.hints.some((hint) => hint.id === loadedHint.id),
      ) ?? [];
    saveMutation.mutate({
      title: data.title,
      description: data.description ?? "",
      target_path: data.hints[0]?.page_path || "/",
      trigger_type: data.trigger_type as TriggerType,
      audience: data.audience as Audience,
      hints: data.hints as TourHint[],
      deletedHints,
      status: scenarioStatus,
    });
  };

  const saveDraft = async () => {
    if (!(await createForm.trigger("title"))) return;
    saveScenario(createForm.getValues(), "draft");
  };

  const onSubmit = (data: InferType<typeof scenarioSchema>) => {
    saveScenario(data, "published");
  };

  const addHint = () => {
    const hints = createForm.getValues("hints");
    createForm.setValue(
      "hints",
      [...hints, createEmptyHint(hints.length + 1)],
      { shouldDirty: true },
    );
    createForm.clearErrors("hints");
  };

  const removeHint = (index: number) => {
    const hints = createForm.getValues("hints");
    createForm.setValue(
      "hints",
      hints
        .filter((_, hintIndex) => hintIndex !== index)
        .map((hint, hintIndex) => ({
          ...hint,
          step: hintIndex + 1,
        })),
      { shouldDirty: true, shouldValidate: true },
    );
  };

  const selectElement = (index: number) => {
    selectedHintIndex.current = index;
    const page = createForm.getValues(`hints.${index}.page_path`) || "/";
    window.open(`${window.location.origin}${page}?builder=true`, "_blank");
  };

  return {
    createForm,
    saveDraft,
    onSubmit,
    addHint,
    removeHint,
    selectElement,
    isPending: saveMutation.isPending,
    errors: createForm.formState.errors,
  };
}