import React from 'react';
import { InputOtp } from 'primereact/inputotp';

export default function OTPInput({ value, onValueChange, disabled }) {
    const customInput = ({ events, props }) => (
        <input
            {...events}
            {...props}
            type="text"
            className="
                w-12 h-12
                text-center
                rounded-md
                bg-[#0b0f14]
                border border-gray-700
                text-white
                text-lg
                tracking-widest
                focus:outline-none
                focus:ring-2
                focus:ring-indigo-500
            "
        />
    );

    return (
        <div className="flex gap-2 mt-2">
            <InputOtp
                value={value}
                onChange={(e) => onValueChange(e.value)}
                integerOnly
                disabled={disabled}
                length={6}
                inputTemplate={customInput}
            />
        </div>
    );
}
