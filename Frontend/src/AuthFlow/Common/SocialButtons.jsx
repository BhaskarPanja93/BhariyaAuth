import React from 'react'

export default function SocialButtons({ disabled }) {
    return (
        <div className="space-y-3">
            <button type="button" disabled={disabled}
                    className="w-full flex items-center gap-3 px-3 py-2 rounded-md bg-[#0b0f14] border border-gray-800 text-sm text-gray-300 hover:border-gray-500">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="flex-none">
                    <path
                        d="M21.35 11.1h-9.18v2.8h5.26c-.23 1.26-1.3 3.5-5.26 3.5-3.16 0-5.74-2.6-5.74-5.8 0-3.2 2.58-5.8 5.74-5.8 1.8 0 3 0.78 3.7 1.44l2.52-2.43C17.46 3.22 15.06 2 12 2 6.48 2 2 6.48 2 12s4.48 10 10 10c5.73 0 9.82-4.02 9.82-9.7 0-.65-.07-1.14-.47-1.2z"
                        fill="#fff"/>
                </svg>
                <span className="flex-1 text-left">Continue with Google</span>
            </button>
            <button type="button" disabled={disabled}
                    className="w-full flex items-center gap-3 px-3 py-2 rounded-md bg-[#0b0f14] border border-gray-800 text-sm text-gray-300 hover:border-gray-500">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="flex-none">
                    <path
                        d="M21.35 11.1h-9.18v2.8h5.26c-.23 1.26-1.3 3.5-5.26 3.5-3.16 0-5.74-2.6-5.74-5.8 0-3.2 2.58-5.8 5.74-5.8 1.8 0 3 0.78 3.7 1.44l2.52-2.43C17.46 3.22 15.06 2 12 2 6.48 2 2 6.48 2 12s4.48 10 10 10c5.73 0 9.82-4.02 9.82-9.7 0-.65-.07-1.14-.47-1.2z"
                        fill="#fff"/>
                </svg>
                <span className="flex-1 text-left">Continue with Discord</span>
            </button>
            <button type="button" disabled={disabled}
                    className="w-full flex items-center gap-3 px-3 py-2 rounded-md bg-[#0b0f14] border border-gray-800 text-sm text-gray-300 hover:border-gray-500">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="flex-none">
                    <path
                        d="M21.35 11.1h-9.18v2.8h5.26c-.23 1.26-1.3 3.5-5.26 3.5-3.16 0-5.74-2.6-5.74-5.8 0-3.2 2.58-5.8 5.74-5.8 1.8 0 3 0.78 3.7 1.44l2.52-2.43C17.46 3.22 15.06 2 12 2 6.48 2 2 6.48 2 12s4.48 10 10 10c5.73 0 9.82-4.02 9.82-9.7 0-.65-.07-1.14-.47-1.2z"
                        fill="#fff"/>
                </svg>
                <span className="flex-1 text-left">Continue with Microsoft</span>
            </button>
        </div>
    )
}