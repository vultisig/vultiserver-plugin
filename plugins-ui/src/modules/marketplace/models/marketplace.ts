import { Plugin, PluginRatings } from "@/modules/plugin/models/plugin";

export type ViewFilter = "grid" | "list";

export type PluginMap = {
  plugins: Plugin[];
  total_count: number;
};

export type Review = {
  id: string;
  address: string;
  rating: number;
  comment: string;
  created_at: string;
  plugin_id: string;
  ratings: PluginRatings[];
};

export type CreateReview = {
  address: string;
  rating: number;
  comment: string;
};

export type ReviewMap = {
  reviews: Review[];
  total_count: number;
};
