import PolicyForm from "../policy-form/PolicyForm";
import { PolicyProvider } from "../../context/PolicyProvider";
import PolicyTable from "../policy-table/PolicyTable";
import { useEffect, useState } from "react";

const Policy = () => {
  const [authToken, setAuthToken] = useState(
    localStorage.getItem("authToken") || ""
  );

  useEffect(() => {
    const handleStorageChange = () => {
      setAuthToken(localStorage.getItem("authToken") as string);
    };

    // Listen for storage changes
    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [authToken]);

  if (!authToken) return <p>Please connect to wallet!</p>;

  return (
    <PolicyProvider>
      <div className="left-section">
        <PolicyForm />
      </div>
      <div className="right-section">
        <PolicyTable />
      </div>
    </PolicyProvider>
  );
};

export default Policy;
