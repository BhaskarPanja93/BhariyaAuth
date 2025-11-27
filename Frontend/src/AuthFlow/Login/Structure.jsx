import {useState} from 'react'
import {Link} from "react-router-dom";
import {BackendURL} from '../../Values/Constants.js'
import EmailInput from '../Common/EmailInput'
import Step2Toggle from './Step2Toggle'
import OTPInput from '../Common/OTPInput'
import PasswordInput from '../Common/PasswordInput'
import RememberCheckbox from '../Common/RememberCheckbox'
import SubmitButton from '../Common/SubmitButton'
import SSOButtons from '../Common/SSOButtons.jsx'
import {FetchConnectionManager} from "../../Contexts/Connection.jsx";
import Divider from "../Common/Divider.jsx";

export default function LoginStructure() {
    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [useOtp, setUseOtp] = useState(false)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState("")

    const {publicAPI, privateAPI, OpenAuthPopup, Logout} = FetchConnectionManager()
    const [tokens, setTokens] = useState ({})

    const Step1 = async () => {
        if (tokens[email] && tokens[email][useOtp]) {
            setCurrentStep(2)
            return
        }
        setUiDisabled(true);
        const form = new FormData();
        form.append("mail_address", email);
        form.append("remember_me", remember ? "yes" : "no");
        privateAPI.post(BackendURL + `/login/step1/${useOtp?"otp":"password"}`, form, {
        })
            .then((data) => {
                if (data["success"]) {

                }
            })
            .catch((error) => {

            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = async () => {
        setUiDisabled(true);
        const form = new FormData();
        form.append("token", tokens[email][useOtp]);
        form.append("verification", verification);
        privateAPI.post(BackendURL + "/login/step2", form, {
        })
            .then((data) => {
                if (data["success"]) {

                }
            })
            .catch((error) => {
                console.log(error);
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const toggleStep2 = () => {
        setUseOtp(u => !u);
        Step1().then((success) => {
            if (!success) {
                setUseOtp(u => !u);
            }
        })
    }

    return (<div className="min-h-screen flex items-center justify-center">
        <div className="w-full max-w-sm">
            <div className="rounded-2xl p-8 shadow-2xl" style={{
                background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                border: '1px solid rgba(255,255,255,0.02)'
            }}>
                <div className="flex flex-col items-center gap-4 mb-4">
                    <h2 className="text-xl font-semibold text-white">Sign In</h2>
                    <p className="text-sm text-gray-400">
                        {currentStep === 1 ? "Access your account" : (<>{email} <div onClick={()=>setCurrentStep(1)}>Not you?</div></>)}
                    </p>
                </div>
                <div className="space-y-4">
                    <EmailInput value={email} onValueChange={setEmail} disabled={uiDisabled || currentStep !== 1}/>
                    <div className="flex items-center justify-between text-xs text-gray-400">
                        <Step2Toggle usingOTP={useOtp} toggleUsingOTP={toggleStep2} disabled={uiDisabled}/>
                        {
                            (tokens[email] && tokens[email][useOtp]) && (
                                <>
                                    {useOtp ? (
                                        <OTPInput value={verification} onValueChange={setVerification} disabled={uiDisabled} />
                                    ) : (
                                        <PasswordInput value={verification} onValueChange={setVerification} disabled={uiDisabled} />
                                    )}
                                </>
                            )
                        }

                        <div className="flex items-center gap-3">
                            {!useOtp && (
                                    <Link to="/passwordreset" className="text-xs text-indigo-400 hover:underline">
                                        Forgot Password?
                                    </Link>
                                )
                            }
                        </div>
                    </div>
                    {currentStep === 2 && (
                        <>
                            {useOtp ?
                                <OTPInput value={verification} onValueChange={setVerification} disabled={uiDisabled}/>
                            :
                                <PasswordInput value={verification} onValueChange={setVerification} disabled={uiDisabled}/>
                            }
                        </>)}
                    <RememberCheckbox checked={remember} onCheckedChange={setRemember} disabled={uiDisabled || currentStep !== 1}/>
                    {currentStep === 1 ?
                        <SubmitButton text={useOtp?"SEND OTP":"ENTER PASSWORD"} onClick={Step1} disabled={uiDisabled || currentStep !== 1}/>
                        :
                        <SubmitButton text={"SIGN IN"} onClick={Step2} disabled={uiDisabled}/>}
                    <Divider/>
                    <SSOButtons disabled={uiDisabled}/>
                    <p className="text-center text-sm text-gray-500 mt-4">
                        New here?
                        <Link to="/register" className="text-indigo-400 hover:underline">
                            Create an account
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    </div>)
}
