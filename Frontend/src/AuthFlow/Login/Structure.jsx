import {useState} from 'react'
import {Link} from "react-router-dom";

import EmailInput from '../Common/EmailInput'
import Step2Toggle from './Step2Toggle'
import OTPInput from '../Common/OTPInput'
import PasswordInput from '../Common/PasswordInput'
import RememberCheckbox from '../Common/RememberCheckbox'
import SubmitButton from '../Common/SubmitButton'
import SSOButtons from '../Common/SSOButtons.jsx'
import {FetchConnectionManager} from "../../Contexts/Connection.jsx";

export default function LoginPage() {
    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [useOtp, setUseOtp] = useState(false)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState("")

    const Step1 = async () => {
        const {publicAPI, privateAPI} = FetchConnectionManager()
    }

    const Step2 = async () => {

    }

    return (<div className="min-h-screen flex items-center justify-center">
        <div className="w-full max-w-sm">
            <div className="rounded-2xl p-8 shadow-2xl" style={{
                background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                border: '1px solid rgba(255,255,255,0.02)'
            }}>
                <div className="flex flex-col items-center gap-4 mb-4">
                    <h2 className="text-xl font-semibold text-white">Sign In</h2>
                    <p className="text-sm text-gray-400">
                        {currentStep === 1 ? ("Access your account") : (email)}</p>
                </div>
                <div className="space-y-4">
                    <EmailInput value={email} onValueChange={setEmail} disabled={uiDisabled || currentStep !== 1}/>
                    {currentStep === 2 && (
                        <>
                            <div className="flex items-center justify-between text-xs text-gray-400">
                                <Step2Toggle usingOTP={useOtp} setUsingOtp={setUseOtp} disabled={uiDisabled}/>
                                <div className="flex items-center gap-3">
                                    {useOtp ? (
                                        <button type="button" className="text-xs text-indigo-400 hover:underline">Resend
                                            OTP</button>) : (
                                        <Link to={"/passwordreset"} className="text-xs text-indigo-400 hover:underline">Forgot
                                            Password?</Link>)}
                                </div>
                            </div>
                            useOtp ? (
                            <OTPInput value={verification} onValueChange={setVerification} disabled={uiDisabled}/>
                            ) : (
                            <PasswordInput value={verification} onValueChange={setVerification} disabled={uiDisabled}/>
                            )
                        </>)}

                    <RememberCheckbox checked={remember} onCheckedChange={setRemember}
                                      disabled={uiDisabled || currentStep !== 1}/>
                    {currentStep === 1 ? (
                        <SubmitButton text={"CONTINUE"} disabled={uiDisabled || currentStep !== 1}/>) : (
                        <SubmitButton text={"SIGN IN"} disabled={uiDisabled || currentStep !== 2}/>)}
                    <Divider/>
                    <SSOButtons disabled={uiDisabled}/>
                    <p className="text-center text-sm text-gray-500 mt-4">
                        New here? <Link to="/register" className="text-indigo-400 hover:underline">Create an
                        account</Link>
                    </p>
                </div>
            </div>
        </div>
    </div>)
}
