export interface Listing {
  id: number;
  title: string;
  description: string;
  price: number;
  imageUrl: string;
  city?: string;
  category?: string;
}