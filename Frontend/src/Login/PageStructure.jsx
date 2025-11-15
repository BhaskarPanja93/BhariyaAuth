import react from "react";
import {createRoot} from "react-dom/client";
import { useState } from 'react'

import EmailInput from './Components/EmailInput.jsx'
import MethodToggle from './Components/MethodToggle'
import OTPInput from './Components/OTPInput'
import PasswordInput from './Components/PasswordInput'
import RememberCheckbox from './Components/RememberCheckbox'
import SubmitButton from './Components/SubmitButton'
import SocialButtons from './Components/SocialButtons'
import Divider from './Components/Divider'

export default function LoginPage(){
    const [useOtp, setUseOtp] = useState(false)
    const [pwVisible, setPwVisible] = useState(false)
    const [form, setForm] = useState({ email:'', password:'', otp:'', remember:false })

    function handleChange(e){
        const { name, value, type, checked } = e.target
        setForm(f => ({ ...f, [name]: type === 'checkbox' ? checked : value }))
    }


    return (
        <div className="w-full max-w-sm">
            <div className="rounded-2xl p-8 shadow-2xl" style={{ background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))', border: '1px solid rgba(255,255,255,0.02)' }}>
                <div className="flex flex-col items-center gap-4 mb-4">
                    <h2 className="text-xl font-semibold text-white">Sign In</h2>
                    <p className="text-sm text-gray-400">Access your account</p>
                </div>

                <form className="space-y-4">
                    <EmailInput value={form.email} onChange={handleChange} />

                    <div className="flex items-center justify-between text-xs text-gray-400">
                        <MethodToggle useOtp={useOtp} setUseOtp={setUseOtp} />
                        <div className="flex items-center gap-3">
                            <button type="button" className={`${useOtp ? '' : 'hidden'} text-xs text-indigo-400 hover:underline`}>Resend OTP</button>
                            <button type="button" className={`${useOtp ? 'hidden' : ''} text-xs text-indigo-400 hover:underline`}>Forgot Password</button>
                        </div>
                    </div>

                    {useOtp ? (
                        <OTPInput value={form.otp} onChange={handleChange} />
                    ) : (
                        <PasswordInput name="password" value={form.password} onChange={handleChange} visible={pwVisible} onToggle={() => setPwVisible(v => !v)} />
                    )}

                    <RememberCheckbox checked={form.remember} onChange={handleChange} />

                    <SubmitButton />

                    <Divider />

                    <SocialButtons />

                    <p className="text-center text-sm text-gray-500 mt-4">New here? <a href="/register" className="text-indigo-400 hover:underline">Create an account</a></p>
                </form>
            </div>
        </div>
    )
}

