import {type RefObject, useEffect, useMemo, useRef, useState} from 'react'
import {Link, useLocation, useNavigate} from "react-router";
import {APIRoute} from '../Values/Constants'
import EmailInput from '../Elements/EmailInput'
import Step2Toggle from '../Elements/Step2Toggle'
import OTPInput from '../Elements/OTPInput'
import PasswordInput from '../Elements/PasswordInput'
import RememberCheckbox from '../Elements/RememberCheckbox'
import SubmitButton from '../Elements/SubmitButton'
import SSOButtons from '../Elements/SSOButtons'
import ConnectionManager from "../Contexts/Connection.tsx";
import {EmailIsValid, OTPIsValid, PasswordIsStrong} from "../Utils/Strings";
import Divider from "../Elements/Divider";
import NotificationManager from "../Contexts/Notification.tsx";
import Countdown from "../Utils/Countdown";
import OTPResendButton from "../Elements/OTPResendButton";

export default function LoginPage() {
    const navigate = useNavigate()
    const location = useLocation();
    const params = useMemo(() => {return new URLSearchParams(location.search)}, [location.search]);

    const {SendNotification} = NotificationManager();
    const {SendAPIRequest} = ConnectionManager()

    const [uiDisabled, setUiDisabled] = useState<boolean>(false)
    const [currentStep, setCurrentStep] = useState<number>(1)
    const [useOtp, setUseOtp] = useState<boolean>(false)
    const [OTPDelay, setOTPDelay] = useState<number>(0)
    const [remember, setRemember] = useState<boolean>(false)
    const [email, setEmail] = useState<string>("")
    const [verification, setVerification] = useState<string>("")
    const otpCountdownRef = useRef<Countdown>(undefined)
    const tokens: RefObject<Record<string, Record<number, string>>> = useRef({})

    const Step1 = (usingOTP: boolean, forceResendOTP: boolean) => {
        if (!EmailIsValid(email)) {
            setUiDisabled(false)
            SendNotification("Email is invalid")
            return
        }
        if (!tokens.current[email]) tokens.current[email] = {}
        if (tokens.current[email][usingOTP ? 1 : 0] && !forceResendOTP) {
            setUseOtp(usingOTP)
            setCurrentStep(2)
            setUiDisabled(false)
            return
        }

        setUiDisabled(true)
        setVerification("")
        const form = new FormData();
        form.append("mail", email);
        form.append("remember", remember ? "yes" : "no");
        form.append("process", usingOTP ? "otp" : "password");
        SendAPIRequest("POST", false, false, false, false, APIRoute, "/signin/step1", form)
            .then((data) => {
                if (data.success) {
                    tokens.current[email][usingOTP ? 1 : 0] = data.reply as string
                    setUseOtp(usingOTP)
                    setCurrentStep(2)
                    SendNotification(`Please enter the ${usingOTP ? "OTP" : "Password"}`)
                } else if (usingOTP && data.reply) {
                    const countdown = otpCountdownRef.current
                    if (!countdown) {
                        otpCountdownRef.current = new Countdown(data.reply as number, 0.1, setOTPDelay).start()
                    } else {
                        countdown.resetDuration(data.reply as number)
                    }
                }
            })
            .catch((error) => {
                console.log("SignIn Step1 stopped because:", error)
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = () => {
        if (!tokens.current[email] || !tokens.current[email][useOtp ? 1 : 0]) {
            setCurrentStep(1)
            SendNotification("Something went wrong. Please enter email again")
            return
        }
        if (!useOtp && !PasswordIsStrong(verification)) return SendNotification("Incorrect Password")
        if (useOtp && !OTPIsValid(verification)) return SendNotification("Incorrect OTP")

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", tokens.current[email][useOtp ? 1 : 0]);
        form.append("verification", verification);
        SendAPIRequest("POST", false, false, true, true, APIRoute, "/signin/step2", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Logged In Successfully")
                    navigate(location.state?.return_to || params.get("return_to") || "/", {replace: true})
                }
            })
            .catch((error) => {
                console.log("SignIn Step2 stopped because:", error)
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    useEffect(() => {
        document.title = "SignIn - Bhariya";
        return () => {
            otpCountdownRef.current?.cancel()
        }
    }, []);

    return (
        <div className="min-h-screen flex items-center justify-center">
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
                                          onClick={() => {
                                              setUseOtp(false);
                                              setCurrentStep(1);
                                          }}>
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
                                    toggleUsingOTP={() => Step1(!useOtp, false)}
                                    disabled={uiDisabled}/>
                                <div className="flex items-center gap-3">
                                    {
                                        !useOtp ?
                                            <Link className="text-xs text-indigo-400 hover:underline"
                                                  to="/passwordreset"
                                                  state={{return_to:"/signin"}}
                                            >
                                                Forgot Password?
                                            </Link>
                                            :
                                            <OTPResendButton delay={OTPDelay} onClick={() => Step1(true, true)} disabled={uiDisabled || currentStep !== 2}/>
                                    }
                                </div>
                            </div>
                            {(currentStep === 2) &&
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
                                            disabled={uiDisabled || currentStep !== 2}
                                            needsConfirm={false}
                                            confirm={""}
                                            onConfirmChange={()=>{}}/>
                                    }
                                </div>}
                        </div>
                        <SubmitButton
                            text={currentStep === 1 ? "Continue with Email" : "Sign In"}
                            onClick={currentStep === 1 ? () => Step1(false, false) : Step2}
                            disabled={uiDisabled}/>
                        <Divider/>
                        <SSOButtons disabled={uiDisabled}/>
                        <p className="text-center text-sm text-gray-500 mt-4">
                            New here?&nbsp;
                            <Link className="text-indigo-400 hover:underline"
                                to="/signup"
                            >
                            Create an account
                            </Link>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    )
}


