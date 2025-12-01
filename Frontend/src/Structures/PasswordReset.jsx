import {useRef, useState} from 'react'
import SubmitButton from "../Elements/SubmitButton.jsx";
import PasswordInput from "../Elements/PasswordInput.jsx";
import {EmailIsValid, OTPIsValid} from "../Utils/Strings.js";
import {BackendURL} from "../Values/Constants.js";
import {useNavigate} from "react-router-dom";
import {FetchNotificationManager} from "../Contexts/Notification.jsx";
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import OTPInput from "../Elements/OTPInput.jsx";
import EmailInput from "../Elements/EmailInput.jsx";

export default function PasswordReset({disabled}) {
    const navigate = useNavigate();
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
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
                    currentToken.current = data["reply"]
                    setCurrentStep(2)
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
                    navigate("/sessions")
                }
            })
            .catch((error)=>{console.log("PasswordReset Step2 stopped because:", error)})
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
                        <h2 className="text-xl font-semibold text-white">
                            Reset Password
                        </h2>
                    </div>
                    <div className="space-y-4">
                        {currentStep === 1 ?
                            <>
                                <EmailInput value={email} onValueChange={setEmail} disabled={uiDisabled || currentStep !== 1}/>
                                <SubmitButton text={"Send OTP"} onClick={Step1} disabled={disabled || currentStep !== 1}/>
                            </>
                            :
                            <>
                                <OTPInput value={verification} onValueChange={setVerification} disabled={uiDisabled || currentStep !== 2}/>
                                <PasswordInput disabled={disabled || currentStep !== 2} value={password} onValueChange={setPassword} confirm={passwordConfirmation} onConfirmChange={setPasswordConfirmation} needsConfirm={true}/>
                                <SubmitButton text={"Update Password"} onClick={Step2} disabled={disabled || currentStep !== 2}/>
                            </>
                        }
                    </div>
                </div>
            </div>
        </div>)
}
