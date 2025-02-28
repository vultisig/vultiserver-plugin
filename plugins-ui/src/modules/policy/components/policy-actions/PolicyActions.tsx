import Button from "@/modules/core/components/ui/button/Button";
import TrashIcon from "@/assets/Trash.svg?react";
import PenIcon from "@/assets/Pen.svg?react";
import BookIcon from "@/assets/Book.svg?react";
import PauseIcon from "@/assets/Pause.svg?react";
import { usePolicies } from "@/modules/policy/context/PolicyProvider";
import Modal from "@/modules/core/components/ui/modal/Modal";
import PolicyForm from "@/modules/policy/components/policy-form/PolicyForm";
import { useState } from "react";
import TransactionHistory from "../transaction-history/TransactionHistory";

type PolicyActionsProps = {
  policyId: string;
};

const PolicyActions = ({ policyId }: PolicyActionsProps) => {
  const [editModalId, setEditModalId] = useState("");
  const [transactionHistoryModalId, setTransactionHistoryModalId] =
    useState("");
  const { policyMap, updatePolicy, removePolicy } = usePolicies();

  const handleUpdate = () => {
    const policy = policyMap.get(policyId);
    if (policy) {
      policy.active = !policy.active;
      updatePolicy(policy);
    }
  };

  return (
    <>
      <div style={{ display: "flex" }}>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={handleUpdate}
        >
          <PauseIcon width="20px" height="20px" color="#F0F4FC" />
        </Button>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={() => setEditModalId(policyId)}
        >
          <PenIcon width="20px" height="20px" color="#F0F4FC" />
        </Button>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={() => setTransactionHistoryModalId(policyId)}
        >
          <BookIcon width="20px" height="20px" color="#F0F4FC" />
        </Button>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={() => removePolicy(policyId)}
        >
          <TrashIcon width="20px" height="20px" color="#FF5C5C" />
        </Button>
      </div>

      <Modal
        isOpen={editModalId !== ""}
        onClose={() => setEditModalId("")}
        variant="panel"
      >
        <PolicyForm
          data={policyMap.get(editModalId)}
          onSubmitCallback={() => setEditModalId("")}
        />
      </Modal>
      <Modal
        isOpen={transactionHistoryModalId !== ""}
        onClose={() => setTransactionHistoryModalId("")}
        variant="panel"
      >
        <TransactionHistory policyId={transactionHistoryModalId} />
      </Modal>
    </>
  );
};

export default PolicyActions;
