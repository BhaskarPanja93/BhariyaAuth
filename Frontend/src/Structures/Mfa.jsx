import OTPInput from '../Elements/OTPInput.jsx'
import SubmitButton from "../Elements/SubmitButton.jsx";
import {OTPIsValid} from "../Utils/Strings.js";
import {BackendURL} from "../Values/Constants.js";
import {useNavigate} from "react-router-dom";
import {FetchNotificationManager} from "../Contexts/Notification.jsx";
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import {useEffect, useRef, useState} from "react";
import OTPResendButton from "../Elements/OTPResendButton.jsx";
import {Countdown} from "../Utils/Countdown.js";

export default function Mfa() {
    const navigate = useNavigate();
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI, EnsureLoggedIn} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const OTPResendTimerID = useRef(0)
    const [OTPDelay, setOTPDelay] = useState(0)
    const [verification, setVerification] = useState("")
    const currentToken = useRef("")

    const Step1 = () => {
        EnsureLoggedIn().then(s=> {
            if (!s) return SendNotification("You need to be logged in to send OTP");
            setUiDisabled(true);
            privateAPI.post(BackendURL + "/mfa/step1", {}, {requiresCSRF: true})
                .then((data) => {
                    if (data["success"]) {
                        SendNotification("Please enter the OTP sent to your mail")
                        currentToken.current = data["reply"]
                        setCurrentStep(2)
                    } else if (data["reply"]) {
                        Countdown(data["reply"], 0.1, OTPResendTimerID, setOTPDelay).then()
                    }
                })
                .catch((error) => {
                    console.log("Mfa Step1 stopped because:", error)
                })
                .finally(() => {
                    setUiDisabled(false);
                });
        })
    }

    const Step2 = () => {
        EnsureLoggedIn().then(s=> {
            if (!s) return SendNotification("You need to be logged in to submit OTP");
            if (!currentToken.current) return SendNotification("Step 1 incomplete. Please resend OTP");
            if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

            setUiDisabled(true);
            const form = new FormData();
            form.append("token", currentToken.current);
            form.append("verification", verification);
            privateAPI.post(BackendURL + "/mfa/step2", form, {forMFA: true})
                .then((data) => {
                    if (data["success"]) {
                        SendNotification("Verification complete")
                        navigate("/");
                    }
                })
                .catch((error) => {
                    console.log("Mfa Step2 stopped because:", error)
                })
                .finally(() => {
                    setUiDisabled(false);
                });
        })
    };

    useEffect(() => {
        document.title = "MFA - Bhariya";
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
                        MFA Verification
                    </h2>
                </div>
                <div className="space-y-4">
                    {currentStep === 2 &&
                        <>
                            <div className="flex justify-between">
                                <p className="text-sm text-gray-400">
                                    Enter OTP
                                </p>
                                <OTPResendButton delay={OTPDelay} onClick={Step1} disabled={uiDisabled || currentStep !== 2} />
                            </div>
                            <OTPInput
                                value={verification}
                                onValueChange={setVerification}
                                disabled={uiDisabled || currentStep !== 2}/>
                        </>
                    }
                    <SubmitButton
                        text={currentStep === 1 ? "Send OTP" : "Verify"}
                        onClick={currentStep === 1 ? Step1 : Step2}
                        disabled={uiDisabled}/>
                </div>
            </div>
        </div>
    </div>)
}
