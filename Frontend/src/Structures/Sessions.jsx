import React, {useEffect, useState} from "react";

function DeviceIcon({type}) {
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

function formatDate(dt) {
    try {
        return new Date(dt).toLocaleString();
    } catch {
        return dt;
    }
}

export default function Sessions() {
    const [sessions, setSessions] = useState([]);
    const [loading, setLoading] = useState(true);
    const [busyIds, setBusyIds] = useState(new Set());

    useEffect(() => {
        // -------------------------------------------------------
        // 🔹 DUMMY DATA FOR TESTING — REMOVE WHEN CONNECTING API
        // -------------------------------------------------------
        const dummy = {
            user_id: "user_123", device_id: "s-1", // current device
            activities: [{
                id: "s-1",
                device: "MacBook Pro",
                os: "macOS 14.1",
                browser: "Brave",
                count: 21,
                remembered: true,
                creation: "2025-10-15T08:12:00Z",
                updated: "2025-11-19T09:22:00Z"
            }, {
                id: "s-2",
                device: "iPhone 13",
                os: "iOS 17.0",
                browser: "Safari",
                count: 12,
                remembered: false,
                creation: "2025-09-03T11:03:00Z",
                updated: "2025-11-18T20:01:00Z"
            }, {
                id: "s-3",
                device: "iPhone 13",
                os: "iOS 17.0",
                browser: "Safari",
                count: 12,
                remembered: false,
                creation: "2025-09-03T11:03:00Z",
                updated: "2025-11-18T20:01:00Z"
            }, {
                id: "s-4",
                device: "iPhone 13",
                os: "iOS 17.0",
                browser: "Safari",
                count: 12,
                remembered: false,
                creation: "2025-09-03T11:03:00Z",
                updated: "2025-11-18T20:01:00Z"
            }, {
                id: "s-5",
                device: "iPhone 13",
                os: "iOS 17.0",
                browser: "Safari",
                count: 12,
                remembered: false,
                creation: "2025-09-03T11:03:00Z",
                updated: "2025-11-18T20:01:00Z"
            }, {
                id: "s-6",
                device: "iPhone 13",
                os: "iOS 17.0",
                browser: "Safari",
                count: 12,
                remembered: false,
                creation: "2025-09-03T11:03:00Z",
                updated: "2025-11-18T20:01:00Z"
            }, {
                id: "s-7",
                device: "Windows 11 PC",
                os: "Windows 11",
                browser: "Chrome",
                count: 6,
                remembered: false,
                creation: "2025-07-01T06:45:00Z",
                updated: "2025-11-17T21:30:00Z"
            }, {
                id: "s-8",
                device: "Unknown System",
                os: "Unknown OS",
                browser: "Unknown",
                count: 1,
                remembered: false,
                creation: "2025-10-01T10:00:00Z",
                updated: "2025-11-18T19:00:00Z"
            }, {
                id: "s-9",
                device: "Unknown System",
                os: "Unknown",
                browser: "Unknown",
                count: 3,
                remembered: false,
                creation: "2025-10-07T14:30:00Z",
                updated: "2025-11-19T09:15:00Z"
            }, {
                id: "s-10",
                device: "Unknown System",
                os: "Unknown",
                browser: "Unknown",
                count: 3,
                remembered: false,
                creation: "2025-10-07T14:30:00Z",
                updated: "2025-11-19T09:15:00Z"
            }

            ]
        };
        // -------------------------------------------------------

        // simulate load
        setTimeout(() => {
            const mapped = dummy.activities.map(a => ({
                id: a.id,
                device: a.device,
                os: a.os,
                browser: a.browser,
                firstSeen: a.creation,
                lastSeen: a.updated,
                remembered: !!a.remembered,
                count: a.count,
                isCurrent: a.id === dummy.device_id,
                type: /iphone|android|mobile|ios/i.test(a.device + " " + a.os) ? "mobile" : /windows|mac|linux|desktop/i.test(a.device + " " + a.os) ? "desktop" : "unknown"

            }));

            // sort: current first then by lastSeen desc
            mapped.sort((x, y) => {
                if (x.isCurrent && !y.isCurrent) return -1;
                if (!x.isCurrent && y.isCurrent) return 1;
                return new Date(y.lastSeen) - new Date(x.lastSeen);
            });

            setSessions(mapped);
            setLoading(false);
        }, 450);
    }, []);

    const setBusy = (id, val) => setBusyIds(prev => {
        const copy = new Set(prev);
        if (val) copy.add(id); else copy.delete(id);
        return copy;
    });

    const revokeAll = () => {
        if (!confirm("Revoke all sessions except this one?")) return;
        // demo optimistic action
        setSessions(prev => prev.filter(s => s.isCurrent));
        alert("Revoke all (demo)");
    };

    const signOut = (id) => {
        if (!confirm("Sign out of this device?")) return;
        setBusy(id, true);
        setTimeout(() => {
            setSessions(prev => prev.filter(s => s.id !== id));
            setBusy(id, false);
            alert("Signed out (demo)");
        }, 550);
    };

    const current = sessions.find(s => s.isCurrent) || null;
    const others = sessions.filter(s => !s.isCurrent);

    return (
        <div className="p-5 box-border overflow-hidden">
            <div className="mx-auto max-w-4xl h-70vh px-4">
                <div
                    className="rounded-2xl p-6 md:p-8 flex flex-col overflow-hidden box-border mx-auto"
                    style={{
                        background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                        border: "1px solid rgba(255,255,255,0.02)"
                    }}
                >

                    <div className="flex items-center gap-10 mb-6 text-md font-medium p-3 rounded-lg border-2 border-gray-800 justify-center">
                        {[{label: "Login", href: "/login"}, {label: "Register", href: "/register"}, {
                            label: "MFA", href: "/mfa"
                        }, {label: "Change Password", href: "/change-password"}].map((item) => (
                            <a
                            key={item.href}
                            href={item.href}
                            className=" relative text-gray-300 hover:text-white transition after:absolute after:left-0 after:right-0 after:-bottom-1 after:h-[2px] after:bg-indigo-500 after:scale-x-0 hover:after:scale-x-100 after:transition-transform after:origin-left">
                            {item.label}
                        </a>))}
                    </div>


                    <div className="flex items-center justify-between mb-6">
                        <h1 className="text-lg md:text-xl font-semibold text-white">Your devices where you are signed in</h1>

                        <button
                            onClick={revokeAll}
                            className="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white rounded-md"
                        >
                            Revoke all
                        </button>
                    </div>


                    <div className="flex flex-col gap-6">

                        {/* Current device card */}
                        <div>
                            <div className="text-sm text-gray-400 mb-2">Current device</div>

                            {current ? (<div
                                className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800">
                                <div
                                    className="w-14 h-14 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex justify-center items-center">
                                    <DeviceIcon type={current.type}/>
                                </div>

                                <div className="flex-1">
                                    <div className="font-medium text-white">
                                        {current.device} <span
                                        className="text-xs text-gray-400">· {current.os}</span>
                                    </div>
                                    <div className="text-xs text-gray-400">Browser: {current.browser}</div>

                                    <div className="mt-2 text-xs text-gray-400">
                                        <div>First seen: {formatDate(current.firstSeen)}</div>
                                        <div>Last active: {formatDate(current.lastSeen)}</div>
                                    </div>
                                </div>

                                <div className="flex-none">
                                    <button
                                        onClick={() => signOut(current.id)}
                                        disabled={busyIds.has(current.id)}
                                        className="px-3 py-2 text-sm rounded-md bg-red-600 hover:bg-red-700 text-white disabled:opacity-60"
                                    >
                                        {busyIds.has(current.id) ? "Signing out…" : "Sign out"}
                                    </button>
                                </div>
                            </div>) : (<div className="text-sm text-gray-400">No current session found.</div>)}
                        </div>

                        {/* Other devices header */}
                        <div className="flex items-center justify-between">
                            <h2 className="text-lg font-semibold text-white">Other devices</h2>
                            <span className="text-sm text-gray-400">{others.length} active</span>
                        </div>

                        <div className=" other-scroll h-[45vh] overflow-y-auto pr-2">
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
                                {loading && (<div className="text-sm text-gray-400">Loading devices…</div>)}

                                {!loading && others.length === 0 && (
                                    <div className="text-sm text-gray-400">No other active devices.</div>)}

                                {!loading && others.map(s => (<div key={s.id}
                                                                   className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800">
                                    <div
                                        className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex justify-center items-center">
                                        <DeviceIcon type={s.type}/>
                                    </div>

                                    <div className="flex-1 min-w-0">
                                        <div className="text-sm text-white font-medium truncate">
                                            {s.device} <span className="text-xs text-gray-400">· {s.os}</span>
                                        </div>
                                        <div className="text-xs text-gray-400">Browser: {s.browser}</div>

                                        <div className="mt-2 text-xs text-gray-400">
                                            <div>First seen: {formatDate(s.firstSeen)}</div>
                                            <div>Last active: {formatDate(s.lastSeen)}</div>
                                        </div>
                                    </div>

                                    <div className="flex-none">
                                        <button
                                            onClick={() => signOut(s.id)}
                                            disabled={busyIds.has(s.id)}
                                            className="px-3 py-1 text-sm rounded-md bg-red-600 hover:bg-red-700 text-white disabled:opacity-60"
                                        >
                                            {busyIds.has(s.id) ? "Signing out…" : "Sign out"}
                                        </button>
                                    </div>
                                </div>))}
                            </div>
                        </div>


                    </div>

                </div>

            </div>
        </div>);
}
