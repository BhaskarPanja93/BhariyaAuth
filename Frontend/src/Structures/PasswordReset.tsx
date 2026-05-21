import {useEffect, useMemo, useRef, useState} from 'react'
import SubmitButton from "../Elements/SubmitButton";
import PasswordInput from "../Elements/PasswordInput";
import {EmailIsValid, OTPIsValid} from "../Utils/Strings";
import {APIRoute} from "../Values/Constants";
import {useLocation, useNavigate} from "react-router";
import NotificationManager from "../Contexts/Notification.tsx";
import ConnectionManager from "../Contexts/Connection.tsx";
import OTPInput from "../Elements/OTPInput";
import EmailInput from "../Elements/EmailInput";
import Countdown from "../Utils/Countdown";
import OTPResendButton from "../Elements/OTPResendButton";

export default function PasswordReset() {
    const navigate = useNavigate();
    const location = useLocation();
    const params = useMemo(() => {return new URLSearchParams(location.search)}, [location.search]);

    const {SendNotification} = NotificationManager();
    const {SendPost} = ConnectionManager()

    const [uiDisabled, setUiDisabled] = useState<boolean>(false)
    const [currentStep, setCurrentStep] = useState<number>(1)
    const otpCountdownRef = useRef<Countdown | null>(null)
    const [OTPDelay, setOTPDelay] = useState<number>(0)
    const [password, setPassword] = useState<string>("")
    const [passwordConfirmation, setPasswordConfirmation] = useState<string>("")
    const [email, setEmail] = useState<string>("")
    const [verification, setVerification] = useState<string>("")
    const currentToken = useRef<string>("")

    const Step1 = () => {
        if (!EmailIsValid(email)) return SendNotification("Email is invalid");

        setUiDisabled(true);
        const form = new FormData();
        form.append("mail", email);
        SendPost(false, false, false, APIRoute,  "/passwordreset/step1", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Please enter the OTP sent to your mail for Password Reset")
                    currentToken.current = data.reply
                    setCurrentStep(2)
                } else if (data.reply) {
                    const countdown = otpCountdownRef.current
                    if (!countdown) {
                        otpCountdownRef.current = new Countdown(data.reply, 0.1, setOTPDelay).start()
                    } else {
                        countdown.resetDuration(data.reply)
                    }
                }
            })
            .catch((error)=>{console.log("PasswordReset Step1 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
        if (!currentToken.current) {
            setCurrentStep(1)
            SendNotification("Something went wrong. Please enter email again")
            return
        }
        if (password !== passwordConfirmation) return SendNotification("Passwords don't match")
        if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", verification);
        form.append("password", password);
        SendPost(false, false, false, APIRoute, "/passwordreset/step2", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Password changed successfully")
                    navigate(location.state?.return_to || params.get("return_to") || "/")
                }
            })
            .catch((error)=>{console.log("PasswordReset Step2 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    useEffect(() => {
        document.title = "PasswordReset - Bhariya";
        return () => {
            otpCountdownRef.current?.cancel()
        }
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
                                    disabled={uiDisabled || currentStep !== 2} />
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


