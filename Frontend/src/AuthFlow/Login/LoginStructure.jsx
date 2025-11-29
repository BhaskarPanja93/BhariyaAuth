import {useRef, useState} from 'react'
import {Link, useNavigate} from "react-router-dom";
import {BackendURL} from '../../Values/Constants.js'
import EmailInput from '../Common/EmailInput'
import Step2Toggle from './Step2Toggle'
import OTPInput from '../Common/OTPInput'
import PasswordInput from '../Common/PasswordInput'
import RememberCheckbox from '../Common/RememberCheckbox'
import SubmitButton from '../Common/SubmitButton'
import SSOButtons from '../Common/SSOButtons.jsx'
import {FetchConnectionManager} from "../../Contexts/Connection.jsx";
import {EmailIsValid, OTPIsValid, PasswordIsStrong} from "../../Utils/Strings.js";
import Divider from "../Common/Divider.jsx";
import {FetchNotificationManager} from "../../Contexts/Notification.jsx";

export default function LoginStructure() {
    const navigate = useNavigate()
    const {SendNotification} = FetchNotificationManager();

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [useOtp, setUseOtp] = useState(false)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState()

    const {privateAPI} = FetchConnectionManager()
    const tokens = useRef({})

    const Step1 = async (tryOTP) => {
        if (!EmailIsValid(email)) return SendNotification("Email is invalid");
        if (!tokens.current[email]) tokens.current[email] = {}

        if (tokens.current[email][tryOTP]) {
            setUseOtp(tryOTP)
            return setCurrentStep(2)
        }

        setUiDisabled(true);
        const form = new FormData();
        form.append("mail_address", email);
        form.append("remember_me", remember ? "yes" : "no");
        privateAPI.post(BackendURL + `/login/step1/${tryOTP ? "otp" : "password"}`, form)
            .then((data) => {
                if (data["success"]) {
                    tokens.current[email][tryOTP] = data["reply"]
                    setUseOtp(tryOTP)
                    setCurrentStep(2)
                }
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = async () => {
        if (!tokens.current[email] || !tokens.current[email][useOtp]) return SendNotification("Step 1 incomplete. Please enter email again");
        if (!useOtp) {
            if (!PasswordIsStrong(verification)) return SendNotification("Incorrect Password");
        } else {
            if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");
        }

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", tokens.current[email][useOtp]);
        form.append("verification", verification);
        privateAPI.post(BackendURL + "/login/step2", form)
            .then((data) => {
                if (data["success"]) {
                    navigate("/sessions");
                }
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    return (<div className="min-h-screen flex items-center justify-center">
        <div className="w-full max-w-sm">
            <div className="rounded-2xl p-8 shadow-2xl" style={{
                background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                border: '1px solid rgba(255,255,255,0.02)'
            }}>
                <div className="flex flex-col items-center gap-4 mb-4">
                    <h2 className="text-xl font-semibold text-white">Sign In</h2>
                    <div className="text-sm text-gray-400">
                        {currentStep === 1 ? ("Access your account") : (<div className="flex items-center gap-2">
                                <span>{email}</span>
                                <span
                                    onClick={() => {
                                        setUseOtp(false);
                                        setCurrentStep(1);
                                    }}
                                    className="text-indigo-400 cursor-pointer"
                                >Not you?
                                </span>
                        </div>)}
                    </div>

                </div>
                <div className="space-y-4">
                    {currentStep === 1 && <>
                        <EmailInput value={email} onValueChange={setEmail} disabled={uiDisabled || currentStep !== 1}/>
                        <RememberCheckbox checked={remember} onCheckedChange={setRemember}
                                          disabled={uiDisabled || currentStep !== 1}/>
                    </>}
                    <div className="text-xs text-gray-400">
                        <div className="flex items-center justify-between">
                            <Step2Toggle usingOTP={useOtp} toggleUsingOTP={() => Step1(!useOtp)} disabled={uiDisabled}/>
                            <div className="flex items-center gap-3">
                                {!useOtp ?
                                    <Link to="/passwordreset" className="text-xs text-indigo-400 hover:underline">
                                        Forgot Password?
                                    </Link> : <button type="button" onClick={() => Step1(true)}
                                                      className="text-xs text-indigo-400 hover:underline">
                                        Resend OTP
                                    </button>}
                            </div>
                        </div>
                        {(currentStep === 2 && tokens.current[email] && tokens.current[email][useOtp]) &&
                            <div className="mt-3">
                                {useOtp ? <OTPInput value={verification} onValueChange={setVerification}
                                                    disabled={uiDisabled}/> :
                                    <PasswordInput value={verification} onValueChange={setVerification}
                                                   disabled={uiDisabled}/>}
                            </div>}
                    </div>

                    {currentStep === 1 ? <SubmitButton text={"Continue with Email"} onClick={() => Step1(false)}
                                                       disabled={uiDisabled || currentStep !== 1}/> :
                        <SubmitButton text={"Sign In"} onClick={Step2} disabled={uiDisabled}/>}
                    <Divider/>
                    <SSOButtons disabled={uiDisabled}/>
                    <p className="text-center text-sm text-gray-500 mt-4">
                        New here?&nbsp;
                        <Link to="/register" className="text-indigo-400 hover:underline">
                            Create an account
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    </div>)
}
