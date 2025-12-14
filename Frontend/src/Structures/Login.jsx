import {useEffect, useRef, useState} from 'react'
import {Link, useNavigate} from "react-router-dom";
import {BackendURL} from '../Values/Constants.js'
import EmailInput from '../Elements/EmailInput.jsx'
import Step2Toggle from '../Elements/Step2Toggle.jsx'
import OTPInput from '../Elements/OTPInput.jsx'
import PasswordInput from '../Elements/PasswordInput.jsx'
import RememberCheckbox from '../Elements/RememberCheckbox.jsx'
import SubmitButton from '../Elements/SubmitButton.jsx'
import SSOButtons from '../Elements/SSOButtons.jsx'
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import {EmailIsValid, OTPIsValid, PasswordIsStrong} from "../Utils/Strings.js";
import Divider from "../Elements/Divider.jsx";
import {FetchNotificationManager} from "../Contexts/Notification.jsx";
import {Countdown} from "../Utils/Countdown.js";
import OTPResendButton from "../Elements/OTPResendButton.jsx";

export default function LoginPage() {
    const navigate = useNavigate()
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [useOtp, setUseOtp] = useState(false)
    const OTPResendTimerID = useRef(0)
    const [OTPDelay, setOTPDelay] = useState(0)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState()

    const tokens = useRef({})

    const Step1 = (tryOTP, resendOTP) => {
        if (!EmailIsValid(email)) return SendNotification("Email is invalid");
        if (!tokens.current[email]) tokens.current[email] = {}

        if (tokens.current[email][tryOTP] && !resendOTP) {
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
                    SendNotification(`Please enter the ${tryOTP?"OTP":"Password"}`)
                    tokens.current[email][tryOTP] = data["reply"]
                    setUseOtp(tryOTP)
                    setCurrentStep(2)
                } else if (data["reply"]){
                    Countdown(data["reply"], 0.1, OTPResendTimerID, setOTPDelay).then()
                }
            })
            .catch((error)=>{console.log("Login Step1 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
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
        privateAPI.post(BackendURL + "/login/step2", form, {forAccessFetch: true})
            .then((data) => {
                if (data["success"]) {
                    SendNotification("Logged In Successfully")
                    navigate("/");
                }
            })
            .catch((error)=>{console.log("Login Step2 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    useEffect(() => {
        document.title = "Login - Bhariya";
    }, []);

    return (<div className="min-h-screen flex items-center justify-center">
        <div className="w-full max-w-sm">
            <div className="rounded-2xl p-8 shadow-2xl"
                style={{
                    background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                    border: '1px solid rgba(255,255,255,0.02)'
                    }}>
                <div className="flex flex-col items-center gap-4 mb-4">
                    <h2 className="text-xl font-semibold text-white">
                        Sign In
                    </h2>
                    <div className="text-sm text-gray-400">
                        {currentStep === 1 ?
                            "Access your account"
                            :
                            <div className="flex items-center gap-2">
                                <span>{email}</span>
                                <span className="text-indigo-400 cursor-pointer"
                                    onClick={() => {setUseOtp(false);setCurrentStep(1);}}>
                                    Not you?
                                </span>
                            </div>
                        }
                    </div>
                </div>
                <div className="space-y-4">
                    <EmailInput
                        value={email}
                        onValueChange={setEmail}
                        disabled={uiDisabled || currentStep !== 1}
                        hidden={currentStep === 2}/>
                    {currentStep === 1 &&
                        <RememberCheckbox
                            checked={remember}
                            onCheckedChange={setRemember}
                            disabled={uiDisabled || currentStep !== 1}/>
                    }
                    <div className="text-xs text-gray-400">
                        <div className="flex items-center justify-between">
                            <Step2Toggle
                                usingOTP={useOtp}
                                toggleUsingOTP={() => Step1(!useOtp)}
                                disabled={uiDisabled}/>
                            <div className="flex items-center gap-3">
                                {!useOtp ?
                                    <Link className="text-xs text-indigo-400 hover:underline"
                                        to="/passwordreset">
                                        Forgot Password?
                                    </Link>
                                    :
                                    <OTPResendButton delay={OTPDelay} onClick={()=>Step1(true, true)} disabled={uiDisabled || currentStep !== 2} />
                                }
                            </div>
                        </div>
                        {(currentStep === 2 && tokens.current[email] && tokens.current[email][useOtp]) &&
                            <div className="mt-3">
                                {useOtp ?
                                    <OTPInput
                                        value={verification}
                                        onValueChange={setVerification}
                                        disabled={uiDisabled || currentStep !== 2}/>
                                    :
                                    <PasswordInput
                                        value={verification}
                                        onValueChange={setVerification}
                                        disabled={uiDisabled|| currentStep !== 2}/>
                                }
                            </div>}
                    </div>
                    <SubmitButton
                        text={currentStep === 1 ? "Continue with Email" : "Sign In"}
                        onClick={currentStep === 1 ? () => Step1(false) : Step2}
                        disabled={uiDisabled}/>
                    <Divider/>
                    <SSOButtons disabled={uiDisabled}/>
                    <p className="text-center text-sm text-gray-500 mt-4">
                        New here?&nbsp;
                        <Link className="text-indigo-400 hover:underline"
                            to="/register">
                            Create an account
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    </div>)
}
