import {lazy, Suspense, useState} from 'react'
import {Link} from "react-router-dom";

const EmailInput = lazy(()=> import('../Common/EmailInput'))
const Step2Toggle = lazy(()=> import('./Step2Toggle'))
const OTPInput = lazy(()=> import('../Common/OTPInput'))
const PasswordInput = lazy(()=> import('../Common/PasswordInput'))
const RememberCheckbox = lazy(()=> import('../Common/RememberCheckbox'))
const SubmitButton = lazy(()=> import('../Common/SubmitButton'))
const SocialButtons = lazy(()=> import('../Common/SocialButtons'))
const Divider = lazy(()=> import('../Common/Divider'))

export default function LoginPage(){
    const [disabled, setDisabled] = useState(false)
    const [useOtp, setUseOtp] = useState(false)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState("")

    return (
        <Suspense fallback={null}>
            <div className="w-full max-w-sm">
                <div className="rounded-2xl p-8 shadow-2xl" style={{ background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))', border: '1px solid rgba(255,255,255,0.02)' }}>
                    <div className="flex flex-col items-center gap-4 mb-4">
                        <h2 className="text-xl font-semibold text-white">Sign In</h2>
                        <p className="text-sm text-gray-400">Access your account</p>
                    </div>
                    <form className="space-y-4">
                        <EmailInput value={email} onValueChange={setEmail} disabled={disabled} />
                        <div className="flex items-center justify-between text-xs text-gray-400">
                            <Step2Toggle useOtp={useOtp} setUseOtp={setUseOtp} disabled={disabled} />
                            <div className="flex items-center gap-3">
                                {useOtp ?
                                    (<button type="button" className="text-xs text-indigo-400 hover:underline">Resend OTP</button>):
                                    (<button type="button" className="text-xs text-indigo-400 hover:underline">Forgot Password</button>)
                                }
                            </div>
                        </div>
                        {useOtp ?
                            (<OTPInput value={verification} onValueChange={setVerification} disabled={disabled} />):
                            (<PasswordInput value={verification} onValueChange={setVerification} disabled={disabled} />)
                        }
                        <RememberCheckbox checked={remember} onCheckedChange={setRemember} disabled={disabled} />
                        <SubmitButton text={"SIGN IN"} disabled={disabled} />
                        <Divider />
                        <SocialButtons disabled={disabled} />
                        <p className="text-center text-sm text-gray-500 mt-4">
                            New here? <Link to="/register" className="text-indigo-400 hover:underline">Create an account</Link>
                        </p>
                    </form>
                </div>
            </div>
        </Suspense>
    )
}
