import {type MouseEvent, useEffect, useMemo, useRef, useState} from "react";

const levelStyles = {
    intent: {style: "bg-sky-900 text-sky-100 hover:bg-sky-800", name: "intent"}, info: {style: "bg-green-900 text-green-100 hover:bg-green-800", name: "info"}, warn: {style: "bg-amber-800 text-amber-50 hover:bg-amber-700", name: "warn"}, error: {style: "bg-red-900 text-red-100 hover:bg-red-800", name: "error"}, benchmark: {style: "bg-fuchsia-900 text-fuchsia-100 hover:bg-fuchsia-800", name: "benchmark"}, test: {style: "bg-teal-900 text-teal-100 hover:bg-teal-800", name: "test"}, blocked: {style: "bg-black text-gray-100 hover:bg-gray-950", name: "blocked"}, unknown: {style: "bg-gray-800 text-gray-100 hover:bg-gray-700", name: "unknown"},
} as const;

type LogLevel = keyof typeof levelStyles;

type LogEntry = {
    c: string; l: number; i: string; f: string; t: string;
};

type LogRow = {
    dayUtc: string; timeRaw: string; file: string; level: LogLevel; id: string; content: string; timestamp: Date;
};

const levelByCode: Record<number, LogLevel> = {
    0: "intent", 1: "info", 2: "warn", 3: "error", 4: "benchmark", 5: "test", 6: "blocked",
};

const columnKeys = ["time", "file", "level", "id", "content"] as const;
type LogColumnKey = (typeof columnKeys)[number];
type SortDirection = "asc" | "desc";

type SortConfig = {
    key: LogColumnKey | null; dir: SortDirection;
};

type ContextMenuState = {
    x: number; y: number; value: string; key: LogColumnKey;
} | null;

type ExpandedCellState = {
    rowIndex: number; key: "content";
} | null;

type ActiveFilter = {
    id: string; key: LogColumnKey; value: string;
};

// TODO: fetch from api
const sampleLogsByDay: Record<string, string> = {
    "2026 05 07": `
{"c":"Credentials refreshing","l":0,"i":"","f":"processors/mail/main","t":"060512.325"}
{"c":"Credentials refreshed","l":1,"i":"","f":"processors/mail/main","t":"060512.326"}
{"c":"Server startup","l":1,"i":"","f":"main","t":"060512.322"}
{"c":"Received request from 114.29.226.213 for path /auth/api/signin/step1/password","l":1,"i":"TOdj_p","f":"middleware/profiling","t":"063011.034"}
{"c":"Requested account: bhaskarpanja93@gmail.com password","l":0,"i":"TOdj_p","f":"routers/signin/step1","t":"063011.034"}
{"c":"Completed request","l":1,"i":"TOdj_p","f":"routers/signin/step1","t":"063011.039"}
{"c":"Received request from 114.29.226.213 for path /auth/api/signin/step2","l":1,"i":"2IPkTp","f":"middleware/profiling","t":"063013.264"}
{"c":"Requested account: -2147483645 password","l":0,"i":"2IPkTp","f":"routers/signin/step2","t":"063013.264"}
{"c":"Completed request: -32766","l":1,"i":"2IPkTp","f":"routers/signin/step2","t":"063014.702"}
{"c":"Completed request","l":1,"i":"CrE0SR","f":"routers/sso/step1","t":"064854.1479"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/callback/google","l":1,"i":"wAKIZa","f":"middleware/profiling","t":"064857.324"}
{"c":"Requested: google","l":0,"i":"wAKIZa","f":"routers/sso/step2","t":"064857.324"}
{"c":"Session decrypt failed: decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Sessiondecrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session.decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session.decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session","l":3,"i":"wAKIZa","f":"routers/sso/step2","t":"064857.322"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/google","l":1,"i":"R_mZfb","f":"middleware/profiling","t":"065449.463"}
{"c":"Requested: google","l":0,"i":"R_mZfb","f":"routers/sso/step1","t":"065449.467"}
{"c":"Completed request","l":1,"i":"R_mZfb","f":"routers/sso/step1","t":"065449.467"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/callback/google","l":1,"i":"MMQX~m","f":"middleware/profiling","t":"065453.847"}
{"c":"Requested: google","l":0,"i":"MMQX~m","f":"routers/sso/step2","t":"065453.844"}
{"c":"Session decrypt failed: decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session","l":3,"i":"MMQX~m","f":"routers/sso/step2","t":"065453.848"}
`, "2026 05 09": `
{"c":"Credentials refreshing","l":0,"i":"","f":"processors/mail/main","t":"071236.038"}
{"c":"Credentials refreshed","l":1,"i":"","f":"processors/mail/main","t":"071236.038"}
{"c":"Unix socket path received: /PROJECTS/BHARIYA-AUTH/live/back.sock","l":1,"i":"","f":"main","t":"071236.068"}
{"c":"Attempting run on unix socket","l":0,"i":"","f":"main","t":"071236.068"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/google","l":1,"i":"TvZ3.K","f":"middleware/profiling","t":"071252.845"}
{"c":"Requested: google","l":0,"i":"TvZ3.K","f":"routers/sso/step1","t":"071252.845"}
{"c":"Completed request","l":1,"i":"TvZ3.K","f":"routers/sso/step1","t":"071252.845"}
{"c":"Received request from 114.29.226.213 for path /auth/api/sso/callback/google","l":1,"i":"anocMu","f":"middleware/profiling","t":"071300.535"}
{"c":"Requested: google","l":0,"i":"anocMu","f":"routers/sso/step2","t":"071300.535"}
{"c":"Session decrypt failed: decrypt interface from string: decrypt interface from bytes - Unmarshal: json: cannot unmarshal goth.Session into Go value of type goth.Session","l":3,"i":"anocMu","f":"routers/sso/step2","t":"071300.535"}
`,
};

const tableColumns = "minmax(0,2.4fr) minmax(0,3fr) minmax(0,1.5fr) minmax(0,1fr) minmax(0,6fr)";

const fallbackTimeZones = ["UTC", "Africa/Cairo", "Africa/Johannesburg", "America/Anchorage", "America/Chicago", "America/Denver", "America/Los_Angeles", "America/New_York", "America/Phoenix", "America/Sao_Paulo", "Asia/Bangkok", "Asia/Dubai", "Asia/Hong_Kong", "Asia/Kolkata", "Asia/Singapore", "Asia/Tokyo", "Australia/Adelaide", "Australia/Melbourne", "Australia/Perth", "Australia/Sydney", "Europe/Amsterdam", "Europe/Berlin", "Europe/London", "Europe/Moscow", "Europe/Paris", "Pacific/Auckland", "Pacific/Honolulu",];

const dayPattern = /^(\d{4}) (\d{2}) (\d{2})$/;
const timePattern = /^(\d{2})(\d{2})(\d{2})(?:\.(\d{1,4}))?$/;

function normalizeDayKey(dayKey: string): string | null {
    const match = dayKey.trim().match(/(\d{4} \d{2} \d{2})$/);
    return match ? match[1] : null;
}

function parseDateTime(dayUtc: string, timeStr: string): Date | null {
    const dayMatch = dayUtc.match(dayPattern);
    const timeMatch = timeStr.match(timePattern);
    if (!dayMatch || !timeMatch) return null;

    const year = Number(dayMatch[1]);
    const month = Number(dayMatch[2]) - 1;
    const day = Number(dayMatch[3]);

    const hour = Number(timeMatch[1]);
    const minute = Number(timeMatch[2]);
    const second = Number(timeMatch[3]);
    const fraction = timeMatch[4] ?? "0";
    const millisecond = Number(fraction.padEnd(3, "0").slice(0, 3));

    if (Number.isNaN(year) || Number.isNaN(month) || Number.isNaN(day) || Number.isNaN(hour) || Number.isNaN(minute) || Number.isNaN(second) || Number.isNaN(millisecond)) {
        return null;
    }

    return new Date(Date.UTC(year, month, day, hour, minute, second, millisecond));
}

function createFormatter(timeZone: string): Intl.DateTimeFormat {
    return new Intl.DateTimeFormat("en-US", {
        timeZone, hour: "2-digit", minute: "2-digit", second: "2-digit", fractionalSecondDigits: 3, hour12: false, hourCycle: "h23",
    });
}

function formatTime(timestamp: Date, formatter: Intl.DateTimeFormat): string {
    const parts = formatter.formatToParts(timestamp);
    const partValues: Record<string, string> = {};
    for (const part of parts) {
        partValues[part.type] = part.value;
    }

    const hour = partValues.hour ?? "00";
    const minute = partValues.minute ?? "00";
    const second = partValues.second ?? "00";
    const fractionalSecond = partValues.fractionalSecond ?? "000";

    return `${hour}:${minute}:${second}.${fractionalSecond}`;
}

function getLevelName(levelCode: number): LogLevel {
    return levelByCode[levelCode] ?? "unknown";
}

function parseLogText(logsByDay: Record<string, string>): LogRow[] {
    const rows: LogRow[] = [];

    for (const [dayKey, rawText] of Object.entries(logsByDay)) {
        const dayUtc = normalizeDayKey(dayKey);
        if (!dayUtc) continue;

        for (const rawLine of rawText.split("\n")) {
            const line = rawLine.trim();
            if (line === "" || line.startsWith("#")) continue;

            try {
                const parsed = JSON.parse(line) as LogEntry;
                const timestamp = parseDateTime(dayUtc, parsed.t);
                if (!timestamp) continue;

                rows.push({
                    dayUtc, timeRaw: parsed.t, file: parsed.f, level: getLevelName(parsed.l), id: parsed.i, content: parsed.c, timestamp,
                });
            } catch {
                // Ignore malformed lines from mixed/raw logs.
            }
        }
    }

    return rows;
}

function getCellRawValue(log: LogRow, key: LogColumnKey): string {
    if (key === "level") return log.level;
    if (key === "time") return log.timeRaw;
    return log[key];
}

function getComparableValue(log: LogRow, key: LogColumnKey): string | number {
    if (key === "time") return log.timestamp.getTime();
    return getCellRawValue(log, key);
}

function getAvailableTimeZones(): string[] {
    const intlWithSupportedValues = Intl as unknown as {
        supportedValuesOf?: (key: string) => string[];
    };

    const candidates = intlWithSupportedValues.supportedValuesOf?.("timeZone") ?? fallbackTimeZones;
    const unique = Array.from(new Set(["UTC", ...candidates]));
    return unique.sort((left, right) => left.localeCompare(right));
}

function getAvailableDays(rows: LogRow[]): string[] {
    return Array.from(new Set(rows.map((row) => row.dayUtc))).sort((left, right) => right.localeCompare(left));
}

function applyFilters(rows: LogRow[], filters: ActiveFilter[]): LogRow[] {
    if (filters.length === 0) return rows;
    return rows.filter((row) => filters.every((filter) => getCellRawValue(row, filter.key) === filter.value));
}

function applySort(rows: LogRow[], sortConfig: SortConfig): LogRow[] {
    if (!sortConfig.key) return rows;

    return [...rows].sort((leftLog, rightLog) => {
        const leftValue = getComparableValue(leftLog, sortConfig.key!);
        const rightValue = getComparableValue(rightLog, sortConfig.key!);

        if (leftValue === rightValue) return 0;
        if (leftValue > rightValue) return sortConfig.dir === "asc" ? 1 : -1;
        return sortConfig.dir === "asc" ? -1 : 1;
    });
}

function formatFilterLabel(filter: ActiveFilter): string {
    return `${filter.key.toUpperCase()} = ${filter.value}`;
}

export default function LogsPage() {
    const allLogs = useMemo(() => parseLogText(sampleLogsByDay), []);
    const allTimeZones = useMemo(() => getAvailableTimeZones(), []);
    const availableDays = useMemo(() => getAvailableDays(allLogs), [allLogs]);

    const [sortConfig, setSortConfig] = useState<SortConfig>({key: null, dir: "asc"});
    const [contextMenu, setContextMenu] = useState<ContextMenuState>(null);
    const [expandedCell, setExpandedCell] = useState<ExpandedCellState>(null);
    const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
    const [selectedDay, setSelectedDay] = useState<string>(availableDays[0] ?? "");
    const [timeZone, setTimeZone] = useState<string>("UTC");
    const [timeZoneDropdownOpen, setTimeZoneDropdownOpen] = useState<boolean>(false);
    const [timeZoneSearch, setTimeZoneSearch] = useState<string>("");
    const [dateDropdownOpen, setDateDropdownOpen] = useState<boolean>(false);
    const timeZoneDropdownRef = useRef<HTMLDivElement | null>(null);
    const dateDropdownRef = useRef<HTMLDivElement | null>(null);

    const formatter = useMemo(() => createFormatter(timeZone), [timeZone]);

    useEffect(() => {
        if (!contextMenu) return;
        const closeMenu = (): void => setContextMenu(null);
        window.addEventListener("click", closeMenu);
        return () => window.removeEventListener("click", closeMenu);
    }, [contextMenu]);

    useEffect(() => {
        if (!timeZoneDropdownOpen) return;

        const closeOnOutsideClick = (event: globalThis.MouseEvent): void => {
            if (!timeZoneDropdownRef.current) return;
            if (!timeZoneDropdownRef.current.contains(event.target as Node)) {
                setTimeZoneDropdownOpen(false);
            }
        };

        window.addEventListener("mousedown", closeOnOutsideClick);
        return () => window.removeEventListener("mousedown", closeOnOutsideClick);
    }, [timeZoneDropdownOpen]);

    useEffect(() => {
        if (!dateDropdownOpen) return;

        const closeOnOutsideClick = (event: globalThis.MouseEvent): void => {
            if (!dateDropdownRef.current) return;
            if (!dateDropdownRef.current.contains(event.target as Node)) {
                setDateDropdownOpen(false);
            }
        };

        window.addEventListener("mousedown", closeOnOutsideClick);
        return () => window.removeEventListener("mousedown", closeOnOutsideClick);
    }, [dateDropdownOpen]);

    const filteredTimeZones = useMemo(() => {
        const query = timeZoneSearch.trim().toLowerCase();
        if (query.length === 0) return [];

        return allTimeZones
            .filter((zone) => zone.toLowerCase().includes(query))
            .slice(0, 100);
    }, [allTimeZones, timeZoneSearch]);

    const logs = useMemo(() => {
        const dayScopedLogs = selectedDay === "" ? [] : allLogs.filter((log) => log.dayUtc === selectedDay);
        const filtered = applyFilters(dayScopedLogs, activeFilters);
        return applySort(filtered, sortConfig);
    }, [allLogs, activeFilters, selectedDay, sortConfig]);

    const sortBy = (key: LogColumnKey): void => {
        const dir: SortDirection = sortConfig.key === key && sortConfig.dir === "asc" ? "desc" : "asc";
        setSortConfig({key, dir});
    };

    const handleRightClick = (event: MouseEvent<HTMLDivElement>, value: string, key: LogColumnKey): void => {
        event.preventDefault();
        setContextMenu({
            x: event.clientX, y: event.clientY, value, key,
        });
    };

    const copyData = (): void => {
        if (!contextMenu) return;
        navigator.clipboard.writeText(contextMenu.value).then();
        setContextMenu(null);
    };

    const filterData = (): void => {
        if (!contextMenu) return;

        const id = `${contextMenu.key}|${contextMenu.value}`;
        setActiveFilters((previousFilters) => {
            if (previousFilters.some((filter) => filter.id === id)) return previousFilters;
            return [...previousFilters, {
                id, key: contextMenu.key, value: contextMenu.value,
            },];
        });
        setContextMenu(null);
    };

    const removeFilter = (id: string): void => {
        setActiveFilters((previousFilters) => previousFilters.filter((filter) => filter.id !== id));
    };

    const toggleExpandedCell = (rowIndex: number): void => {
        const nextCell: ExpandedCellState = expandedCell?.rowIndex === rowIndex ? null : {rowIndex, key: "content"};
        setExpandedCell(nextCell);
    };

    const contentNeedsExpand = (value: string): boolean => value.length > 120;

    const getDisplayValue = (log: LogRow, key: LogColumnKey): string => {
        if (key === "time") {
            return formatTime(log.timestamp, formatter);
        }
        if (key === "level") {
            return log.level.toUpperCase();
        }
        return getCellRawValue(log, key);
    };

    return (<div className="p-5 box-border overflow-hidden min-h-screen bg-linear-to-br from-gray-700 via-[#1a1c20] to-[#0b0d10]">
            <div className="mx-auto max-w-6xl h-[85vh] px-4">
                <div
                    className="rounded-2xl p-6 md:p-8 flex flex-col overflow-hidden box-border mx-auto h-full"
                    style={{
                        background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))", border: "1px solid rgba(255,255,255,0.02)",
                    }}
                >
                    <div className="flex-1 flex flex-col overflow-hidden rounded-xl border border-gray-800">
                        <div className="border-b border-gray-700/70 bg-gray-900/60 p-3 flex items-start gap-3">
                            <div ref={timeZoneDropdownRef} className="relative min-w-64">
                                <button
                                    type="button"
                                    onClick={() => {
                                        setTimeZoneDropdownOpen((open) => !open);
                                        setTimeZoneSearch("");
                                    }}
                                    className="w-full text-left rounded-md border border-gray-600 bg-gray-900 px-3 py-2 text-sm text-gray-100 hover:border-gray-500"
                                >
                                    {timeZone}
                                </button>
                                {timeZoneDropdownOpen && (<div className="absolute mt-2 w-full rounded-md border border-gray-600 bg-[#111317] p-2 shadow-xl z-40">
                                        <input
                                            value={timeZoneSearch}
                                            onChange={(event) => setTimeZoneSearch(event.target.value)}
                                            placeholder="Type timezone..."
                                            className="w-full rounded border border-gray-600 bg-gray-900 px-2 py-1 text-sm text-gray-100 outline-none focus:border-gray-400"
                                        />
                                        {timeZoneSearch.trim().length === 0 ? (<div className="mt-2 text-xs text-gray-400 px-1 py-1">
                                                Type at least one letter to view timezones.
                                            </div>) : (<div className="mt-2 max-h-56 overflow-y-auto rounded border border-gray-700">
                                                {filteredTimeZones.length === 0 ? (<div className="px-2 py-2 text-xs text-gray-400">No timezone found.</div>) : (filteredTimeZones.map((zone) => (<button
                                                            key={zone}
                                                            type="button"
                                                            onClick={() => {
                                                                setTimeZone(zone);
                                                                setTimeZoneDropdownOpen(false);
                                                                setTimeZoneSearch("");
                                                            }}
                                                            className={`block w-full px-2 py-1 text-left text-sm hover:bg-gray-700 ${zone === timeZone ? "bg-gray-800 text-white" : "text-gray-200"}`}
                                                        >
                                                            {zone}
                                                        </button>)))}
                                            </div>)}
                                    </div>)}
                            </div>

                            <div className="min-w-0 flex-1">
                                <div className="flex flex-wrap items-center gap-2">
                                    {activeFilters.map((filter) => (<span
                                            key={filter.id}
                                            className="inline-flex items-center gap-2 rounded-md border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-200"
                                        >
                                                {formatFilterLabel(filter)}
                                            <button
                                                type="button"
                                                onClick={() => removeFilter(filter.id)}
                                                className="rounded bg-gray-700 px-1 text-[10px] leading-4 hover:bg-gray-600"
                                            >
                                                    x
                                                </button>
                                            </span>))}
                                    {activeFilters.length > 0 && (<button
                                            type="button"
                                            onClick={() => setActiveFilters([])}
                                            className="rounded-md border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-200 hover:border-gray-500"
                                        >
                                            Clear filters
                                        </button>)}
                                </div>
                            </div>

                            <div ref={dateDropdownRef} className="relative ml-auto shrink-0">
                                <button
                                    type="button"
                                    onClick={() => setDateDropdownOpen((open) => !open)}
                                    className="rounded-md border border-gray-600 bg-gray-900 px-3 py-2 text-sm text-gray-100 hover:border-gray-500"
                                >
                                    {selectedDay || "None"}
                                </button>
                                {dateDropdownOpen && (<div className="absolute right-0 mt-2 w-44 rounded-md border border-gray-600 bg-[#111317] p-2 shadow-xl z-40">
                                        <div className="max-h-56 overflow-y-auto rounded border border-gray-700">
                                            {availableDays.map((day) => (<button
                                                    key={day}
                                                    type="button"
                                                    onClick={() => {
                                                        setSelectedDay(day);
                                                        setDateDropdownOpen(false);
                                                    }}
                                                    className={`block w-full px-2 py-1 text-left text-sm hover:bg-gray-700 ${day === selectedDay ? "bg-gray-800 text-white" : "text-gray-200"}`}
                                                >
                                                    {day}
                                                </button>))}
                                        </div>
                                    </div>)}
                            </div>
                        </div>

                        <div className="sticky top-0 z-20 bg-gray-800 border-b border-gray-600 shadow-md">
                            <div className="grid text-xs font-semibold text-gray-300" style={{gridTemplateColumns: tableColumns}}>
                                {columnKeys.map((key) => (<div
                                        key={key}
                                        onClick={() => sortBy(key)}
                                        className="min-w-0 overflow-hidden border-r border-gray-700/80 p-3 cursor-pointer hover:bg-gray-800 transition last:border-r-0"
                                    >
                                        {key.toUpperCase()}
                                    </div>))}
                            </div>
                        </div>

                        <div className="flex-1 overflow-y-auto text-sm">
                            {logs.length === 0 && (<div className="px-4 py-6 text-sm text-gray-400">No logs match current filters.</div>)}
                            {logs.map((log, idx) => {
                                const styleEntry = levelStyles[log.level] ?? levelStyles.unknown;
                                return (<div
                                        key={`${log.dayUtc}-${log.timeRaw}-${log.file}-${idx}`}
                                        className={`grid w-full border-b ${styleEntry.style}`}
                                        style={{gridTemplateColumns: tableColumns}}
                                    >
                                        {columnKeys.map((key) => {
                                            const isContentCell = key === "content";
                                            const rawValue = getCellRawValue(log, key);
                                            const displayValue = getDisplayValue(log, key);
                                            const expandable = isContentCell && contentNeedsExpand(log.content);
                                            const isExpanded = expandedCell?.rowIndex === idx && expandedCell.key === "content";

                                            return (<div
                                                    key={key}
                                                    onClick={() => {
                                                        if (expandable) toggleExpandedCell(idx);
                                                    }}
                                                    onContextMenu={(event) => handleRightClick(event, rawValue, key)}
                                                    className={`min-w-0 overflow-hidden border-r p-1 transition last:border-r-0 ${styleEntry.style} ${key === "time" ? "font-mono text-xs tracking-normal" : ""} border-white/10`}
                                                >
                                                    <div
                                                        className={isContentCell ? isExpanded ? "max-w-full whitespace-pre-wrap break-words text-xs leading-4" : "max-w-full overflow-hidden break-words text-xs leading-4" : "max-w-full overflow-hidden text-ellipsis whitespace-nowrap"}
                                                    >
                                                        {displayValue}
                                                    </div>
                                                </div>);
                                        })}
                                    </div>);
                            })}
                        </div>
                    </div>
                </div>
            </div>

            {contextMenu && (<div
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
                </div>)}
        </div>);
}
