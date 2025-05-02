import { Outlet } from "react-router-dom";
import Wallet from "./modules/shared/wallet/Wallet";
import "./Layout.css";
import Toast from "./modules/core/components/ui/toast/Toast";

const Layout = () => {
  return (
    <div className="container">
      <Toast />
      <div className="navbar">
        <>Vultisig</>
        <Wallet />
      </div>
      <div className="content">
        <Outlet />
      </div>
    </div>
  );
};

export default Layout;
