import OTPInput from '../Common/OTPInput'
import SubmitButton from "../Common/SubmitButton.jsx";

export default function VerifyOTP({ disabled, verification, setVerification }){
    return (
        <div className="min-h-screen flex items-center justify-center">
            <div className="w-full max-w-sm">
                <div className="rounded-2xl p-8 shadow-2xl" style={{
                    background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                    border: '1px solid rgba(255,255,255,0.02)'
                }}>
                    <div className="flex flex-col items-center gap-4 mb-4">
                        <h2 className="text-xl font-semibold text-white">Account Verification</h2>
                        <p className="text-sm text-gray-400">Enter OTP Code</p>
                    </div>
                    <div className="space-y-4">

                        <OTPInput value={verification} onValueChange={setVerification} disabled={disabled}/>

                        <div className="flex justify-end">
                            <button type="button" className="text-xs text-indigo-400 hover:underline"
                                    disabled={disabled}>
                                Resend OTP
                            </button>
                        </div>

                        <SubmitButton text={"Verify OTP"} disabled={disabled}/>

                    </div>
                </div>
            </div>
        </div>
            )
            }
