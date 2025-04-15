import { Plugin } from "@/modules/plugin/models/plugin";

export type ViewFilter = "grid" | "list";

export type PluginMap = {
  plugins: Plugin[];
  total_count: number;
};
