import { api } from './api';
import type { Listing } from '../types/listings';

export const listingsAPI = {
  getAll: () =>
    api.get<{ data: Listing[] }>('/listings').then((res) => res.data.data),

  getById: (id: number) =>
    api.get<{ data: Listing }>(`/listings/${id}`).then((res) => res.data.data),

  create: (data: Omit<Listing, 'id'>) =>
    api.post<{ data: Listing }>('/listings', data).then((res) => res.data.data),

  update: (id: number, data: Partial<Omit<Listing, 'id'>>) =>
    api.put<{ data: Listing }>(`/listings/${id}`, data).then((res) => res.data.data),

  delete: (id: number) =>
    api.delete<{ data: null }>(`/listings/${id}`).then((res) => res.data.data),
};