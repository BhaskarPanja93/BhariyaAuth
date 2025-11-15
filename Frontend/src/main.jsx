import react from 'react';
import {BrowserRouter, Navigate, Route, Routes} from 'react-router-dom';
import { createRoot } from 'react-dom/client'
import '../index.css'

import LoginPage from './Login/PageStructure.jsx'

createRoot(document.getElementById('root')).render(
    <BrowserRouter>
        <Routes>
            <Route path="/" element={<Navigate to="/login" replace />} />
            <Route path="/login" element={<LoginPage />} />

            <Route path="*" element={<div className="p-6">Page not found</div>} />
        </Routes>
    </BrowserRouter>
)
