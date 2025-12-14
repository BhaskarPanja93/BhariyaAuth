import {lazy, Suspense} from "react";
import {BrowserRouter, Navigate, Route, Routes} from 'react-router-dom';
import {createRoot} from 'react-dom/client'
import '../index.css'

const DarkModeProvider = lazy(() => import('./Contexts/DarkMode'))
const NotificationProvider = lazy(() => import('./Contexts/Notification'))
const ConnectionProvider = lazy(() => import('./Contexts/Connection'))
const LoginStructure = lazy(() => import('./Structures/Login'))
const RegisterStructure = lazy(() => import('./Structures/Register'))
const SessionsStructure = lazy(() => import('./Structures/Sessions'))
const PasswordResetStructure = lazy(() => import('./Structures/PasswordReset'))
const MfaStructure = lazy(() => import('./Structures/Mfa'))

createRoot(document.getElementById('root')).render(<BrowserRouter basename="/auth">
    <Suspense fallback={null}>
        <DarkModeProvider>
            <NotificationProvider>
                <ConnectionProvider>
                    <Routes>
                        <Route path="/login" element={<LoginStructure/>}/>
                        <Route path="/register" element={<RegisterStructure/>}/>
                        <Route path="/sessions" element={<SessionsStructure/>}/>
                        <Route path="/mfa" element={<MfaStructure/>}/>
                        <Route path="/passwordreset" element={<PasswordResetStructure/>}/>
                        <Route path="*" element={<Navigate to="/sessions" replace/>}/>
                    </Routes>
                </ConnectionProvider>
            </NotificationProvider>
        </DarkModeProvider>
    </Suspense>
</BrowserRouter>);
