import React from 'react'
export default function RememberCheckbox({ checked=false, onChange=() => {} }){
    return (
        <div className="flex items-center gap-3">
            <label className="inline-flex items-center text-sm text-gray-300">
                <input name="remember" type="checkbox" checked={checked} onChange={onChange} className="form-checkbox h-4 w-4 text-indigo-500 rounded focus:ring-0" />
                <span className="ml-2 text-gray-300">Keep me signed in</span>
            </label>
        </div>
    )
}