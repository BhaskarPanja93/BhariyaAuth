import {useEffect, useRef, useState} from 'react'
import {Link, useNavigate, useLocation} from "react-router";
import EmailInput from '../Elements/EmailInput'
import PasswordInput from '../Elements/PasswordInput'
import RememberCheckbox from '../Elements/RememberCheckbox'
import SubmitButton from '../Elements/SubmitButton'
import SSOButtons from '../Elements/SSOButtons'
import Divider from '../Elements/Divider'
import NameInput from '../Elements/NameInput'
import OTPInput from "../Elements/OTPInput";
import {APIRoute} from "../Values/Constants";
import ConnectionManager from "../Contexts/Connection.tsx";
import NotificationManager from "../Contexts/Notification.tsx";
import {EmailIsValid, NameIsValid, OTPIsValid, PasswordIsStrong} from "../Utils/Strings";
import Countdown from "../Utils/Countdown";
import OTPResendButton from "../Elements/OTPResendButton";

export default function RegisterPage() {
    const navigate = useNavigate();
    const location = useLocation();
    const params = new URLSearchParams(location.search);

    const {SendNotification} = NotificationManager();
    const {SendPost} = ConnectionManager()

    const [uiDisabled, setUiDisabled] = useState<boolean>(false)
    const [currentStep, setCurrentStep] = useState<number>(1)
    const [OTPDelay, setOTPDelay] = useState<number>(0)
    const [remember, setRemember] = useState<boolean>(false)
    const [email, setEmail] = useState<string>("")
    const [password, setPassword] = useState<string>("")
    const [passwordConfirmation, setPasswordConfirmation] = useState<string>("")
    const [name, setName] = useState<string>("")
    const [verification, setVerification] = useState<string>("")
    const otpCountdownRef = useRef<Countdown>(undefined)
    const currentToken = useRef<string>("")

    const Step1 = () => {
        if (!NameIsValid(name)) return SendNotification("Name is invalid")
        if (!EmailIsValid(email)) return SendNotification("Email is invalid")
        if (!PasswordIsStrong(password)) return SendNotification("Password is too weak")
        if (password !== passwordConfirmation) return SendNotification("Passwords don't match")

        setUiDisabled(true);
        setVerification("")
        const form = new FormData();
        form.append("mail", email);
        form.append("name", name);
        form.append("password", password);
        form.append("remember", remember ? "yes" : "no");
        SendPost(false, false, false, APIRoute, "/signup/step1", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Please enter the OTP sent to your mail")
                    currentToken.current = data.reply;
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
            .catch((error)=>{console.log("SignUp Step1 stopped because:", error)})
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
        if (!OTPIsValid(verification)) return SendNotification("Incorrect OTP");

        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken.current);
        form.append("verification", verification);
        SendPost(false, false, true, APIRoute,"/signup/step2", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Registered and logged in Successfully")
                    navigate(location.state?.return_to || params.get("return_to") || "/", {replace: true})
                }
            })
            .catch((error)=>{console.log("SignUp Step2 stopped because:", error)})
            .finally(() => {
                setUiDisabled(false);
            });
    };

    useEffect(() => {
        document.title = "SignUp - Bhariya";
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
                            Sign Up
                        </h2>
                        <div className="text-sm text-gray-400">
                            {
                                currentStep === 1 ?
                                "Create an account"
                                :
                                <div className="flex items-center gap-2">
                                    <span>{email}</span>
                                    <span className="text-indigo-400 cursor-pointer"
                                        onClick={() => setCurrentStep(1)}>
                                        Not you?
                                    </span>
                                </div>
                            }
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
                                disabled={uiDisabled || currentStep !== 1}
                                hidden={currentStep !== 1}/>
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
                                <OTPResendButton
                                    delay={OTPDelay}
                                    onClick={Step1}
                                    disabled={uiDisabled || currentStep !== 2} />
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
                            <Link to="/signin" className="text-indigo-400 hover:underline">
                                SignIn
                            </Link>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    )
}


