import React from 'react'


export default function EmailInput({ value, onChange }){
    return (
        <div>
            <input name="email" type="email" required value={value} onChange={onChange}
                   className="w-full px-3 py-3 rounded-md bg-transparent border border-gray-700 text-sm text-white placeholder:opacity-40" placeholder="Email" />
        </div>
    )
}