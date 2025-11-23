import {useState} from 'react'
import SubmitButton from "./Common/SubmitButton.jsx";
import PasswordInput from "./Common/PasswordInput.jsx";

export default function ResetPassword({disabled}) {
    const [verification, setVerification] = useState("")
    const [confirmation, setConfirmation] = useState("")

    return (<div className="min-h-screen flex items-center justify-center">
            <div className="w-full max-w-sm">
                <div className="rounded-2xl p-8 shadow-2xl" style={{
                    background: 'linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))',
                    border: '1px solid rgba(255,255,255,0.02)'
                }}>
                    <div className="flex flex-col items-center gap-4 mb-4">
                        <h2 className="text-xl font-semibold text-white">Reset Password</h2>
                    </div>
                    <div className="space-y-4">
                        <PasswordInput disabled={disabled} value={verification} onValueChange={setVerification} confirm={confirmation} onConfirmChange={setConfirmation} needsConfirm={true}/>
                        <SubmitButton text={"Change my Password"} disabled={disabled}/>
                    </div>
                </div>
            </div>
        </div>)
}
