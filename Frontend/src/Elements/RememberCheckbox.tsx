import React from 'react'

export default function RememberCheckbox(
    { checked, onCheckedChange, disabled }:
    {
        checked:boolean,
        onCheckedChange:React.Dispatch<React.SetStateAction<boolean>>,
        disabled:boolean
    }) {
    return (
        <div className="flex items-center gap-3">
            <label className="inline-flex items-center text-sm text-gray-300">
                <input
                    checked={checked}
                    onChange={(e) => onCheckedChange(e.target.checked)}
                    disabled={disabled}
                    name="remember" type="checkbox" className="form-checkbox h-4 w-4 text-indigo-500 rounded focus:ring-0" />
                <span className="ml-2 text-gray-300">Keep me signed in</span>
            </label>
        </div>
    )
}