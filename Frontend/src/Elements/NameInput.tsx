import React from 'react'

export default function NameInput(
    { value, onValueChange, disabled }:
    {
        value:string,
        onValueChange:React.Dispatch<React.SetStateAction<string>>,
        disabled:boolean
    }){
    return (
        <div>
            <input
                value={value}
                onChange={(e) => onValueChange(e.target.value)}
                disabled={disabled}
                name="name" type="text" required className="w-full px-3 py-3 rounded-md bg-transparent border border-gray-700 text-sm text-white placeholder:opacity-40" placeholder="Name" autoComplete="name"/>
        </div>
    )
}

