import React from 'react'
export default function Divider(){
    return (
        <div className="flex items-center gap-3 my-2">
            <div className="flex-1" style={{ height: 1, background: 'linear-gradient(90deg,transparent,#2b2f36,transparent)' }} />
            <div className="text-xs text-gray-500">OR</div>
            <div className="flex-1" style={{ height: 1, background: 'linear-gradient(90deg,transparent,#2b2f36,transparent)' }} />
        </div>
    )
}