/// <reference types="vite-plugin-svgr/client" />

import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import Policy from "./modules/policy/components/policy-main/Policy";
import Marketplace from "./modules/marketplace/components/marketplace-main/Marketplace";
import Layout from "./Layout";
import PluginDetail from "./modules/plugin/components/plugin-detail/PluginDetail";

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        {/* Redirect / to /plugins */}
        <Route path="/" element={<Navigate to="/plugins" replace />} />
        <Route path="/plugins" element={<Layout />}>
          <Route index element={<Marketplace />} />
          <Route path=":pluginId" element={<PluginDetail />} />
          <Route path=":pluginId/policies" element={<Policy />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default App;
