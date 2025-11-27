import {useRef, useState} from 'react'
import {Link} from "react-router-dom";

import EmailInput from '../Common/EmailInput'
import PasswordInput from '../Common/PasswordInput'
import RememberCheckbox from '../Common/RememberCheckbox'
import SubmitButton from '../Common/SubmitButton'
import SSOButtons from '../Common/SSOButtons.jsx'
import Divider from '../Common/Divider'
import NameInput from './NameInput'
import OTPInput from "../Common/OTPInput.jsx";
import {BackendURL} from "../../Values/Constants.js";
import {FetchConnectionManager} from "../../Contexts/Connection.jsx";

export default function RegisterPage(){
    const [uiDisabled, setUiDisabled] = useState(false)
    const [currentStep, setCurrentStep] = useState(1)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [passwordConfirmation, setPasswordConfirmation] = useState("")
    const [otp, setOTP] = useState("");
    const [name, setName] = useState("")

    const {publicAPI, privateAPI, OpenAuthPopup, Logout} = FetchConnectionManager()
    let currentToken = "";

    const Step1 = async () => {
        setUiDisabled(true);
        const form = new FormData();
        form.append("mail_address", email);
        form.append("name", name);
        form.append("password", password);
        form.append("remember_me", remember ? "yes" : "no");
        privateAPI.post(BackendURL + "/register/step1", form, {
        })
            .then((response) => {
                if (response.data.success) {
                    console.log(response)
                    currentToken = response.data["reply"];
                    console.log(currentToken)
                    setCurrentStep(2)
                }
            })
            .catch((error) => {
                console.log(error);
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    const Step2 = async () => {
        setUiDisabled(true);
        const form = new FormData();
        form.append("token", currentToken);
        form.append("verification", otp);
        privateAPI.post(BackendURL + "/register/step2", form, {
        })
            .then((response) => {
                if (response.data.success) {

                }
            })
            .catch((error) => {
                console.log(error);
            })
            .finally(() => {
                setUiDisabled(false);
            });
    };

    return (
        <div className="min-h-screen flex items-center justify-center">
            <div className="w-full max-w-sm">
                <div className="rounded-2xl p-8 shadow-2xl" style={{
                    background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                    border: '1px solid rgba(255,255,255,0.02)'
                }}>
                    <div className="flex flex-col items-center gap-4 mb-4">
                        <h2 className="text-xl font-semibold text-white">Sign Up</h2>
                        <p className="text-sm text-gray-400">Create an account</p>
                    </div>
                    <div className="space-y-4">
                        <NameInput value={name} onValueChange={setName} disabled={currentStep !== 1 || uiDisabled}/>
                        <EmailInput value={email} onValueChange={setEmail} disabled={currentStep !== 1 || uiDisabled}/>
                        <PasswordInput disabled={currentStep !== 1 || uiDisabled} value={password} onValueChange={setPassword}
                                       confirm={passwordConfirmation} onConfirmChange={setPasswordConfirmation} needsConfirm={true}/>
                        <RememberCheckbox checked={remember} onCheckedChange={setRemember} disabled={currentStep !== 1 || uiDisabled}/>
                        <Divider/>
                        {currentStep === 2 && <OTPInput value={otp} onValueChange={setOTP} disabled={uiDisabled}/>}
                        <SubmitButton text={currentStep===1?"CONTINUE WITH OTP":"VERIFY OTP"} onClick={currentStep===1?Step1:Step2} disabled={uiDisabled}/>
                        <SSOButtons disabled={uiDisabled}/>
                        <p className="text-center text-sm text-gray-500 mt-4">
                            Already have an account?
                            <Link to="/login" className="text-indigo-400 hover:underline">
                                SignIn
                            </Link>
                        </p>
                    </div>
                </div>
            </div>
        </div>
            )
            }
