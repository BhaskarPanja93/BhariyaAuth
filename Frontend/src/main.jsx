import React from "react";
import {BrowserRouter, Navigate, Route, Routes} from 'react-router-dom';
import {createRoot} from 'react-dom/client'
import '../index.css'

import LoginPage from './AuthFlow/Login/Structure.jsx'
import RegisterPage from './AuthFlow/Register/Structure.jsx'
import VerifyOTP from "./AuthFlow/Register/VerifyOTP.jsx";
import ResetPassword from "./AuthFlow/ResetPassword.jsx";
import Sessions from "./AuthFlow/Sessions.jsx";
import Test from "./AuthFlow/Test.jsx";
import {NotificationProvider} from "./Contexts/Notification.jsx";
import {ConnectionProvider} from "./Contexts/Connection.jsx";

createRoot(document.getElementById('root')).render(<React.StrictMode>
    <NotificationProvider>
        <ConnectionProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/test" element={<Test/>}/>
                    <Route path="/login" element={<LoginPage/>}/>
                    <Route path="/register" element={<RegisterPage/>}/>
                    <Route path="/verifyOTP" element={<VerifyOTP/>}/>
                    <Route path="/forgotPassword" element={<ResetPassword/>}/>
                    <Route path="/sessions" element={<Sessions/>}/>

                    <Route path="/" element={<Navigate to="/login" replace/>}/>
                    <Route path="*" element={<div className="p-6">Page not found</div>}/>
                </Routes>
            </BrowserRouter>
        </ConnectionProvider>
    </NotificationProvider>
</React.StrictMode>)
