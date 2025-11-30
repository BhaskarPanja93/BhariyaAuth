import React, {useEffect, useRef, useState} from "react";
import {DeviceIcons} from "../Elements/DeviceIcons"
import {FetchConnectionManager} from "../Contexts/Connection.jsx";
import {BackendURL} from "../Values/Constants.js";

export default function Sessions() {
    const {privateAPI} = FetchConnectionManager()

    const [loading, setLoading] = useState(false)
    const userID = useRef("");
    const currentDevice = useRef(null);
    const otherDevices = useRef([]);
    const [sessions, setSessions] = useState([]);

    function formatDate(dt) {
        try {
            return new Date(dt).toLocaleString();
        } catch {
            return dt;
        }
    }

    const FetchDevices = () => {
        setLoading(true);
        privateAPI.post(BackendURL + "/sessions/fetch")
            .then((data) => {
                if (data["success"]) {
                    userID.current = data["reply"]["user_id"]
                    if (!data["reply"]["activities"]) {
                        currentDevice.current = null
                        otherDevices.current = []
                        return setSessions(null)
                    }
                    const mapped = data["reply"]["activities"].map(a => ({
                        id: a["id"], device: a["device"], os: a["os"], browser: a["browser"], firstSeen: a["creation"], lastSeen: a["updated"], remembered: a["remembered"], count: a["count"], isCurrent: a["id"] === data["reply"]["device_id"], type: /iphone|android|mobile|ios/i.test(a["device"] + " " + a["os"]) ? "mobile" : /windows|mac|linux|desktop/i.test(a["device"] + " " + a["os"]) ? "desktop" : "unknown"
                    }));
                    mapped.sort((x, y) => {
                        if (x.isCurrent && !y.isCurrent) return -1;
                        if (!x.isCurrent && y.isCurrent) return 1;
                        return new Date(y.lastSeen) - new Date(x.lastSeen);
                    });
                    currentDevice.current = mapped.find(s => s.isCurrent) || null;
                    otherDevices.current = mapped.filter(s => !s.isCurrent);
                    setSessions(mapped);
                }
            })
            .catch(_ => {
            })
            .finally(_ => {
                setLoading(false);
            })
    }

    const RevokeDevice = (revokeAll, deviceID) => {
        const form = new FormData();
        form.append("user_id", userID.current)
        form.append("revoke_all", revokeAll ? "yes" : "no")
        !revokeAll && form.append("device_id", deviceID)
        privateAPI.post(BackendURL + "/sessions/revoke", form)
            .then((data) => {
                if (data["success"]) {
                    FetchDevices()
                }
            })
            .catch(_ => {
            })
            .finally(_ => {
            })
    }

    useEffect(() => {
        FetchDevices()
    }, []);

    return <div className="p-5 box-border overflow-hidden">
        <div className="mx-auto max-w-4xl h-70vh px-4">
            <div
                className="rounded-2xl p-6 md:p-8 flex flex-col overflow-hidden box-border mx-auto"
                style={{
                    background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))", border: "1px solid rgba(255,255,255,0.02)"
                }}
            >
                <div className="flex items-center gap-10 mb-6 text-md font-medium p-3 rounded-lg border-2 border-gray-800 justify-center">
                    {[{label: "Login", href: "/login"}, {label: "Register", href: "/register"}, {label: "MFA", href: "/mfa"}, {label: "Change Password", href: "/change-password"}].map(item => <a key={item.href} href={item.href} className=" relative text-gray-300 hover:text-white transition after:absolute after:left-0 after:right-0 after:-bottom-1 after:h-[2px] after:bg-indigo-500 after:scale-x-0 hover:after:scale-x-100 after:transition-transform after:origin-left">
                        {item.label}
                    </a>)}
                </div>
                <div className="flex items-center justify-between mb-6">
                    <h1 className="text-lg md:text-xl font-semibold text-white">Your devices where you are signed in</h1>

                    <button onClick={() => RevokeDevice(true)} className="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white rounded-md">
                        Revoke all
                    </button>
                </div>
                <div className="flex flex-col gap-6">
                    <div>
                        <div className="text-sm text-gray-400 mb-2">Current device</div>
                        {currentDevice.current ? <div
                            className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800">
                            <div
                                className="w-14 h-14 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex justify-center items-center">
                                <DeviceIcons type={currentDevice.current.type}/>
                            </div>

                            <div className="flex-1">
                                <div className="font-medium text-white">
                                    {currentDevice.current.device}
                                    <span className="text-xs text-gray-400">
                                            · {currentDevice.current.os}
                                        </span>
                                </div>
                                <div className="text-xs text-gray-400">
                                    Browser: {currentDevice.current.browser}
                                </div>
                                <div className="mt-2 text-xs text-gray-400">
                                    <div>First seen: {formatDate(currentDevice.current.firstSeen)}</div>
                                    <div>Last active: {formatDate(currentDevice.current.lastSeen)}</div>
                                </div>
                            </div>

                            <div className="flex-none">
                                <button onClick={() => RevokeDevice(false, currentDevice.current.id)} className="px-3 py-2 text-sm rounded-md bg-red-600 hover:bg-red-700 text-white disabled:opacity-60">
                                    Sign out
                                </button>
                            </div>
                        </div> : <div className="text-sm text-gray-400">No current session found.</div>}
                    </div>

                    <div className="flex items-center justify-between">
                        <h2 className="text-lg font-semibold text-white">Other devices</h2>
                        <span className="text-sm text-gray-400">{otherDevices.current.length} active</span>
                    </div>

                    <div className=" other-scroll h-[45vh] overflow-y-auto pr-2">
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
                            {loading && (<div className="text-sm text-gray-400">Loading devices…</div>)}
                            {!loading && otherDevices.current.length === 0 && (<div className="text-sm text-gray-400">No other active devices.</div>)}
                            {!loading && otherDevices.current.map(s => <div key={s.id} className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800">
                                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 flex justify-center items-center">
                                    <DeviceIcons type={s.type}/>
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="text-sm text-white font-medium truncate">
                                        {s.device}
                                        <span className="text-xs text-gray-400">
                                            · {s.os}
                                        </span>
                                    </div>
                                    <div className="text-xs text-gray-400">
                                        Browser: {s.browser}
                                    </div>
                                    <div className="mt-2 text-xs text-gray-400">
                                        <div>First seen: {formatDate(s.firstSeen)}</div>
                                        <div>Last active: {formatDate(s.lastSeen)}</div>
                                    </div>
                                </div>

                                <div className="flex-none">
                                    <button onClick={() => RevokeDevice(false, s.id)} className="px-3 py-1 text-sm rounded-md bg-red-600 hover:bg-red-700 text-white disabled:opacity-60">
                                        Sign out
                                    </button>
                                </div>
                            </div>)}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
}
