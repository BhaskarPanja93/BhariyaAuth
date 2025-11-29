import React, { useEffect, useState } from "react";

/* Simple SVG icons */
function DeviceIcon({ type }) {
    if (type === "mobile") {
        return (
            <svg width="44" height="44" viewBox="0 0 24 24" fill="none">
                <rect x="7" y="2" width="10" height="20" rx="2" stroke="#374151" fill="#0b0f14"/>
                <circle cx="12" cy="18" r="1" fill="#6b7280"/>
            </svg>
        );
    }
    return (
        <svg width="44" height="44" viewBox="0 0 24 24" fill="none">
            <rect x="2" y="4" width="20" height="12" rx="1.5" stroke="#374151" fill="#0b0f14"/>
            <rect x="6" y="18" width="12" height="1.5" rx="0.75" fill="#374151"/>
        </svg>
    );
}

function formatDate(dt) {
    try {
        return new Date(dt).toLocaleString();
    } catch {
        return dt;
    }
}

export default function SessionsStructure() {
    const [sessions, setSessions] = useState([]);
    const [currentDeviceId, setCurrentDeviceId] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {

        // -------------------------------------------------------
        // 🔹 DUMMY DATA FOR TESTING — REMOVE WHEN CONNECTING API
        // -------------------------------------------------------
        const dummy = {
            user_id: "user_123",
            device_id: "s-1", // current device
            activities: [
                {
                    id: "s-1",
                    device: "MacBook Pro",
                    os: "macOS 14.1",
                    browser: "Brave",
                    count: 21,
                    remembered: true,
                    creation: "2025-10-15T08:12:00Z",
                    updated: "2025-11-19T09:22:00Z"
                },
                {
                    id: "s-2",
                    device: "iPhone 13",
                    os: "iOS 17.0",
                    browser: "Safari",
                    count: 12,
                    remembered: false,
                    creation: "2025-09-03T11:03:00Z",
                    updated: "2025-11-18T20:01:00Z"
                },{
                    id: "s-3",
                    device: "iPhone 13",
                    os: "iOS 17.0",
                    browser: "Safari",
                    count: 12,
                    remembered: false,
                    creation: "2025-09-03T11:03:00Z",
                    updated: "2025-11-18T20:01:00Z"
                },{
                    id: "s-4",
                    device: "iPhone 13",
                    os: "iOS 17.0",
                    browser: "Safari",
                    count: 12,
                    remembered: false,
                    creation: "2025-09-03T11:03:00Z",
                    updated: "2025-11-18T20:01:00Z"
                },{
                    id: "s-5",
                    device: "iPhone 13",
                    os: "iOS 17.0",
                    browser: "Safari",
                    count: 12,
                    remembered: false,
                    creation: "2025-09-03T11:03:00Z",
                    updated: "2025-11-18T20:01:00Z"
                },{
                    id: "s-6",
                    device: "iPhone 13",
                    os: "iOS 17.0",
                    browser: "Safari",
                    count: 12,
                    remembered: false,
                    creation: "2025-09-03T11:03:00Z",
                    updated: "2025-11-18T20:01:00Z"
                },
                {
                    id: "s-7",
                    device: "Windows 11 PC",
                    os: "Windows 11",
                    browser: "Chrome",
                    count: 6,
                    remembered: false,
                    creation: "2025-07-01T06:45:00Z",
                    updated: "2025-11-17T21:30:00Z"
                }
            ]
        };
        // -------------------------------------------------------

        // simulate loading
        setTimeout(() => {
            setCurrentDeviceId(dummy.device_id);

            const mapped = dummy.activities.map(a => ({
                ...a,
                isCurrent: a.id === dummy.device_id,
                firstSeen: a.creation,
                lastSeen: a.updated,
                type: /iphone|android|mobile|ios/i.test(a.device + a.os)
                    ? "mobile"
                    : "desktop"
            }));

            setSessions(mapped);
            setLoading(false);
        }, 500);

    }, []);

    const revokeAll = () => {
        alert("Revoke all (dummy action)");
        setSessions(prev => prev.filter(s => s.isCurrent));
    };

    const signOut = (id) => {
        alert("Sign out device (dummy action)");
        setSessions(prev => prev.filter(s => s.id !== id));
    };

    const current = sessions.find(s => s.isCurrent);
    const others = sessions.filter(s => !s.isCurrent);

    return (
        <div className="min-h-screen p-5 bg-gradient-to-br from-gray-700 via-[#1a1c20] to-[#0b0d10] box-border overflow-hidden">
            <div className="w-full h-full">

                <div
                    className="rounded-2xl p-6 md:p-8 shadow-2xl max-h-[calc(100vh-40px)] flex flex-col box-border mx-auto w-full"
                    style={{
                        background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                        border: "1px solid rgba(255,255,255,0.02)"
                    }}
                >
                    <div className="flex items-center justify-between mb-6">
                        <h1 className="text-xl md:text-2xl font-semibold text-white">
                            Your devices where you are signed in
                        </h1>

                        <button
                            onClick={revokeAll}
                            className="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white rounded-md"
                        >
                            Revoke all
                        </button>
                    </div>

                    {loading ? (
                        <p className="text-sm text-gray-400">Loading sessions...</p>
                    ) : (
                        <>
                            {/* Current device */}
                            {current && (
                                <div className="mb-8">
                                    <p className="text-sm text-gray-400 mb-2">Current device</p>

                                    <div className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800">

                                        <div className="w-14 h-14 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex justify-center items-center">
                                            <DeviceIcon type={current.type}/>
                                        </div>

                                        <div className="flex-1">
                                            <p className="text-white font-medium">
                                                {current.device}
                                                <span className="text-xs text-gray-400"> · {current.os}</span>
                                            </p>
                                            <p className="text-xs text-gray-400">Browser: {current.browser}</p>

                                            <p className="mt-2 text-xs text-gray-400">
                                                First seen: {formatDate(current.firstSeen)}<br/>
                                                Last active: {formatDate(current.lastSeen)}
                                            </p>
                                        </div>

                                        <button
                                            disabled
                                            className="px-3 py-2 text-sm border border-gray-700 text-gray-500 rounded-md cursor-not-allowed"
                                        >
                                            Current
                                        </button>
                                    </div>
                                </div>
                            )}

                            {/* Other devices */}
                            <div>
                                <div className="flex items-center justify-between mb-3">
                                    <h2 className="text-lg font-semibold text-white">Other devices</h2>
                                    <span className="text-sm text-gray-400">{others.length} active</span>
                                </div>

                                <div className="max-h-[60vh] overflow-y-auto pr-2 space-y-3">
                                    {others.map(s => (
                                        <div
                                            key={s.id}
                                            className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800"
                                        >
                                            <div className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex justify-center items-center">
                                                <DeviceIcon type={s.type}/>
                                            </div>

                                            <div className="flex-1">
                                                <p className="text-white font-medium text-sm">
                                                    {s.device}
                                                    <span className="text-xs text-gray-400"> · {s.os}</span>
                                                </p>
                                                <p className="text-xs text-gray-400">Browser: {s.browser}</p>

                                                <p className="mt-2 text-xs text-gray-400">
                                                    First seen: {formatDate(s.firstSeen)}<br/>
                                                    Last active: {formatDate(s.lastSeen)}
                                                </p>
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
                        </>
                    )}

                </div>

            </div>
        </div>
    );
}
