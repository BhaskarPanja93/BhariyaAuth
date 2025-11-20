import {useState} from 'react'
import {Link} from "react-router-dom";

import EmailInput from '../Common/EmailInput'
import PasswordInput from '../Common/PasswordInput'
import RememberCheckbox from '../Common/RememberCheckbox'
import SubmitButton from '../Common/SubmitButton'
import SocialButtons from '../Common/SocialButtons'
import Divider from '../Common/Divider'
import NameInput from './NameInput'

export default function RegisterPage(){
    const [disabled, setDisabled] = useState(false)
    const [useOtp, setUseOtp] = useState(false)
    const [remember, setRemember] = useState(false)
    const [email, setEmail] = useState("")
    const [verification, setVerification] = useState("")
    const [confirmation, setConfirmation] = useState("")
    const [name, setName] = useState("")

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
                    <form className="space-y-4">
                        <NameInput value={name} onValueChange={setName} disabled={disabled}/>
                        <EmailInput value={email} onValueChange={setEmail} disabled={disabled}/>

                        <PasswordInput disabled={disabled} value={verification} onValueChange={setVerification}
                                       confirm={confirmation} onConfirmChange={setConfirmation} needsConfirm={true}/>

                        <RememberCheckbox checked={remember} onCheckedChange={setRemember} disabled={disabled}/>
                        <SubmitButton text={"Verify Email"} disabled={disabled}/>
                        <Divider/>
                        <SocialButtons disabled={disabled}/>
                        <p className="text-center text-sm text-gray-500 mt-4">
                            Already have an account? <Link to="/login" className="text-indigo-400 hover:underline">Sign
                            In</Link>
                        </p>
                    </form>
                </div>
            </div>
        </div>
            )
            }
