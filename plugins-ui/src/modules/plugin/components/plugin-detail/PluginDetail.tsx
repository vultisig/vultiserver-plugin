import Button from "@/modules/core/components/ui/button/Button";
import { useNavigate, useParams } from "react-router-dom";
import ChevronLeft from "@/assets/ChevronLeft.svg?react";
import logo from "../../../../assets/DCA-image.png"; // todo hardcoded until this image is stored in DB
import "./PluginDetail.css";
import { useEffect, useState } from "react";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import Toast from "@/modules/core/components/ui/toast/Toast";
import { Plugin } from "../../models/plugin";
import Reviews from "@/modules/review/components/reviews/Reviews";
import ApproveFeeButton from "./ApproveFeeButton";

const PluginDetail = () => {
  const navigate = useNavigate();
  const [plugin, setPlugin] = useState<Plugin | null>(null);
  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  const { pluginId } = useParams();

  useEffect(() => {
    const fetchPlugin = async (): Promise<void> => {
      if (!pluginId) return;

      try {
        const fetchedPlugin = await MarketplaceService.getPlugin(pluginId);
        setPlugin(fetchedPlugin);
      } catch (error: any) {
        console.error("Failed to get plugin:", error.message);
        setToast({
          message: "Failed to get plugin",
          error: error.error,
          type: "error",
        });
      }
    };

    fetchPlugin();
  }, []);

  return (
    <>
      <div className="only-section plugin-detail">
        <Button
          size="small"
          type="button"
          style={{ paddingLeft: "0px", paddingTop: "2rem" }}
          styleType="tertiary"
          onClick={() => navigate(`/plugins`)}
        >
          <ChevronLeft width="20px" height="20px" color="#F0F4FC" />
          Back to All Plugins
        </Button>

        {plugin && (
          <>
            <section className="plugin-header">
              <img src={logo} alt="" />
              <section className="plugin-details">
                <h2 className="plugin-title">{plugin.title}</h2>
                <p className="plugin-description">{plugin.description}</p>
                <section className="plugin-installaion">
                  {plugin && <ApproveFeeButton pluginUrl={plugin.server_endpoint} pluginType={plugin.type} />}
                  <Button
                    size="small"
                    type="button"
                    styleType="primary"
                    onClick={() => navigate(`/plugins/${plugin.id}/policies`)}
                  >
                    Install
                  </Button>
                  <aside>Plugin fee: 0.1% per trade</aside>
                </section>
              </section>
            </section>

            <Reviews plugin={plugin} />
          </>
        )}
      </div>

      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </>
  );
};

export default PluginDetail;
