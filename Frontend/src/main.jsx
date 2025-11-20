import react from 'react';
import {BrowserRouter, Navigate, Route, Routes} from 'react-router-dom';
import { createRoot } from 'react-dom/client'
import '../index.css'

import LoginPage from './AuthFlow/Login/Structure.jsx'
import RegisterPage from './AuthFlow/Register/Structure.jsx'
import VerifyOTP from "./AuthFlow/Register/VerifyOTP.jsx";
import ForgotPassword from "./AuthFlow/ForgotPassword.jsx";
import Sessions from "./AuthFlow/Sessions.jsx";

createRoot(document.getElementById('root')).render(
    <BrowserRouter>
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/verifyOTP" element={<VerifyOTP />} />
            <Route path="/forgotPassword" element={<ForgotPassword />} />
            <Route path="/sessions" element={<Sessions />} />

            <Route path="/" element={<Navigate to="/login" replace />} />
            <Route path="*" element={<div className="p-6">Page not found</div>} />
        </Routes>
    </BrowserRouter>
)
