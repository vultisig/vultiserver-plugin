import { useState } from 'react';
import './App.css'
import Modal from './modules/core/components/ui/modal/Modal';

const App = () => {
    const [isModalOpen, setModalOpen] = useState(false);

    return (
        <>
            {!isModalOpen && <button onClick={() => setModalOpen(true)}>View DCA plugin policy</button>}
            <Modal isOpen={isModalOpen} onClose={() => setModalOpen(false)} />
        </>
    );
};

export default App;
