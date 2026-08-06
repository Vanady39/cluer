import type { Tour } from "../../types/sdk";

const KEY = "onboarding_tours";

export function getTours(): Tour[] {
  const data = localStorage.getItem(KEY);
  return data ? JSON.parse(data) : [];
}

export function saveTour(tour: Tour) {
  const tours = getTours();
  const existingIndex = tours.findIndex((t) => t.id === tour.id);

  if (existingIndex !== -1) {
    tours[existingIndex] = tour;
  } else {
    tours.push(tour);
  }

  localStorage.setItem(KEY, JSON.stringify(tours));
}

export function deleteTour(id: string) {
  const tours = getTours();
  const updated = tours.filter((t) => t.id !== id);
  localStorage.setItem(KEY, JSON.stringify(updated));
  return updated;
}

export function getPublishedTour(path?: string) {
  const tours = getTours();
  
  return tours.find((tour) => {
    if (tour.status !== "published") return false;
    if (path) {
      return tour.target_path === path;
    }
    return true;
  });
}