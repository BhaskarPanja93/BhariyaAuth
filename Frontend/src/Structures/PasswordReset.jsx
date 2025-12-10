import {useEffect, useRef, useState} from 'react'
import SubmitButton from "../Elements/SubmitButton.jsx";
import PasswordInput from "../Elements/PasswordInput.jsx";
import {EmailIsValid, OTPIsValid} from "../Utils/Strings.js";
import {BackendURL} from "../Values/Constants.js";
import {useNavigate} from "react-router-dom";
import {FetchNotificationManager} from "../Contexts/Notification.jsx";
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import OTPInput from "../Elements/OTPInput.jsx";
import EmailInput from "../Elements/EmailInput.jsx";
import {Countdown} from "../Utils/Countdown.js";
import OTPResendButton from "../Elements/OTPResendButton.jsx";

export default function PasswordReset({disabled}) {
    const navigate = useNavigate();
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const OTPResendTimerID = useRef(0)
    const [OTPDelay, setOTPDelay] = useState(0)
    const [password, setPassword] = useState("")
    const [passwordConfirmation, setPasswordConfirmation] = useState("")
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState()
    const currentToken = useRef("")

    const Step1 = () => {
        if (!EmailIsValid(email)) return SendNotification("Email is invalid");

        setUiDisabled(true);
        const form = new FormData();
        form.append("mail_address", email);
        privateAPI.post(BackendURL + "/passwordreset/step1/", form)
            .then((data) => {
                if (data["success"]) {
                    SendNotification("Please enter the OTP sent to your mail")
                    currentToken.current = data["reply"]
                    setCurrentStep(2)
                } else if (data["reply"]) {
                    Countdown(data["reply"], OTPResendTimerID, setOTPDelay).then()
                }
            })
            .catch((error)=>{console.log("PasswordReset Step1 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
        if (!currentToken.current) return SendNotification("Step 1 incomplete. Please resend OTP");
        if (password !== passwordConfirmation) return SendNotification("Passwords don't match")
        if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", verification);
        form.append("new_password", password);
        privateAPI.post(BackendURL + "/passwordreset/step2", form)
            .then((data) => {
                if (data["success"]) {
                    SendNotification("Password changed successfully")
                    navigate("/sessions")
                }
            })
            .catch((error)=>{console.log("PasswordReset Step2 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    useEffect(() => {
        document.title = "PasswordReset - Bhariya";
    }, [])

    return (<div className="min-h-screen flex items-center justify-center">
            <div className="w-full max-w-sm">
                <div className="rounded-2xl p-8 shadow-2xl"
                     style={{
                        background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                        border: '1px solid rgba(255,255,255,0.02)'
                    }}>
                    <div className="flex flex-col items-center gap-4 mb-4">
                        <h2 className="text-xl font-semibold text-white">
                            Reset Password
                        </h2>
                    </div>
                    <div className="space-y-4">
                        <EmailInput
                            value={email}
                            onValueChange={setEmail}
                            disabled={uiDisabled || currentStep !== 1}
                            hidden={currentStep === 2}/>
                        {currentStep === 2 &&
                            <>
                                <OTPInput
                                    value={verification}
                                    onValueChange={setVerification}
                                    disabled={uiDisabled || currentStep !== 2}/>
                                <div className="flex justify-end">
                                    <OTPResendButton delay={OTPDelay} onClick={Step1} disabled={uiDisabled || currentStep !== 2} />
                                </div>
                                <PasswordInput
                                    value={password} onValueChange={setPassword}
                                    needsConfirm={true}
                                    confirm={passwordConfirmation}
                                    onConfirmChange={setPasswordConfirmation}
                                    disabled={disabled || currentStep !== 2} />
                            </>
                        }
                        <SubmitButton
                            text={currentStep === 1 ? "Send OTP" : "Update Password"}
                            onClick={currentStep === 1 ? Step1 : Step2}
                            disabled={uiDisabled}/>
                    </div>
                </div>
            </div>
        </div>)
}
