export interface City {
  value: string;
  label: string;
}

export const cities: City[] = [
  { value: "moscow", label: "Москва" },
  { value: "saint-petersburg", label: "Санкт-Петербург" },
  { value: "kazan", label: "Казань" },
  { value: "ekaterinburg", label: "Екатеринбург" },
  { value: "novosibirsk", label: "Новосибирск" },
  { value: "nnovgorod", label: "Нижний Новгород" },
];

export type CityValue = string;
