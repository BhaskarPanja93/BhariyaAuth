import React, { useState } from "react";

function DeviceIcon({ type }) {
    if (type === "mobile") {
        return (
            <svg width="44" height="44" viewBox="0 0 24 24" fill="none" className="flex-none">
                <rect x="7" y="2" width="10" height="20" rx="2" fill="#0b0f14" stroke="#374151" />
                <circle cx="12" cy="18" r="1" fill="#6b7280" />
            </svg>
        );
    }
    return (
        <svg width="44" height="44" viewBox="0 0 24 24" fill="none" className="flex-none">
            <rect x="2" y="4" width="20" height="12" rx="1.5" fill="#0b0f14" stroke="#374151" />
            <rect x="6" y="18" width="12" height="1.5" rx="0.75" fill="#374151" />
        </svg>
    );
}

function formatDate(dtString) {
    try {
        const d = new Date(dtString);
        return d.toLocaleString();
    } catch {
        return dtString;
    }
}

export default function Sessions() {
    const [sessions, setSessions] = useState([
        {
            id: "s-1",
            type: "desktop",
            device: "MacBook Pro",
            os: "macOS 14.1",
            browser: "Brave",
            firstSeen: "2025-10-15T08:12:00Z",
            lastSeen: "2025-11-19T09:22:00Z",
            isCurrent: true,
        },
        {
            id: "s-2",
            type: "mobile",
            device: "iPhone 13",
            os: "iOS 17.0",
            browser: "Safari",
            firstSeen: "2025-09-03T11:03:00Z",
            lastSeen: "2025-11-18T20:01:00Z",
            isCurrent: false,
        },
        {
            id: "s-3",
            type: "desktop",
            device: "Windows 11 PC",
            os: "Windows 11",
            browser: "Chrome",
            firstSeen: "2025-07-01T06:45:00Z",
            lastSeen: "2025-11-17T21:30:00Z",
            isCurrent: false,
        },
    ]);

    const [loading, setLoading] = useState(false);

    const revokeAll = async () => {
        if (!confirm("Revoke all sessions except this one?")) return;
        setLoading(true);
        try {
            await new Promise((r) => setTimeout(r, 700));
            setSessions((prev) => prev.filter((s) => s.isCurrent));
        } finally {
            setLoading(false);
        }
    };

    const signOut = async (id) => {
        if (!confirm("Sign out of this device?")) return;
        setLoading(true);
        try {
            await new Promise((r) => setTimeout(r, 500));
            setSessions((prev) => prev.filter((s) => s.id !== id));
        } finally {
            setLoading(false);
        }
    };

    const current = sessions.find((s) => s.isCurrent);
    const others = sessions.filter((s) => !s.isCurrent);

    return (
        <div className="min-h-screen flex items-center justify-center">
            <div className="w-full max-w-[1200px]">

                <div
                    className="rounded-2xl p-8 shadow-2xl"
                    style={{
                        background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                        border: "1px solid rgba(255,255,255,0.02)",
                    }}
                >
                    {/* Title Row */}
                    <div className="flex items-center justify-between mb-8">
                        <h1 className="text-xl md:text-2xl font-semibold text-white">Your devices where you are signed
                            in</h1>

                        <button
                            onClick={revokeAll}
                            disabled={loading}
                            className="px-4 py-2 rounded-md text-sm bg-red-600 hover:bg-red-700 text-white disabled:opacity-50"
                        >
                            Revoke all
                        </button>
                    </div>

                    {/* Current Device */}
                    {current && (
                        <div className="mb-10">
                            <div className="text-sm text-gray-400 mb-2">Current device</div>

                            <div className="flex items-center gap-5 p-5 rounded-lg bg-[#0b0f14] border border-gray-800">
                                <div
                                    className="w-14 h-14 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex items-center justify-center">
                                    <DeviceIcon type={current.type}/>
                                </div>

                                <div className="flex-1">
                                    <div className="font-medium text-white">
                                        {current.device} <span className="text-xs text-gray-400">· {current.os}</span>
                                    </div>
                                    <div className="text-xs text-gray-400">Browser: {current.browser}</div>

                                    <div className="text-xs text-gray-400 mt-2">
                                        <div>First seen: {formatDate(current.firstSeen)}</div>
                                        <div>Last active: {formatDate(current.lastSeen)}</div>
                                    </div>
                                </div>

                                <button
                                    disabled
                                    className="px-3 py-2 rounded-md border border-gray-700 text-gray-500 cursor-not-allowed"
                                >
                                    Current
                                </button>
                            </div>
                        </div>
                    )}

                    {/* Other Devices */}
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <h2 className="text-lg font-semibold text-white">Other devices</h2>
                            <span className="text-sm text-gray-400">{others.length} active</span>
                        </div>

                        <div className="max-h-[50vh] overflow-y-auto pr-2 space-y-3">
                            {others.length === 0 && (
                                <p className="text-sm text-gray-400">No other active devices.</p>
                            )}

                            {others.map((s) => (
                                <div
                                    key={s.id}
                                    className="flex items-center gap-5 p-4 rounded-lg bg-[#0b0f14] border border-gray-800"
                                >
                                    <div
                                        className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex items-center justify-center">
                                        <DeviceIcon type={s.type}/>
                                    </div>

                                    <div className="flex-1">
                                        <div className="text-white font-medium text-sm">
                                            {s.device} <span className="text-xs text-gray-400">· {s.os}</span>
                                        </div>
                                        <div className="text-xs text-gray-400">Browser: {s.browser}</div>

                                        <div className="text-xs text-gray-400 mt-2">
                                            <div>First seen: {formatDate(s.firstSeen)}</div>
                                            <div>Last active: {formatDate(s.lastSeen)}</div>
                                        </div>
                                    </div>

                                    <button
                                        onClick={() => signOut(s.id)}
                                        className="px-3 py-1 text-sm rounded-md bg-red-600 hover:bg-red-700 text-white"
                                    >
                                        Sign out
                                    </button>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
                </div>
            </div>
            );
            }
