import {useEffect, useRef, useState} from 'react'
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
import {EmailIsValid, NameIsValid, OTPIsValid, PasswordIsStrong} from "../Utils/Strings.js";
import {Countdown} from "../Utils/Countdown.js";
import OTPResendButton from "../Elements/OTPResendButton.jsx";

export default function RegisterPage() {
    const navigate = useNavigate();
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const OTPResendTimerID = useRef(0)
    const [OTPDelay, setOTPDelay] = useState(0)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [passwordConfirmation, setPasswordConfirmation] = useState("")
    const [verification, setVerification] = useState("");
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
                    SendNotification("Please enter the OTP sent to your mail")
                    currentToken.current = data["reply"];
                    setCurrentStep(2)
                } else if (data["reply"]) {
                    Countdown(data["reply"], OTPResendTimerID, setOTPDelay).then()
                }
            })
            .catch((error)=>{console.log("Register Step1 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
        if (!currentToken.current) return SendNotification("Step 1 incomplete. Please resend OTP");
        if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", verification);
        privateAPI.post(BackendURL + "/register/step2", form, {forAccessFetch: true})
            .then((data) => {
                if (data["success"]) {
                    SendNotification("Registered and logged in Successfully")
                    navigate("/sessions")
                }
            })
            .catch((error)=>{console.log("Register Step2 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    useEffect(() => {
        document.title = "Register - Bhariya";
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
                        Sign Up
                    </h2>
                    <div className="text-sm text-gray-400">
                        {currentStep === 1 ?
                            "Create an account"
                            :
                            <div className="flex items-center gap-2">
                            <span>{email}</span>
                            <span className="text-indigo-400 cursor-pointer"
                                onClick={() => setCurrentStep(1)}>
                                Not you?
                            </span>
                        </div>}
                    </div>
                </div>
                <div className="space-y-4">
                    {currentStep === 1 && <>
                        <NameInput
                            value={name}
                            onValueChange={setName}
                            disabled={uiDisabled || currentStep !== 1}/>
                        <EmailInput
                            value={email}
                            onValueChange={setEmail}
                            disabled={uiDisabled || currentStep !== 1}/>
                        <PasswordInput
                            disabled={uiDisabled || currentStep !== 1}
                            value={password}
                            onValueChange={setPassword}
                            confirm={passwordConfirmation}
                            onConfirmChange={setPasswordConfirmation}
                            needsConfirm={true}/>
                        <RememberCheckbox
                            checked={remember}
                            onCheckedChange={setRemember}
                            disabled={uiDisabled || currentStep !== 1}/>
                    </>}
                    {currentStep === 2 &&
                        <>
                            <OTPResendButton delay={OTPDelay} onClick={Step1} disabled={uiDisabled || currentStep !== 2} />
                            <OTPInput
                                value={verification}
                                onValueChange={setVerification}
                                disabled={uiDisabled || currentStep !== 2}/>
                        </>
                    }
                    <SubmitButton
                        text={currentStep === 1 ? "Verify Email" : "Verify OTP"}
                        onClick={currentStep === 1 ? Step1 : Step2}
                        disabled={uiDisabled}/>
                    <Divider/>
                    <SSOButtons
                        disabled={uiDisabled}/>
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
