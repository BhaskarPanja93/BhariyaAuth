import {useState} from 'react'

export default function PasswordInput({ value, onValueChange, confirm, onConfirmChange, disabled, needsConfirm }){
    const [visible, setVisible] = useState(false);
    return (
        <div className="mt-3">
            <div className="relative">

                <input name='password' type={visible ? 'text' : 'password'} value={value} onChange={(e) => onValueChange(e.target.value)} disabled={disabled} placeholder='Password'
                       className="w-full px-4 py-3 rounded-md bg-[#0b0f14] border border-gray-700 text-sm text-white placeholder:opacity-40" autoComplete="current-password" />

                <button type="button" onClick={()=>setVisible(v => !v)} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 text-sm" aria-label="Toggle password visibility">
                    {visible ? (
                        <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M12 5c4.477 0 8.268 2.943 9.542 7-1.274 4.057-5.065 7-9.542 7S3.732 16.057 2.458 12C3.732 7.943 7.523 5 12 5zm0 5.5a2.5 2.5 0 100 5 2.5 2.5 0 000-5z"/></svg>
                    ) : (
                        <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path strokeLinecap="round" strokeLinejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5s8.268 2.943 9.542 7c-1.274 4.057-5.065 7-9.542 7s-8.268-2.943-9.542-7z"/></svg>
                    )}
                </button>

            </div>

            {needsConfirm && (<input name='password' type='password' value={confirm} onChange={(e) => onConfirmChange(e.target.value)} disabled={disabled} placeholder='Confirm Password'
                                     className="mt-3 w-full px-4 py-3 rounded-md bg-[#0b0f14] border border-gray-700 text-sm text-white placeholder:opacity-40" />)}
        </div>
    )
}