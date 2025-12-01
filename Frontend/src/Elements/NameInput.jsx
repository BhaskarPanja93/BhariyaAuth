import React from 'react'

export default function NameInput({ value, onValueChange, disabled }){
    return (
        <div>
            <input name="name" type="text" required value={value} onChange={(e) => onValueChange(e.target.value)} disabled={disabled}
                   className="w-full px-3 py-3 rounded-md bg-transparent border border-gray-700 text-sm text-white placeholder:opacity-40" placeholder="Name" autoComplete="name"/>
        </div>
    )
}