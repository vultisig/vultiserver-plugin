import { useCallback, useEffect, useState } from "react";
import "./Toast.css";
import Button from "../button/Button";
import CloseIcon from "@/assets/Close.svg?react";
import { subscribe, unsubscribe } from "@/utils/eventBus";

type Toast = {
  message: string;
  type: "success" | "warning" | "error";
  title?: string;
  close?: boolean;
  duration?: number; // Auto-close time (ms)
  onClose?: () => void;
};

const Toast = () => {
  const [snackbars, setSnackbars] = useState<Toast[]>([]);

  const handleOnToast = useCallback((evt: CustomEventInit<Toast>) => {
    const toast = evt.detail;
    if (!toast) return;
    setSnackbars((prev) => [...prev, toast]);
    if (!toast.duration) return;
    setTimeout(() => {
      setSnackbars((prev) => prev.filter((t) => t !== toast));
    }, toast.duration);
  }, []);

  const handleCloseToast = (toast: Toast) => {
    setSnackbars((prev) => prev.filter((t) => t !== toast));
  };

  useEffect(() => {
    subscribe("onToast", handleOnToast);
    return () => {
      unsubscribe("onToast", handleOnToast);
    };
  }, [handleOnToast]);

  return (
    <div className="toasts-wrapper">
      <div className="toasts-container">
        {snackbars.map((toast, index) => {
          return (
            <div
              key={`${toast.message}-${index}`}
              className={`toast toast-${toast.type}`}
              data-testid="toast-item"
            >
              <div className="toast-content">
                {toast.title ? (
                  <div className="toast-title">{toast.title}</div>
                ) : null}
                <div className="toast-message">{toast.message}</div>
              </div>
              <Button
                ariaLabel="Close message"
                style={{ paddingRight: 0 }}
                size="small"
                type="button"
                styleType="tertiary"
                onClick={() => handleCloseToast(toast)}
                data-testid="toast-item-close-btn"
              >
                <CloseIcon />
              </Button>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default Toast;
