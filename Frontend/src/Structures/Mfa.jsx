import OTPInput from '../Elements/OTPInput.jsx'
import SubmitButton from "../Elements/SubmitButton.jsx";
import {OTPIsValid} from "../Utils/Strings.js";
import {BackendURL} from "../Values/Constants.js";
import {useNavigate} from "react-router-dom";
import {FetchNotificationManager} from "../Contexts/Notification.jsx";
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import {useEffect, useRef, useState} from "react";

export default function Mfa() {
    const navigate = useNavigate();
    const {SendNotification} = FetchNotificationManager();
    const {privateAPI} = FetchConnectionManager()

    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [verification, setVerification] = useState("")
    const currentToken = useRef("")

    const Step1 = () => {
        setCurrentStep(1)
        setUiDisabled(true);
        privateAPI.post(BackendURL + "/mfa/step1", {}, {requiresCSRF: true})
            .then((data) => {
                if (data["success"]) {
                    currentToken.current = data["reply"]
                }
            })
            .catch((error) => {
                console.log("Mfa Step1 stopped because:", error)
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
        setCurrentStep(2)
        if (!currentToken.current) return SendNotification("Step 1 incomplete. Please resend OTP");
        if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", verification);
        privateAPI.post(BackendURL + "/mfa/step2", form, {forMFA: true})
            .then((data) => {
                if (data["success"]) {
                    navigate("/sessions");
                }
            })
            .catch((error) => {
                console.log("Mfa Step2 stopped because:", error)
            })
            .finally(() => {
                setUiDisabled(false);
            });
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
                    {currentStep === 1 ?
                        <p className="text-sm text-gray-400">
                            Enter OTP
                        </p>
                        :
                        <>
                            <OTPInput
                                value={verification}
                                onValueChange={setVerification}
                                disabled={uiDisabled || currentStep !== 2}/>
                            <div className="flex justify-end">
                                <button className="text-xs text-indigo-400 hover:underline"
                                        type="button"
                                        onClick={Step1}
                                        disabled={uiDisabled || currentStep !== 2}>
                                    Resend OTP
                                </button>
                            </div>
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
