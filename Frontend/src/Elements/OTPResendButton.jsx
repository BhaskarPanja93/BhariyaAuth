import React from 'react'

export default function OTPResendButton({ delay, onClick, disabled }) {
    return (
        <button className="text-xs text-indigo-400 hover:underline"
                type="button"
                onClick={onClick}
                disabled={disabled || delay !== 0}>
            {delay === 0 ? "Resend OTP" : "Resend in "+delay.toFixed(1)}
        </button>
    )
}