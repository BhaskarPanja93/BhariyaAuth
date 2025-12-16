import React from 'react'

export default function SubmitButton({ text, onClick, disabled }) {
    return (
        <div>
            <button onClick={onClick} disabled={disabled} type="submit" className="w-full py-3 rounded-md font-semibold text-sm text-black bg-linear-to-r from-purple-500 to-violet-600 shadow-md transition-all duration-300 hover:brightness-125">{text}</button>
        </div>
    )
}