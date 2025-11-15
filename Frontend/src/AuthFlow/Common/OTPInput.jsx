import React from 'react'

export default function OTPInput({ value, onValueChange, disabled }){
    return (
        <div className="mt-2">
            <input name="otp" type="text" inputMode="numeric" maxLength={6} value={value} onChange={(e) => onValueChange(e.target.value)} disabled={disabled}
                   className="w-full px-4 py-3 rounded-md bg-[#0b0f14] border border-gray-700 text-sm text-white placeholder:opacity-40" placeholder="Enter OTP" />
        </div>
    )
}