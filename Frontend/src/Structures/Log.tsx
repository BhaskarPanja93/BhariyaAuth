import {type MouseEvent, useMemo, useState} from "react";
import {Link} from "react-router";

const levelStyles = {
    intent: {style: "bg-sky-900 text-sky-100 hover:bg-sky-800", name: "intent"},
    info: {style: "bg-green-900 text-green-100 hover:bg-green-800", name: "info"},
    warn: {style: "bg-amber-800 text-amber-50 hover:bg-amber-700", name: "warn"},
    error: {style: "bg-red-900 text-red-100 hover:bg-red-800", name: "error"},
    benchmark: {style: "bg-fuchsia-900 text-fuchsia-100 hover:bg-fuchsia-800", name: "benchmark"},
    test: {style: "bg-teal-900 text-teal-100 hover:bg-teal-800", name: "test"},
    blocked: {style: "bg-black text-gray-100 hover:bg-gray-950", name: "blocked"},
    unknown: {style: "bg-gray-800 text-gray-100 hover:bg-gray-700", name: "unknown"},
} as const;

type LogLevel = keyof typeof levelStyles;

type LogEntry = {
    c: string;
    l: number;
    i: string;
    f: string;
    t: string;
};

type LogRow = {
    time: string;
    file: string;
    level: LogLevel;
    id: string;
    content: string;
    timestamp: Date;
};

const levelByCode: Record<number, LogLevel> = {
    0: "intent",
    1: "info",
    2: "warn",
    3: "error",
    4: "benchmark",
    5: "test",
    6: "blocked",
};

const navItems = [
    {label: "SignIn", href: "/signin"},
    {label: "SignUp", href: "/signup"},
    {label: "MFA", href: "/mfa"},
    {label: "Change Password", href: "/passwordreset"},
] as const;

const columnKeys = ["time", "file", "level", "id", "content"] as const;
type LogColumnKey = (typeof columnKeys)[number];
type SortDirection = "asc" | "desc";

type SortConfig = {
    key: LogColumnKey | null;
    dir: SortDirection;
};

type ContextMenuState = {
    x: number;
    y: number;
    value: string;
    key: LogColumnKey;
} | null;

type ExpandedCellState = {
    rowIndex: number;
    key: "content";
} | null;

// TODO: fetch from api
const sampleLog = `
# ----------------
# LOGGER START
# ----------------
{"c":"Credentials refreshing","l":0,"i":"","f":"processors/mail/main","t":"060512.3259"}
{"c":"Credentials refreshed","l":1,"i":"","f":"processors/mail/main","t":"060512.3261"}
{"c":"Server startup","l":1,"i":"","f":"main","t":"060512.3262"}
{"c":"Received request from 114.29.226.213 for path /auth/api/signin/step1/password","l":1,"i":"TOdj_p","f":"middleware/profiling","t":"063011.0345"}
{"c":"Requested account: bhaskarpanja93@gmail.com password","l":0,"i":"TOdj_p","f":"routers/signin/step1","t":"063011.0346"}
{"c":"Completed request","l":1,"i":"TOdj_p","f":"routers/signin/step1","t":"063011.0393"}
{"c":"Received request from 114.29.226.213 for path /auth/api/signin/step2","l":1,"i":"2IPkTp","f":"middleware/profiling","t":"063013.2640"}
{"c":"Requested account: -2147483645 password","l":0,"i":"2IPkTp","f":"routers/signin/step2","t":"063013.2643"}
{"c":"Completed request: -32766","l":1,"i":"2IPkTp","f":"routers/signin/step2","t":"063014.7025"}
{"c":"Completed request","l":1,"i":"CrE0SR","f":"routers/sso/step1","t":"064854.1479"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/callback/google","l":1,"i":"wAKIZa","f":"middleware/profiling","t":"064857.3241"}
{"c":"Requested: google","l":0,"i":"wAKIZa","f":"routers/sso/step2","t":"064857.3242"}
{"c":"Session decrypt failed: decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session","l":3,"i":"wAKIZa","f":"routers/sso/step2","t":"064857.3242"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/google","l":1,"i":"R_mZfb","f":"middleware/profiling","t":"065449.4673"}
{"c":"Requested: google","l":0,"i":"R_mZfb","f":"routers/sso/step1","t":"065449.4673"}
{"c":"Completed request","l":1,"i":"R_mZfb","f":"routers/sso/step1","t":"065449.4674"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/callback/google","l":1,"i":"MMQX~m","f":"middleware/profiling","t":"065453.8447"}
{"c":"Requested: google","l":0,"i":"MMQX~m","f":"routers/sso/step2","t":"065453.8448"}
{"c":"Session decrypt failed: decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session","l":3,"i":"MMQX~m","f":"routers/sso/step2","t":"065453.8448"}
# ----------------
# LOGGER START
# ----------------
{"c":"Credentials refreshing","l":0,"i":"","f":"processors/mail/main","t":"071236.0383"}
{"c":"Credentials refreshed","l":1,"i":"","f":"processors/mail/main","t":"071236.0385"}
{"c":"Unix socket path received: /PROJECTS/BHARIYA-AUTH/live/back.sock","l":1,"i":"","f":"main","t":"071236.0682"}
{"c":"Attempting run on unix socket","l":0,"i":"","f":"main","t":"071236.0682"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/google","l":1,"i":"TvZ3.K","f":"middleware/profiling","t":"071252.8451"}
{"c":"Requested: google","l":0,"i":"TvZ3.K","f":"routers/sso/step1","t":"071252.8451"}
{"c":"Completed request","l":1,"i":"TvZ3.K","f":"routers/sso/step1","t":"071252.8453"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/callback/google","l":1,"i":"anocMu","f":"middleware/profiling","t":"071300.5352"}
{"c":"Requested: google","l":0,"i":"anocMu","f":"routers/sso/step2","t":"071300.5352"}
{"c":"Session decrypt failed: decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session","l":3,"i":"anocMu","f":"routers/sso/step2","t":"071300.5354"}
# ----------------
# LOGGER START
# ----------------
{"c":"Credentials refreshing","l":0,"i":"","f":"processors/mail/main","t":"071434.5897"}
{"c":"Credentials refreshed","l":1,"i":"","f":"processors/mail/main","t":"071434.5899"}
{"c":"Server startup","l":1,"i":"","f":"main","t":"071434.5902"}
{"c":"Connecting SQL","l":0,"i":"","f":"stores/sql","t":"071434.5902"}
{"c":"SQL using unix socket","l":0,"i":"","f":"stores/sql","t":"071434.5902"}
{"c":"SQL Connected","l":1,"i":"","f":"stores/sql","t":"071434.5905"}
{"c":"SQL Connected and Pinged","l":1,"i":"","f":"stores/sql","t":"071434.5982"}
{"c":"Connecting Redis","l":0,"i":"","f":"stores/sql","t":"071434.5982"}
{"c":"Redis using unix socket","l":0,"i":"","f":"stores/redis","t":"071434.5982"}
{"c":"Redis Connected and Pinged","l":1,"i":"","f":"stores/redis","t":"071434.5993"}
{"c":"Initializing goth providers","l":0,"i":"","f":"routers/sso/main","t":"071434.5993"}
{"c":"Attaching Profiling Middleware","l":0,"i":"","f":"main","t":"071434.5994"}
{"c":"Attaching Routers","l":0,"i":"","f":"main","t":"071434.5995"}

`;

const tableColumns = "minmax(0,2fr) minmax(0,3fr) minmax(0,1.5fr) minmax(0,1fr) minmax(0,6fr)";

function parseDateTime(timeStr: string, baseDate: Date): Date {
    const [hhmmss, ms = "000"] = timeStr.split(".");
    const hour = Number(hhmmss.slice(0, 2));
    const minute = Number(hhmmss.slice(2, 4));
    const second = Number(hhmmss.slice(4, 6));
    const millisecond = Number(ms.padEnd(3, "0").slice(0, 3));
    return new Date(Date.UTC(
        baseDate.getUTCFullYear(),
        baseDate.getUTCMonth(),
        baseDate.getUTCDate(),
        hour,
        minute,
        second,
        millisecond
    ));
}

function createFormatter(timeZone: string): Intl.DateTimeFormat {
    return new Intl.DateTimeFormat("en-US", {
        timeZone,
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        fractionalSecondDigits: 3,
        hour12: false,
    });
}

function getLevelName(levelCode: number): LogLevel {
    return levelByCode[levelCode] ?? "unknown";
}

function parseLogText(rawText: string): LogRow[] {
    const baseDate = new Date();
    const rows: LogRow[] = [];

    for (const rawLine of rawText.split("\n")) {
        const line = rawLine.trim();
        if (line === "" || line.startsWith("#")) continue;

        try {
            const parsed = JSON.parse(line) as LogEntry;
            rows.push({
                time: parsed.t,
                file: parsed.f,
                level: getLevelName(parsed.l),
                id: parsed.i,
                content: parsed.c,
                timestamp: parseDateTime(parsed.t, baseDate),
            });
        } catch {
            // Ignore malformed lines from mixed/raw logs.
        }
    }

    return rows;
}

function getCellRawValue(log: LogRow, key: LogColumnKey): string {
    if (key === "level") return log.level;
    return log[key];
}

function getComparableValue(log: LogRow, key: LogColumnKey): string | number {
    if (key === "time") return log.timestamp.getTime();
    return getCellRawValue(log, key);
}

export default function LogsPage() {
    const allLogs = useMemo(() => parseLogText(sampleLog), []);
    const [logs, setLogs] = useState<LogRow[]>(allLogs);
    const [sortConfig, setSortConfig] = useState<SortConfig>({key: null, dir: "asc"});
    const [contextMenu, setContextMenu] = useState<ContextMenuState>(null);
    const [expandedCell, setExpandedCell] = useState<ExpandedCellState>(null);
    const [showLocalTime, setShowLocalTime] = useState<boolean>(false);

    const timeZone = useMemo(() => Intl.DateTimeFormat().resolvedOptions().timeZone ?? "UTC", []);
    const formatter = useMemo(() => createFormatter(timeZone), [timeZone]);

    const sortBy = (key: LogColumnKey): void => {
        const dir: SortDirection = sortConfig.key === key && sortConfig.dir === "asc" ? "desc" : "asc";
        const sorted = [...logs].sort((leftLog, rightLog) => {
            const leftValue = getComparableValue(leftLog, key);
            const rightValue = getComparableValue(rightLog, key);

            if (leftValue === rightValue) return 0;
            if (leftValue > rightValue) return dir === "asc" ? 1 : -1;
            return dir === "asc" ? -1 : 1;
        });

        setLogs(sorted);
        setSortConfig({key, dir});
    };

    const handleRightClick = (event: MouseEvent<HTMLDivElement>, value: string, key: LogColumnKey): void => {
        event.preventDefault();
        setContextMenu({
            x: event.clientX,
            y: event.clientY,
            value,
            key,
        });
    };

    const copyData = (): void => {
        if (!contextMenu) return;
        navigator.clipboard.writeText(contextMenu.value).then();
        setContextMenu(null);
    };

    const filterData = (): void => {
        if (!contextMenu) return;
        const filtered = allLogs.filter((log) => getCellRawValue(log, contextMenu.key) === contextMenu.value);
        setLogs(filtered);
        setContextMenu(null);
    };

    const toggleExpandedCell = (rowIndex: number): void => {
        const nextCell: ExpandedCellState = expandedCell?.rowIndex === rowIndex ? null : {rowIndex, key: "content"};
        setExpandedCell(nextCell);
    };

    const contentNeedsExpand = (value: string): boolean => value.length > 120;

    const getDisplayValue = (log: LogRow, key: LogColumnKey): string => {
        if (key === "time") {
            return showLocalTime ? formatter.format(log.timestamp) : log.time;
        }
        if (key === "level") {
            return log.level.toUpperCase();
        }
        return getCellRawValue(log, key);
    };

    return (
        <div className="p-5 box-border overflow-hidden min-h-screen bg-linear-to-br from-gray-700 via-[#1a1c20] to-[#0b0d10]">
            <div className="mx-auto max-w-6xl h-[85vh] px-4">
                <div
                    className="rounded-2xl p-6 md:p-8 flex flex-col overflow-hidden box-border mx-auto h-full"
                    style={{
                        background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                        border: "1px solid rgba(255,255,255,0.02)",
                    }}
                >
                    <div className="flex flex-wrap items-center gap-6 md:gap-10 mb-6 text-md font-medium p-3 rounded-lg border-2 border-gray-800 justify-center">
                        {navItems.map((item) => (
                            <Link
                                className="relative text-gray-300 hover:text-white transition after:absolute after:left-0 after:right-0 after:-bottom-1 after:h-0.5 after:bg-indigo-500 after:scale-x-0 hover:after:scale-x-100 after:transition-transform after:origin-left"
                                to={item.href}
                                key={item.href}
                            >
                                {item.label}
                            </Link>
                        ))}
                    </div>

                    <div className="flex-1 flex flex-col overflow-hidden rounded-xl border border-gray-800">
                        <div className="sticky top-0 z-20 bg-gray-800 border-b border-gray-600 shadow-md">
                            <div className="grid text-xs font-semibold text-gray-300" style={{gridTemplateColumns: tableColumns}}>
                                {columnKeys.map((key) => (
                                    <div
                                        key={key}
                                        onClick={() => sortBy(key)}
                                        className="min-w-0 overflow-hidden border-r border-gray-700/80 p-3 cursor-pointer hover:bg-gray-800 transition last:border-r-0"
                                    >
                                        {key === "time" ? (
                                            <div className="flex items-center justify-between gap-2">
                                                <span>{key.toUpperCase()}</span>
                                                <label
                                                    className="flex items-center gap-1 text-[10px] font-medium text-gray-300"
                                                    onClick={(event) => event.stopPropagation()}
                                                >
                                                    <input
                                                        type="checkbox"
                                                        checked={showLocalTime}
                                                        onChange={(event) => setShowLocalTime(event.target.checked)}
                                                        className="h-3 w-3 rounded border-gray-500 bg-gray-900"
                                                    />
                                                    <span>Local</span>
                                                </label>
                                            </div>
                                        ) : key.toUpperCase()}
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div className="flex-1 overflow-y-auto text-sm">
                            {logs.map((log, idx) => {
                                const styleEntry = levelStyles[log.level] ?? levelStyles.unknown;
                                return (
                                    <div
                                        key={`${log.time}-${log.file}-${idx}`}
                                        className={`grid w-full border-b ${styleEntry.style}`}
                                        style={{gridTemplateColumns: tableColumns}}
                                    >
                                        {columnKeys.map((key) => {
                                            const isContentCell = key === "content";
                                            const rawValue = getCellRawValue(log, key);
                                            const displayValue = getDisplayValue(log, key);
                                            const expandable = isContentCell && contentNeedsExpand(log.content);
                                            const isExpanded = expandedCell?.rowIndex === idx && expandedCell.key === "content";

                                            return (
                                                <div
                                                    key={key}
                                                    onClick={() => {
                                                        if (expandable) toggleExpandedCell(idx);
                                                    }}
                                                    onContextMenu={(event) => handleRightClick(event, rawValue, key)}
                                                    className={`min-w-0 overflow-hidden border-r p-1 transition last:border-r-0 ${styleEntry.style} ${key === "time" ? "font-mono text-sm tracking-wider" : ""} ${expandable ? "cursor-pointer" : ""} border-white/10`}
                                                >
                                                    <div
                                                        className={isContentCell
                                                            ? isExpanded
                                                                ? "max-w-full whitespace-pre-wrap break-words text-xs leading-4"
                                                                : "line-clamp-2 max-w-full overflow-hidden break-words text-xs leading-4"
                                                            : "max-w-full overflow-hidden text-ellipsis whitespace-nowrap"}
                                                    >
                                                        {displayValue}
                                                    </div>
                                                </div>
                                            );
                                        })}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                </div>
            </div>

            {contextMenu && (
                <div
                    style={{top: contextMenu.y, left: contextMenu.x}}
                    className="fixed bg-[#1a1c20] border border-gray-700 rounded-md shadow-lg text-sm z-50"
                >
                    <div
                        onClick={copyData}
                        className="px-4 py-2 hover:bg-gray-700 cursor-pointer text-gray-300"
                    >
                        Copy data
                    </div>
                    <div
                        onClick={filterData}
                        className="px-4 py-2 hover:bg-gray-700 cursor-pointer text-gray-300"
                    >
                        Filter data
                    </div>
                    <div
                        onClick={() => setContextMenu(null)}
                        className="px-4 py-2 hover:bg-gray-700 cursor-pointer text-gray-300"
                    >
                        Close
                    </div>
                </div>
            )}
        </div>
    );
}
