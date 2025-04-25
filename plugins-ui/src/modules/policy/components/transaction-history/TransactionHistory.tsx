import { useEffect, useState } from "react";
import { usePolicies } from "../../context/PolicyProvider";
import { PolicyTransactionHistory } from "../../models/policy";
import "./TransactionHistory.css";
import Toast from "@/modules/core/components/ui/toast/Toast";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";

const ITEMS_PER_PAGE = 25;

type TransactionHistoryProps = {
  policyId: string;
};

const formatDate = (dateString: string): { date: string; time: string } => {
  const dateObj = new Date(dateString);
  return {
    date: dateObj.toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }), // "26 Feb 2025"
    time: dateObj.toLocaleTimeString("en-GB", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    }), // "14:25:35"
  };
};

const TransactionHistory = ({ policyId }: TransactionHistoryProps) => {
  const { getPolicyHistory } = usePolicies();
  const [historyData, setHistoryData] = useState<PolicyTransactionHistory[]>(
    []
  );

  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "warning" | "error";
  } | null>(null);

  const [currentPage, setCurrentPage] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  useEffect(() => {
    const fetchPolicyHistory = async (): Promise<void> => {
      try {
        const fetchedHistory = await getPolicyHistory(
          policyId,
          currentPage > 1 ? (currentPage - 1) * ITEMS_PER_PAGE : 0,
          ITEMS_PER_PAGE
        );

        if (!fetchedHistory) return;

        setHistoryData(fetchedHistory.history);
        setTotalPages(Math.ceil(fetchedHistory.total_count / ITEMS_PER_PAGE));

        if (
          fetchedHistory.total_count / ITEMS_PER_PAGE > 1 &&
          currentPage === 0
        ) {
          setCurrentPage(1);
        }
      } catch (error: any) {
        console.error("Failed to get policy history:", error.message);
        setToast({
          message: error.message || "Failed to get policy history",
          error: error.error,
          type: "error",
        });
      }
    };

    fetchPolicyHistory();
  }, [currentPage]);

  const onCurrentPageChange = (page: number): void => {
    setCurrentPage(page);
  };

  return (
    <div className="history-panel">
      <h2>Transaction History</h2>
      <ul>
        {historyData &&
          historyData.map((item) => {
            const { date, time } = formatDate(item.updated_at);
            return (
              <li key={item.id} className="history-item">
                <span className="history-status">{item.status}</span>
                <span className="history-date">{date}</span>
                <span className="history-time">{time}</span>
              </li>
            );
          })}
        {!historyData && (
          <li key={1} className="history-item">
            Nothing to see here yet.
          </li>
        )}
      </ul>

      {totalPages > 1 && (
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={onCurrentPageChange}
        />
      )}
      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  );
};

export default TransactionHistory;
