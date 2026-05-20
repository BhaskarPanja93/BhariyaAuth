import type {SingleUserDeviceProcessed} from "../Structures/Sessions";

function formatDate(datetime: Date): string {
    try {
        return new Date(datetime).toLocaleString();
    } catch {
        return "Unknown time";
    }
}

export default function SessionDevice(
    {device, disabled, revoke}:
        {
            device: SingleUserDeviceProcessed;
            disabled: boolean;
            revoke: (deviceId: string) => void;
        },
) {
    return <div key={device.id} className="flex items-center gap-4 p-4 rounded-lg bg-[#0b0f14] border border-gray-800">
        <div className="w-14 h-14 rounded-full bg-linear-to-br from-gray-800 to-gray-900 flex justify-center items-center">
            <img
                alt={device.device}
                src={device.icon}
                onError={(event) => {
                    if (event.currentTarget.dataset.fallbackApplied) return;
                    event.currentTarget.dataset.fallbackApplied = "true";
                    event.currentTarget.src = "/auth/device-icons/Unknown.svg";
                }}/>
        </div>

        <div className="min-w-0 flex-1">
            <div className="font-medium flex items-center gap-2">
                <div className="text-white">
                    {device.device}
                    <span className="text-xs text-gray-400">&nbsp;-&nbsp;{device.os}</span>
                </div>
                {device.remembered &&
                    <div className="px-2 py-0.5 text-xs font-semibold text-red-300 bg-red-500/10 border border-red-500/30 rounded-md">
                        Session Saved
                    </div>}
            </div>

            <div className="text-xs text-gray-400">
                Browser: {device.browser}
            </div>

            <div className="mt-2 text-xs text-gray-400">
                <div>First seen: {formatDate(device.firstSeen)}</div>
                <div>Last active: {formatDate(device.lastSeen)}</div>
            </div>
        </div>

        <div className="flex-none">
            <button
                className="px-3 py-2 text-sm rounded-md bg-red-600 hover:bg-red-700 text-white disabled:opacity-60"
                onClick={() => revoke(device.id)}
                disabled={disabled}>
                Sign out
            </button>
        </div>
    </div>;
}


