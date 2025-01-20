import DCAPluginPolicyForm from "@/modules/dca-plugin/components/DCAPluginPolicyForm";
import "./Modal.css";
import closeIcon from "@/assets/Close.svg";


function Modal({ isOpen, onClose }: { isOpen: boolean, onClose: () => void; }) {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <div className="modal-title">DCA Plugin Policy</div>
                <div className="modal-subtitle">Set up configuration settings for DCA Plugin Policy</div>
                <button className="modal-close" onClick={onClose}>
                    <img src={closeIcon} alt="" />
                </button>
                <DCAPluginPolicyForm />
            </div>
        </div>
    );
}

export default Modal;
