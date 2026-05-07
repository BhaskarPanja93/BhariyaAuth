import {lazy, Suspense} from "react";
import {Navigate, Route, Routes} from 'react-router';
const LoginStructure = lazy(() => import('../Structures/SignIn'))
const RegisterStructure = lazy(() => import('../Structures/SignUp'))
const SessionsStructure = lazy(() => import('../Structures/Sessions'))
const PasswordResetStructure = lazy(() => import('../Structures/PasswordReset'))
const MfaStructure = lazy(() => import('../Structures/Mfa'))
const LogStructure = lazy(() => import('../Structures/Log'))

import '../index.css'

export default function Router() {
    return (
        <Suspense fallback={null}>
            <Routes>
                <Route path="/signin" element={<LoginStructure/>}/>
                <Route path="/signup" element={<RegisterStructure/>}/>
                <Route path="/sessions" element={<SessionsStructure/>}/>
                <Route path="/mfa" element={<MfaStructure/>}/>
                <Route path="/passwordreset" element={<PasswordResetStructure/>}/>
                <Route path="/log" element={<LogStructure/>}/>
                <Route path="*" element={<Navigate to="/sessions" replace/>}/>
            </Routes>
        </Suspense>
    )
}
