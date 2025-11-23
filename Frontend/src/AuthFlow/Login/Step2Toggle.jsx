import React from 'react'

export default function Step2Toggle({ usingOTP, setUsingOtp }){
    return (
        <button type="button" onClick={() => setUsingOtp(u => !u)} className="text-xs text-indigo-400 hover:underline">{usingOTP ? 'Use Password' : 'Use OTP'}</button>
    )
}