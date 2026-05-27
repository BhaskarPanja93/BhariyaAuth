import OTPInput from '../Elements/OTPInput'
import SubmitButton from "../Elements/SubmitButton";
import {OTPIsValid} from "../Utils/Strings";
import {APIRoute} from "../Values/Constants";
import {useLocation, useNavigate} from "react-router";
import ConnectionManager from "../Contexts/Connection";
import NotificationManager from "../Contexts/Notification";
import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import OTPResendButton from "../Elements/OTPResendButton";
import Countdown from "../Utils/Countdown";

export default function Mfa() {
    const navigate = useNavigate();
    const location = useLocation();
    const params = useMemo(() => {return new URLSearchParams(location.search)}, [location.search]);

    const {SendNotification} = NotificationManager();
    const {SendAPIRequest} = ConnectionManager()

    const [uiDisabled, setUiDisabled] = useState<boolean>(false)
    const [currentStep, setCurrentStep] = useState<number>(1)
    const otpCountdownRef = useRef<Countdown>(undefined)
    const [OTPDelay, setOTPDelay] = useState<number>(0)
    const [verification, setVerification] = useState<string>("")
    const currentToken = useRef<string>("")

    const Step1 = useCallback(() => {
        setUiDisabled(true);
        SendAPIRequest("POST", true, false, false, false, APIRoute, "/mfa/step1")
            .then((data) => {
                if (data.success) {
                    SendNotification("Please enter the OTP sent to your mail for MFA")
                    currentToken.current = data.reply as string
                    setCurrentStep(2)
                } else if (data.reply) {
                    const countdown = otpCountdownRef.current
                    if (!countdown) {
                        otpCountdownRef.current = new Countdown(data.reply as number, 0.1, setOTPDelay).start()
                    } else {
                        countdown.resetDuration(data.reply as number)
                    }
                }
            })
            .catch((error) => {
                console.log("Mfa Step1 stopped because:", error)
            })
            .finally(() => {
                setUiDisabled(false);
            });
    },[SendNotification, SendAPIRequest])

    const Step2 = () => {
        if (!currentToken.current) {
            setCurrentStep(1)
            SendNotification("Step 1 incomplete. Please resend OTP")
            return
        }
        if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", verification);
        SendAPIRequest("POST", true, false, true, true, APIRoute, "/mfa/step2", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Verification complete")
                    navigate(location.state?.return_to || params.get("return_to") || "/");
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
                        text={currentStep === 1 ? OTPDelay === 0 ? "Send OTP" : `OTP disabled for  ${OTPDelay.toFixed(0)}s`: "Complete MFA"}
                        onClick={currentStep === 1 ? Step1 : Step2}
                        disabled={uiDisabled || currentStep === 1 && OTPDelay !== 0}/>
                </div>
            </div>
        </div>
    </div>)
}


