import React from 'react'

export default function Step2Toggle({ usingOTP, toggleUsingOTP }){
    return (
        <button type="button" onClick={toggleUsingOTP} className="text-xs text-indigo-400 hover:underline">{usingOTP ? 'Use Password' : 'Use OTP'}</button>
    )
}