import {useRef, useState} from 'react'
import {Link, useNavigate} from "react-router-dom";
import EmailInput from '../Elements/EmailInput.jsx'
import PasswordInput from '../Elements/PasswordInput.jsx'
import RememberCheckbox from '../Elements/RememberCheckbox.jsx'
import SubmitButton from '../Elements/SubmitButton.jsx'
import SSOButtons from '../Elements/SSOButtons.jsx'
import Divider from '../Elements/Divider.jsx'
import NameInput from '../Elements/NameInput.jsx'
import OTPInput from "../Elements/OTPInput.jsx";
import {BackendURL} from "../Values/Constants.js";
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import {FetchNotificationManager} from "../Contexts/Notification.jsx";
import {EmailIsValid, NameIsValid, PasswordIsStrong} from "../Utils/Strings.js";

export default function RegisterPage() {
    const navigate = useNavigate();
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [passwordConfirmation, setPasswordConfirmation] = useState("")
    const [otp, setOTP] = useState("");
    const [name, setName] = useState("")

    const currentToken = useRef("")

    const Step1 = () => {
        if (!NameIsValid(name)) return SendNotification("Name is invalid")
        if (!EmailIsValid(email)) return SendNotification("Email is invalid")
        if (!PasswordIsStrong(password)) return SendNotification("Password is too weak")
        if (password !== passwordConfirmation) return SendNotification("Passwords don't match")

        setUiDisabled(true);

        const form = new FormData();
        form.append("mail_address", email);
        form.append("name", name);
        form.append("password", password);
        form.append("remember_me", remember ? "yes" : "no");
        privateAPI.post(BackendURL + "/register/step1", form)
            .then((data) => {
                if (data["success"]) {
                    currentToken.current = data["reply"];
                    setCurrentStep(2)
                }
            })
            .catch((error)=>{console.log("Register Step1 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
        if (!currentToken.current) return SendNotification("Step 1 incomplete. Please enter email again");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", otp);
        privateAPI.post(BackendURL + "/register/step2", form, {})
            .then((data) => {
                if (data["success"]) {
                    navigate("/sessions")
                }
            })
            .catch((error)=>{console.log("Register Step2 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    return (<div className="min-h-screen flex items-center justify-center">
        <div className="w-full max-w-sm">
            <div className="rounded-2xl p-8 shadow-2xl" style={{
                background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))', border: '1px solid rgba(255,255,255,0.02)'
            }}>
                <div className="flex flex-col items-center gap-4 mb-4">
                    <h2 className="text-xl font-semibold text-white">Sign Up</h2>
                    <div className="text-sm text-gray-400">
                        {currentStep === 1 ? ("Access your account") : (<div className="flex items-center gap-2">
                            <span>{email}</span>
                            <span
                                onClick={() => setCurrentStep(1)}
                                className="text-indigo-400 cursor-pointer"
                            >Not you?
                                </span>
                        </div>)}
                    </div>
                </div>
                <div className="space-y-4">

                    {currentStep === 1 && <>
                        <NameInput value={name} onValueChange={setName} disabled={currentStep !== 1 || uiDisabled}/>
                        <EmailInput value={email} onValueChange={setEmail} disabled={currentStep !== 1 || uiDisabled}/>
                        <PasswordInput disabled={currentStep !== 1 || uiDisabled} value={password}
                                       onValueChange={setPassword}
                                       confirm={passwordConfirmation}
                                       onConfirmChange={setPasswordConfirmation}
                                       needsConfirm={true}/>
                        <RememberCheckbox checked={remember} onCheckedChange={setRemember} disabled={currentStep !== 1 || uiDisabled}/>
                    </>}
                    {currentStep === 2 && <>
                        <button type="button" onClick={() => Step1(true)} className="flex-end text-xs text-indigo-400 hover:underline">
                            Resend OTP
                        </button>
                        <OTPInput value={otp} onValueChange={setOTP} disabled={uiDisabled}/></>}
                    <SubmitButton text={currentStep === 1 ? "Verify Email" : "VERIFY OTP"}
                                  onClick={currentStep === 1 ? Step1 : Step2} disabled={uiDisabled}/>
                    <Divider/>
                    <SSOButtons disabled={uiDisabled}/>
                    <p className="text-center text-sm text-gray-500 mt-4">
                        Already have an account?&nbsp;
                        <Link to="/login" className="text-indigo-400 hover:underline">
                            SignIn
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    </div>)
}
