import {useCallback, useEffect, useState} from "react";
import ConnectionManager from "../Contexts/Connection.tsx";
import {APIRoute} from "../Values/Constants";
import NotificationManager from "../Contexts/Notification.tsx";
import {Link} from "react-router";
import SessionDevice from "../Elements/SessionDevice.tsx";

type SingleUserDeviceUnprocessed = {
    id: string;
    count: number;
    remembered: boolean;
    created: string;
    updated: string;
    os: string;
    device: string;
    browser: string;
};

type UserDevicesResponse = {
    current: string; devices: SingleUserDeviceUnprocessed[];
};

const placeholderUser = {
    username: "Riya Bhariya",
    role: "Admin",
    email: "riya@example.com",
};

export type SingleUserDeviceProcessed = {
    id: string,
    device: string,
    os: string,
    browser: string,
    firstSeen: Date,
    lastSeen: Date,
    remembered: boolean,
    count: number,
    isCurrent: boolean,
    icon: string
}

export default function Sessions() {
    const {SendNotification} = NotificationManager();
    const {SendAPIRequest} = ConnectionManager()

    const [uiDisabled, setUiDisabled] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(false)
    const [currentDevice, setCurrentDevice] = useState<SingleUserDeviceProcessed | undefined>(undefined);
    const [otherDevices, setOtherDevices] = useState<SingleUserDeviceProcessed[]>([]);
    const [userDropdownOpen, setUserDropdownOpen] = useState<boolean>(false);

    const FetchDevices = useCallback(() => {
        setLoading(true);
        SendAPIRequest("GET", true, false, false, false, APIRoute, "/sessions/fetch")
            .then((data) => {
                if (data.success) {
                    const reply: UserDevicesResponse = data.reply as UserDevicesResponse
                    const mapped: SingleUserDeviceProcessed[] = reply.devices.map(device => ({
                        id: device.id,
                        device: device.device,
                        os: device.os,
                        browser: device.browser,
                        firstSeen: new Date(device.created),
                        lastSeen: new Date(device.updated),
                        remembered: device.remembered,
                        count: device.count,
                        isCurrent: device.id === reply.current,
                        icon: `/auth/device-icons/${device.device || "Unknown"}.svg`
                    }))

                    mapped.sort((a, b) => {
                        if (a.isCurrent !== b.isCurrent) {
                            return a.isCurrent ? -1 : 1;
                        }
                        return b.lastSeen.getTime() - a.lastSeen.getTime();
                    });
                    setCurrentDevice(mapped.find(s => s.isCurrent))
                    setOtherDevices(mapped.filter(s => !s.isCurrent));
                } else {
                    SendNotification("Failed to fetch devices")
                    setCurrentDevice(undefined)
                    setOtherDevices([])
                }
            })
            .catch((error) => {
                console.log("Devices fetch stopped because:", error)
                setCurrentDevice(undefined)
                setOtherDevices([])
            })
            .finally(() => {
                setLoading(false);
            })
    }, [SendNotification, SendAPIRequest])

    const RevokeDevice = useCallback((revokeAll: boolean, deviceID: string) => {
        if (currentDevice == undefined) return SendNotification("No devices visible. Refresh this page and retry.");

        setUiDisabled(true)
        const form = new FormData();
        form.append("all", revokeAll ? "yes" : "no")
        if (!revokeAll) form.append("device", deviceID)
        SendAPIRequest("POST", true, true, false, false, APIRoute, "/sessions/revoke", form)
            .then((data) => {
                if (data.success) {
                    if (revokeAll) SendNotification("All sessions have been revoked.")
                    else SendNotification("Session has been revoked.")

                    if (!revokeAll && deviceID !== currentDevice?.id) setOtherDevices((current) => current.filter((s) => s.id !== deviceID)); else FetchDevices()
                }
            })
            .catch((error) => {
                console.log("Device revoke stopped because:", error)
            })
            .finally(() => {
                setUiDisabled(false)
            })
    },[FetchDevices, SendNotification, SendAPIRequest, currentDevice])

    useEffect(() => {
        document.title = "Sessions - Bhariya";
        const timeoutId = window.setTimeout(() => {
            FetchDevices()
        }, 0);
        return () => window.clearTimeout(timeoutId);
    }, [FetchDevices]);

    return <div className="p-5 box-border overflow-hidden">
        <div className="mx-auto max-w-4xl h-70vh px-4">
            <div className="rounded-2xl p-6 md:p-8 flex flex-col overflow-hidden box-border mx-auto"
                 style={{
                     background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))", border: "1px solid rgba(255,255,255,0.02)"
                 }}>
                <div className="relative flex flex-wrap items-center gap-4 mb-6 text-md font-medium p-3 rounded-lg border-2 border-gray-800">
                    <div className="flex flex-1 flex-wrap items-center justify-center gap-6 md:gap-10">
                        {[
                            {label: "SignIn", href: "/signin"},
                            {label: "SignUp", href: "/signup"},
                            {label: "MFA", href: "/mfa"},
                            {label: "Change Password", href: "/passwordreset"}
                        ].map(item =>
                            <Link className="relative text-gray-300 hover:text-white transition after:absolute after:left-0 after:right-0 after:-bottom-1 after:h-0.5 after:bg-indigo-500 after:scale-x-0 hover:after:scale-x-100 after:transition-transform after:origin-left"
                                  to={item.href}
                                  key={item.label}
                                  state={{return_to:"/sessions"}}
                            >
                                {item.label}
                            </Link>)}
                    </div>
                    <div className="relative ml-auto">
                        <button
                            type="button"
                            className="flex items-center gap-2 rounded-md border border-gray-700 bg-gray-900/70 px-3 py-2 text-sm text-gray-200 transition hover:border-indigo-500 hover:text-white"
                            onClick={() => setUserDropdownOpen((current) => !current)}
                            aria-haspopup="true"
                            aria-expanded={userDropdownOpen}
                        >
                            <svg
                                className="h-5 w-5 text-indigo-300"
                                viewBox="0 0 24 24"
                                fill="none"
                                xmlns="http://www.w3.org/2000/svg"
                                aria-hidden="true"
                            >
                                <path
                                    d="M12 12C14.2091 12 16 10.2091 16 8C16 5.79086 14.2091 4 12 4C9.79086 4 8 5.79086 8 8C8 10.2091 9.79086 12 12 12Z"
                                    stroke="currentColor"
                                    strokeWidth="1.8"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                />
                                <path
                                    d="M4.5 20C5.24785 16.55 8.30743 14 12 14C15.6926 14 18.7522 16.55 19.5 20"
                                    stroke="currentColor"
                                    strokeWidth="1.8"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                />
                            </svg>
                            <span>{placeholderUser.username}</span>
                        </button>
                        {userDropdownOpen && <div className="absolute right-0 top-full z-20 mt-2 w-64 rounded-lg border border-gray-700 bg-gray-950 p-4 text-left shadow-xl shadow-black/40">
                            <div className="text-sm font-semibold text-white">
                                {placeholderUser.role}
                            </div>
                            <div className="mt-1 break-all text-sm font-normal text-gray-400">
                                {placeholderUser.email}
                            </div>
                        </div>}
                    </div>
                </div>

                <div className="flex items-center justify-between mb-6">
                    <h1 className="text-lg md:text-xl font-semibold text-white">
                        Devices signed in to your account
                    </h1>
                    <button className="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white rounded-md"
                            onClick={() => RevokeDevice(true, "")}
                            disabled={uiDisabled}>
                        Revoke all
                    </button>
                </div>
                <div className="flex flex-col gap-6">
                    <div>
                        {loading ? <div className="text-sm text-gray-400">
                            Loading current device
                        </div> : <>
                            {currentDevice ? <>
                                <div className="text-sm text-gray-400 mb-2">
                                    Current device
                                </div>
                                <SessionDevice device={currentDevice} disabled={uiDisabled} revoke={(deviceId:string)=>RevokeDevice(false, deviceId)} />
                            </> : <div className="text-sm text-gray-400">
                                No current session found.
                            </div>}
                        </>}
                    </div>

                    <div className="flex items-center justify-between">
                        <h2 className="text-lg font-semibold text-white">
                            Other devices
                        </h2>
                        <span className="text-sm text-gray-400">
                            {otherDevices.length} active
                        </span>
                    </div>

                    <div className="other-scroll h-[45vh] overflow-y-auto pr-2">
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
                            {loading && <div className="text-sm text-gray-400">
                                Loading devices
                            </div>}
                            {!loading && otherDevices.length === 0 && <div className="text-sm text-gray-400">
                                No other active devices.
                            </div>}
                            {
                                !loading && otherDevices.map(device =>
                                    <SessionDevice key={device.id} device={device} disabled={uiDisabled} revoke={(deviceId:string)=>RevokeDevice(false, deviceId)} />
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
}


