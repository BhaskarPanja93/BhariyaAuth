import React from "react";
import {BrowserRouter, Navigate, Route, Routes} from 'react-router-dom';
import {createRoot} from 'react-dom/client'
import '../index.css'

import Login from './Structures/Login.jsx'
import RegisterPage from './Structures/Register.jsx'
import Mfa from "./Structures/Mfa.jsx";
import PasswordReset from "./Structures/PasswordReset.jsx";
import Sessions from "./Structures/Sessions.jsx";
import {NotificationProvider} from "./Contexts/Notification.jsx";
import {ConnectionProvider} from "./Contexts/Connection.jsx";

createRoot(document.getElementById('root')).render(
    <NotificationProvider>
        <ConnectionProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/login" element={<Login/>}/>
                    <Route path="/register" element={<RegisterPage/>}/>
                    <Route path="/mfa" element={<Mfa/>}/>
                    <Route path="/passwordreset" element={<PasswordReset/>}/>
                    <Route path="/sessions" element={<Sessions/>}/>

                    <Route path="/" element={<Navigate to="/login" replace/>}/>
                    <Route path="*" element={<div className="p-6">Page not found</div>}/>
                </Routes>
            </BrowserRouter>
        </ConnectionProvider>
    </NotificationProvider>
)
