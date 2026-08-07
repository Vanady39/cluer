export interface Category {
  key: string;
  label: string;
}

export const categories : Category[] = [
  { key: "electronics", label: "Электроника" },
  { key: "things", label: "Личные вещи" },
  { key: "transport", label: "Транспорт" },
  { key: "real-estate", label: "Недвижимость" },
  { key: "services", label: "Услуги" },
  { key: "jobs", label: "Работа" },
  { key: "home", label: "Для дома и дачи" },
  { key: "hobbies", label: "Хобби и отдых" },
  { key: "animals", label: "Животные" },
  { key: "accessories", label: "Запчасти и аксессуары" }
];

export type CategoryKey = string;

export const getCategoryLabel = (key: string): string => {
  const category = categories.find(category => category.key === key);
  return category ? category.label : key;
};
