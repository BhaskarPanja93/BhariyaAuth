import React from 'react'
export default function SocialButtons(){
    return (
        <div className="space-y-3">
            <button type="button" className="w-full flex items-center gap-3 px-3 py-2 rounded-md bg-[#0b0f14] border border-gray-800 text-sm text-gray-300 hover:border-gray-500">{/* google */}<span className="flex-1 text-left">Continue with Google</span></button>
            <button type="button" className="w-full flex items-center gap-3 px-3 py-2 rounded-md bg-[#0b0f14] border border-gray-800 text-sm text-gray-300 hover:border-gray-500">{/* discord */}<span className="flex-1 text-left">Continue with Discord</span></button>
            <button type="button" className="w-full flex items-center gap-3 px-3 py-2 rounded-md bg-[#0b0f14] border border-gray-800 text-sm text-gray-300 hover:border-gray-500">{/* microsoft */}<span className="flex-1 text-left">Continue with Microsoft</span></button>
        </div>
    )
}