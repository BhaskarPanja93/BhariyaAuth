import React from 'react'

export default function EmailInput({ value, onValueChange, disabled, hidden }){
    return (
        <div>
            <input name="email" type="email" required value={value} onChange={(e) => onValueChange(e.target.value)} disabled={disabled}
                   className={`w-full px-3 py-3 rounded-md bg-transparent border border-gray-700 text-sm text-white placeholder:opacity-40 ${hidden&&"hidden"}`} placeholder="Email" autoComplete="email"/>
        </div>
    )
}