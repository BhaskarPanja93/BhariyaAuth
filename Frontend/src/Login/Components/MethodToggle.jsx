import React from 'react'


export default function MethodToggle({ useOtp, setUseOtp }){
    return (
        <button type="button" onClick={() => setUseOtp(u => !u)} className="text-xs text-indigo-400 hover:underline">{useOtp ? 'Use Password' : 'Use OTP'}</button>
    )
}