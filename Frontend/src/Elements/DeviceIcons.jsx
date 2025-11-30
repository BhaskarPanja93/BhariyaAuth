import React from "react";

export function DeviceIcons({type}) {
    if (type === "mobile") {
        return (<svg width="44" height="44" viewBox="0 0 24 24" fill="none">
            <rect x="7" y="2" width="10" height="20" rx="2" stroke="#374151" fill="#0b0f14"/>
            <circle cx="12" cy="18" r="1" fill="#6b7280"/>
        </svg>);
    }
    if (type === "desktop") {
        return (<svg width="44" height="44" viewBox="0 0 24 24" fill="none">
            <rect x="2" y="4" width="20" height="12" rx="1.5" stroke="#374151" fill="#0b0f14"/>
            <rect x="6" y="18" width="12" height="1.5" rx="0.75" fill="#374151"/>
        </svg>);
    }
    return (<svg width="44" height="44" viewBox="0 0 24 24" fill="none">
        <rect x="4" y="4" width="16" height="16" rx="3" stroke="#374151" fill="#0b0f14"/>
        <text x="12" y="16" fontSize="10" textAnchor="middle" fill="#6b7280" fontFamily="Arial, sans-serif">?</text>
    </svg>);
}