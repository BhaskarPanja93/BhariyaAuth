import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import ConnectionManager from "../Contexts/Connection.tsx";
import NotificationManager from "../Contexts/Notification.tsx";
import {APIRoute} from "../Values/Constants";

type RawLogLine = {
    c: string;
    l: number;
    i: string;
    f: string;
    t: string;
};

type LogRow = {
    timestamp: Date;
    timeRaw: string;
    level: number;
    file: string;
    id: string;
    content: string;
};

type ColumnKey = "time" | "file" | "level" | "id" | "content";
type SortDirection = "asc" | "desc";

type ActiveFilter = {
    id: string;
    key: ColumnKey;
    value: string;
};

const levelNames = ["INTENT", "INFO", "WARN", "ERROR", "BENCHMARK", "TEST", "BLOCKED"];

const levelStyles: Record<number, string> = {
    0: "bg-sky-900/70 text-sky-100",
    1: "bg-green-900/70 text-green-100",
    2: "bg-amber-800/70 text-amber-50",
    3: "bg-red-900/70 text-red-100",
    4: "bg-fuchsia-900/70 text-fuchsia-100",
    5: "bg-teal-900/70 text-teal-100",
    6: "bg-black text-gray-100",
};

const levelFilterOptions = ["INTENT", "INFO", "WARN", "ERROR", "BENCHMARK", "TEST", "BLOCKED", "UNKNOWN"] as const;

const columns: { key: ColumnKey; label: string }[] = [
    {key: "time", label: "TIME"},
    {key: "file", label: "FILE"},
    {key: "level", label: "LEVEL"},
    {key: "id", label: "ID"},
    {key: "content", label: "CONTENT"},
];

const tableColumns = "minmax(0,1.5fr) minmax(0,2fr) minmax(0,1.2fr) minmax(0,1fr) minmax(0,6fr)";
const dayMs = 24 * 60 * 60 * 1000;
const rowHeight = 36;
const overscanRows = 30;

const emptySearch: Record<ColumnKey, string> = {
    time: "",
    file: "",
    level: "",
    id: "",
    content: "",
};

function getLevelName(levelCode: number): string {
    return levelNames[levelCode] || "UNKNOWN";
}

function getRowCellValue(row: LogRow, key: ColumnKey): string {
    if (key === "time") return row.timeRaw;
    if (key === "level") return getLevelName(row.level);
    return row[key];
}

function getSortValue(row: LogRow, key: ColumnKey): string | number {
    if (key === "time") return row.timestamp.getTime();
    if (key === "level") return row.level;
    return getRowCellValue(row, key);
}

function extractDateParts(dayFileName: string): { year: number; month: number; day: number } | null {
    const match = dayFileName.match(/(\d{8})$/);
    if (!match) return null;

    const value = match[1];
    const year = Number(value.slice(0, 4));
    const month = Number(value.slice(4, 6));
    const day = Number(value.slice(6, 8));

    if (Number.isNaN(year) || Number.isNaN(month) || Number.isNaN(day)) return null;
    return {year, month, day};
}

function formatDayAge(dayFileName: string): string {
    const dateParts = extractDateParts(dayFileName);
    if (!dateParts) return "";

    const dayStart = Date.UTC(dateParts.year, dateParts.month - 1, dateParts.day);
    const now = new Date();
    const todayStart = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
    const diff = Math.floor((todayStart - dayStart) / dayMs);

    if (diff === 0) return "Today";
    return `${diff} day${diff === 1 ? "" : "s"} ago`;
}

function parseRows(dayFileName: string, rawText: string): LogRow[] {
    const dateParts = extractDateParts(dayFileName);
    if (!dateParts) return [];

    const parsedRows: LogRow[] = [];
    for (const rawLine of rawText.split("\n")) {
        if (rawLine.startsWith("#")) continue;
        const line = rawLine.trim();
        if (!line) continue;

        try {
            const parsed = JSON.parse(line) as RawLogLine;
            const timePieces = parsed.t.split(".");
            const hhmmss = timePieces[0] || "000000";
            const fraction = (timePieces[1] || "000");

            const timestamp = new Date(Date.UTC(
                dateParts.year,
                dateParts.month - 1,
                dateParts.day,
                Number(hhmmss.slice(0, 2)),
                Number(hhmmss.slice(2, 4)),
                Number(hhmmss.slice(4, 6)),
                Number(fraction),
            ));

            parsedRows.push({
                timestamp,
                timeRaw: parsed.t,
                level: parsed.l,
                file: parsed.f,
                id: parsed.i,
                content: parsed.c,
            });
        } catch {
            console.log(line)
        }
    }

    return parsedRows;
}

export default function LogsPage() {
    const {SendNotification} = NotificationManager();
    const {SendAPIRequest} = ConnectionManager();

    const listRef = useRef<HTMLDivElement | null>(null);
    const timeZoneBlurTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const timeZoneOptions = useMemo(() => {
        const local = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
        const intlWithSupportedValues = Intl as unknown as { supportedValuesOf?: (key: string) => string[] };
        const supported = intlWithSupportedValues.supportedValuesOf?.("timeZone") || [];
        return Array.from(new Set(["UTC", local, ...supported]));
    }, []);

    const [availableDays, setAvailableDays] = useState<string[]>([]);
    const [selectedDay, setSelectedDay] = useState<string>("");
    const [rows, setRows] = useState<LogRow[]>([]);
    const [daysLoading, setDaysLoading] = useState<boolean>(false);
    const [logsLoading, setLogsLoading] = useState<boolean>(false);
    const [timeZone, setTimeZone] = useState<string>(Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC");
    const [timeZoneSearch, setTimeZoneSearch] = useState<string>(Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC");
    const [timeZoneDropdownOpen, setTimeZoneDropdownOpen] = useState<boolean>(false);

    const [sortKey, setSortKey] = useState<ColumnKey>("time");
    const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
    const [searchByColumn, setSearchByColumn] = useState<Record<ColumnKey, string>>(emptySearch);
    const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

    const [viewportHeight, setViewportHeight] = useState<number>(0);
    const [scrollTop, setScrollTop] = useState<number>(0);
    const [scrollbarWidth, setScrollbarWidth] = useState<number>(0);

    const formatter = useMemo(() => new Intl.DateTimeFormat("en-US", {
        timeZone,
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        fractionalSecondDigits: 3,
        hour12: false,
        hourCycle: "h23",
    }), [timeZone]);

    const filteredTimeZoneOptions = useMemo(() => {
        const query = timeZoneSearch.trim().toLowerCase();
        if (!query) return timeZoneOptions;
        return timeZoneOptions.filter((zone) => zone.toLowerCase().includes(query));
    }, [timeZoneOptions, timeZoneSearch]);

    const filteredAndSortedRows = useMemo(() => {
        const filtered = rows.filter((row) => {
            for (const filter of activeFilters) {
                if (getRowCellValue(row, filter.key) !== filter.value) return false;
            }

            for (const column of columns) {
                const query = searchByColumn[column.key].trim().toLowerCase();
                if (!query) continue;

                const value = getRowCellValue(row, column.key).toLowerCase();
                if (column.key === "level") {
                    if (value !== query) return false;
                } else if (!value.includes(query)) return false;
            }

            return true;
        });

        return [...filtered].sort((left, right) => {
            const leftValue = getSortValue(left, sortKey);
            const rightValue = getSortValue(right, sortKey);

            let result;
            if (typeof leftValue === "number" && typeof rightValue === "number") {
                result = leftValue - rightValue;
            } else {
                result = String(leftValue).localeCompare(String(rightValue));
            }

            if (result === 0) result = left.timestamp.getTime() - right.timestamp.getTime();
            return sortDirection === "asc" ? result : -result;
        });
    }, [rows, activeFilters, searchByColumn, sortKey, sortDirection]);

    const totalRows = filteredAndSortedRows.length;

    const {startIndex, endIndex, topSpacerHeight, bottomSpacerHeight} = useMemo(() => {
        const visibleCount = Math.ceil((viewportHeight || 1) / rowHeight) + overscanRows * 2;
        const start = Math.max(0, Math.floor(scrollTop / rowHeight) - overscanRows);
        const end = Math.min(totalRows, start + visibleCount);

        return {
            startIndex: start,
            endIndex: end,
            topSpacerHeight: start * rowHeight,
            bottomSpacerHeight: Math.max(0, (totalRows - end) * rowHeight),
        };
    }, [viewportHeight, scrollTop, totalRows]);

    const visibleRows = useMemo(() => {
        return filteredAndSortedRows.slice(startIndex, endIndex);
    }, [filteredAndSortedRows, startIndex, endIndex]);

    const loadAvailableDays = useCallback(() => {
        setDaysLoading(true);
        SendAPIRequest("GET", true, false, false,false, APIRoute,  "/logs/available")
            .then((data) => {
                if (!data.success) {
                    setAvailableDays([]);
                    setSelectedDay("");
                    SendNotification("Could not load log files");
                    return;
                }

                const days = [...new Set((data.reply as string[] | undefined) || [])].sort((left, right) => right.localeCompare(left));
                setAvailableDays(days);
                setSelectedDay((currentDay) => {
                    if (currentDay && days.includes(currentDay)) return currentDay;
                    return days[0] || "";
                });
            })
            .catch((loadError) => {
                console.log("Log files fetch stopped because:", loadError);
                setAvailableDays([]);
                setSelectedDay("");
                SendNotification("Could not load log files.");
            })
            .finally(() => {
                setDaysLoading(false);
            });
    }, [SendNotification, SendAPIRequest]);

    const loadLogs = useCallback((dayFileName: string) => {
        if (!dayFileName) {
            setRows([]);
            return;
        }

        setLogsLoading(true);
        setScrollTop(0);
        if (listRef.current) listRef.current.scrollTop = 0;

        const safeDay = encodeURIComponent(dayFileName);
        SendAPIRequest("GET", true, false, false,false, APIRoute, `/logs/${safeDay}`)
            .then((data) => {
                if (!data.success) {
                    setRows([]);
                    SendNotification("Could not load logs for selected file.");
                    return;
                }

                const rawText = (data.reply as string | undefined) || "";
                setRows(parseRows(dayFileName, rawText));
            })
            .catch((loadError) => {
                console.log(`Logs fetch stopped for ${dayFileName} because:`, loadError);
                setRows([]);
                SendNotification("Could not load logs for selected file.");
            })
            .finally(() => {
                setLogsLoading(false);
            });
    }, [SendNotification, SendAPIRequest]);

    const addFilter = (key: ColumnKey, value: string) => {
        const id = `${key}:${value}`;
        setActiveFilters((previous) => {
            if (previous.some((filter) => filter.id === id)) return previous;
            return [...previous, {id, key, value}];
        });
    };

    const removeFilter = (id: string) => {
        setActiveFilters((previous) => previous.filter((filter) => filter.id !== id));
    };

    const clearFilters = () => {
        setActiveFilters([]);
    };

    const updateSearch = (key: ColumnKey, value: string) => {
        setSearchByColumn((previous) => ({...previous, [key]: value}));
    };

    const clearSearch = () => {
        setSearchByColumn(emptySearch);
    };

    const updateSort = (key: ColumnKey) => {
        if (sortKey === key) {
            setSortDirection((current) => current === "asc" ? "desc" : "asc");
            return;
        }

        setSortKey(key);
        setSortDirection(key === "time" ? "desc" : "asc");
    };

    const selectTimeZone = (zone: string) => {
        setTimeZone(zone);
        setTimeZoneSearch(zone);
        setTimeZoneDropdownOpen(false);
    };

    const commitTimeZoneSearch = () => {
        const normalizedQuery = timeZoneSearch.trim().toLowerCase();
        if (!normalizedQuery) {
            setTimeZoneSearch(timeZone);
            return;
        }

        const matchedZone = timeZoneOptions.find((zone) => zone.toLowerCase() === normalizedQuery);
        if (!matchedZone) {
            setTimeZoneSearch(timeZone);
            return;
        }

        setTimeZone(matchedZone);
        setTimeZoneSearch(matchedZone);
    };

    const handleTimeZoneFocus = () => {
        if (timeZoneBlurTimeoutRef.current) {
            clearTimeout(timeZoneBlurTimeoutRef.current);
            timeZoneBlurTimeoutRef.current = null;
        }
        setTimeZoneDropdownOpen(true);
    };

    const handleTimeZoneBlur = () => {
        if (timeZoneBlurTimeoutRef.current) clearTimeout(timeZoneBlurTimeoutRef.current);
        timeZoneBlurTimeoutRef.current = setTimeout(() => {
            setTimeZoneDropdownOpen(false);
            commitTimeZoneSearch();
        }, 120);
    };

    useEffect(() => {
        document.title = "Logs - Bhariya";
        const timeoutId = window.setTimeout(() => {
            loadAvailableDays();
        }, 0);
        return () => window.clearTimeout(timeoutId);
    }, [loadAvailableDays]);

    useEffect(() => {
        const timeoutId = window.setTimeout(() => {
            loadLogs(selectedDay);
        }, 0);
        return () => window.clearTimeout(timeoutId);
    }, [loadLogs, selectedDay]);

    useEffect(() => {
        const updateViewportMetrics = () => {
            const element = listRef.current;
            if (!element) {
                setViewportHeight(0);
                setScrollbarWidth(0);
                return;
            }
            setViewportHeight(element.clientHeight);
            setScrollbarWidth(Math.max(0, element.offsetWidth - element.clientWidth));
        };

        updateViewportMetrics();
        window.addEventListener("resize", updateViewportMetrics);
        return () => window.removeEventListener("resize", updateViewportMetrics);
    }, []);

    useEffect(() => {
        const element = listRef.current;
        if (!element) return;
        setScrollbarWidth(Math.max(0, element.offsetWidth - element.clientWidth));
    }, [totalRows, logsLoading]);

    useEffect(() => {
        if (!listRef.current) return;

        const maxScroll = Math.max(0, totalRows * rowHeight - listRef.current.clientHeight);
        if (listRef.current.scrollTop > maxScroll) {
            listRef.current.scrollTop = maxScroll;
            setScrollTop(maxScroll);
        }
    }, [totalRows]);

    useEffect(() => {
        return () => {
            if (timeZoneBlurTimeoutRef.current) clearTimeout(timeZoneBlurTimeoutRef.current);
        };
    }, []);

    return <div className="p-5 box-border overflow-hidden">
        <div className="mx-auto max-w-6xl h-70vh px-4">
            <div className="rounded-2xl p-6 md:p-8 flex flex-col overflow-hidden box-border mx-auto"
                 style={{
                     background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                     border: "1px solid rgba(255,255,255,0.02)",
                 }}>
                <div className="flex flex-wrap items-center gap-3 mb-3">
                    <div className="text-sm text-gray-300">Log file</div>
                    <select
                        value={selectedDay}
                        disabled={daysLoading || availableDays.length === 0}
                        onChange={(event) => setSelectedDay(event.target.value)}
                        className="rounded-md border border-gray-600 bg-gray-900 px-3 py-2 text-sm text-gray-100">
                        {availableDays.length === 0 && <option value="">No files</option>}
                        {availableDays.map((dayFileName) =>
                            <option key={dayFileName} value={dayFileName}>
                                {`${dayFileName} (${formatDayAge(dayFileName)})`}
                            </option>)}
                    </select>

                    <div className="text-sm text-gray-300 ml-2">Timezone</div>
                    <div className="relative min-w-65">
                        <input
                            value={timeZoneSearch}
                            onFocus={handleTimeZoneFocus}
                            onBlur={handleTimeZoneBlur}
                            onChange={(event) => {
                                setTimeZoneSearch(event.target.value);
                                setTimeZoneDropdownOpen(true);
                            }}
                            onKeyDown={(event) => {
                                if (event.key === "Escape") {
                                    event.preventDefault();
                                    setTimeZoneSearch(timeZone);
                                    setTimeZoneDropdownOpen(false);
                                    return;
                                }

                                if (event.key === "Enter") {
                                    event.preventDefault();
                                    if (filteredTimeZoneOptions.length > 0) {
                                        selectTimeZone(filteredTimeZoneOptions[0]);
                                    } else {
                                        commitTimeZoneSearch();
                                        setTimeZoneDropdownOpen(false);
                                    }
                                }
                            }}
                            placeholder="Search timezone"
                            role="combobox"
                            aria-autocomplete="list"
                            aria-expanded={timeZoneDropdownOpen}
                            className="w-full rounded-md border border-gray-600 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-gray-400"/>
                        {timeZoneDropdownOpen &&
                            <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded-md border border-gray-600 bg-gray-900 shadow-xl">
                                {filteredTimeZoneOptions.length === 0 &&
                                    <div className="px-3 py-2 text-xs text-gray-400">No matching timezones</div>}
                                {filteredTimeZoneOptions.map((zone) =>
                                    <button
                                        key={zone}
                                        type="button"
                                        onMouseDown={(event) => event.preventDefault()}
                                        onClick={() => selectTimeZone(zone)}
                                        className={`block w-full px-3 py-2 text-left text-sm ${zone === timeZone ? "bg-gray-700 text-white" : "text-gray-100 hover:bg-gray-800"}`}>
                                        {zone}
                                    </button>)}
                            </div>}
                    </div>
                </div>

                <div className="flex flex-wrap items-center gap-2 mb-3">
                    <div className="text-xs text-gray-400">Right-click any cell to add exact filter. Left-click drag to select text.</div>
                    {activeFilters.map((filter) =>
                        <div key={filter.id}
                             className="inline-flex items-center gap-2 rounded border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-200">
                            <span>{`${filter.key.toUpperCase()} = ${filter.value}`}</span>
                            <button
                                type="button"
                                onClick={() => removeFilter(filter.id)}
                                className="rounded bg-gray-700 px-1 text-[10px] hover:bg-gray-600">
                                x
                            </button>
                        </div>)}
                    {activeFilters.length > 0 &&
                        <button
                            type="button"
                            onClick={clearFilters}
                            className="rounded-md border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-200 hover:border-gray-500">
                            Clear filters
                        </button>}
                    {columns.some((column) => searchByColumn[column.key].trim().length > 0) &&
                        <button
                            type="button"
                            onClick={clearSearch}
                            className="rounded-md border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-200 hover:border-gray-500">
                            Clear search
                        </button>}
                </div>

                <div className="rounded-lg border border-gray-800 overflow-hidden">
                    <div className="grid text-xs font-semibold text-gray-300 bg-gray-900/70 border-b border-gray-700"
                         style={{gridTemplateColumns: tableColumns, width: `calc(100% - ${scrollbarWidth}px)`}}>
                        {columns.map((column) =>
                            <button
                                key={column.key}
                                type="button"
                                onClick={() => updateSort(column.key)}
                                className="p-2 border-r border-gray-700 last:border-r-0 text-left hover:bg-gray-800/70">
                                <span>{column.label}</span>
                                {sortKey === column.key &&
                                    <span className="ml-2 text-cyan-300">{sortDirection === "asc" ? "˄" : "˅"}</span>}
                            </button>)}
                    </div>

                    <div className="grid bg-[#111317] border-b border-gray-700"
                         style={{gridTemplateColumns: tableColumns, width: `calc(100% - ${scrollbarWidth}px)`}}>
                        {columns.map((column) =>
                            <div key={column.key} className="p-1 border-r border-gray-700 last:border-r-0">
                                {column.key === "level" ?
                                    <select
                                        value={searchByColumn.level}
                                        onChange={(event) => updateSearch("level", event.target.value)}
                                        className="w-full rounded border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-100 outline-none focus:border-gray-400">
                                        <option value="">All levels</option>
                                        {levelFilterOptions.map((level) =>
                                            <option key={level} value={level}>
                                                {level}
                                            </option>)}
                                    </select>
                                    :
                                    <input
                                        value={searchByColumn[column.key]}
                                        onChange={(event) => updateSearch(column.key, event.target.value)}
                                        placeholder={`Search ${column.label.toLowerCase()}`}
                                        className="w-full rounded border border-gray-600 bg-gray-900 px-2 py-1 text-xs text-gray-100 outline-none focus:border-gray-400"/>
                                }
                            </div>)}
                    </div>

                    <div
                        ref={listRef}
                        onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
                        className="other-scroll h-[62vh] overflow-y-auto text-sm">
                        {logsLoading && <div className="px-4 py-6 text-sm text-gray-400">Loading logs...</div>}
                        {!logsLoading && totalRows === 0 &&
                            <div className="px-4 py-6 text-sm text-gray-400">No logs match current filters/search.</div>}

                        {!logsLoading && totalRows > 0 &&
                            <div style={{height: `${totalRows * rowHeight}px`}}>
                                {topSpacerHeight > 0 && <div style={{height: `${topSpacerHeight}px`}}/>}
                                {visibleRows.map((log, index) => {
                                    const rowStyle = levelStyles[log.level] || "bg-gray-800 text-gray-100";
                                    const actualIndex = startIndex + index;

                                    return <div key={`${log.timeRaw}-${log.file}-${actualIndex}`}
                                                className={`grid border-b border-white/10 ${rowStyle}`}
                                                style={{gridTemplateColumns: tableColumns, minHeight: `${rowHeight}px`}}>
                                        {columns.map((column) => {
                                            const rawValue = getRowCellValue(log, column.key);
                                            const displayValue = column.key === "time" ? formatter.format(log.timestamp) : rawValue;
                                            return <div key={column.key}
                                                        title={`Filter ${column.label}: ${rawValue}`}
                                                        onContextMenu={(event) => {
                                                            event.preventDefault();
                                                            addFilter(column.key, rawValue);
                                                        }}
                                                        className={`p-2 border-r border-white/10 last:border-r-0 select-text cursor-text ${column.key === "time" ? "font-mono text-xs" : "text-xs"}`}>
                                                <div className={`${column.key === "time" ? "whitespace-pre-wrap wrap-break-word" : "whitespace-pre-wrap wrap-break-word"}`}>
                                                    {displayValue}
                                                </div>
                                            </div>;
                                        })}
                                    </div>;
                                })}
                                {bottomSpacerHeight > 0 && <div style={{height: `${bottomSpacerHeight}px`}}/>}
                            </div>}
                    </div>
                </div>
            </div>
        </div>
    </div>;
}


