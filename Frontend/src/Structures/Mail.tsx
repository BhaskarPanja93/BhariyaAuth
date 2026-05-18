import {type KeyboardEvent, useEffect, useRef, useState} from "react";
import {Link} from "react-router";
import ConnectionManager from "../Contexts/Connection.tsx";
import NotificationManager from "../Contexts/Notification.tsx";
import {EmailIsValid} from "../Utils/Strings.ts";
import {APIRoute, FrontendRoute} from "../Values/Constants.ts";

type AudienceMode = "individuals" | "groups" | "everyone";

const GROUP_OPTIONS = [
    {id: "Unknown", label: "Unknown"},
    {id: "Viewer", label: "Viewer"},
    {id: "Moderator", label: "Moderator"},
    {id: "Admin", label: "Admin"},
    {id: "Owner", label: "Owner"},
];

const FONT_SIZES = ["13", "15", "17", "19", "22", "26", "30"];

const STARTER_BODY = `
<p style="margin: 0 0 16px; font-size: 15px; color: #374151; line-height: 1.5;">
    Your account is ready. You can now sign in and start with our services.
</p>
<p style="margin: 0; font-size: 15px; color: #374151; line-height: 1.5;">
    Add your update here and use the toolbar to format the message.
</p>`;

function escapeHtml(value: string) {
    return value
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;");
}

function escapeAttribute(value: string) {
    return escapeHtml(value).replace(/'/g, "&#39;");
}

function normalizeUrl(raw: string) {
    const trimmed = raw.trim();

    try {
        const parsed = new URL(trimmed);
        if (["http:", "https:", "mailto:", "tel:"].includes(parsed.protocol)) {
            return parsed.toString();
        }
    } catch {
        return "";
    }

    return "";
}

function sanitizeStyle(element: HTMLElement) {
    const allowed = ["color", "background-color", "font-size", "font-weight", "font-style", "text-decoration", "text-align", "line-height", "margin", "padding", "border-radius", "max-width", "width", "height", "display"];
    const kept: string[] = [];

    allowed.forEach((property) => {
        const value = element.style.getPropertyValue(property);
        if (!value || /url\(|expression\(|javascript:/i.test(value)) return;
        kept.push(`${property}: ${value}`);
    });

    if (kept.length > 0) {
        element.setAttribute("style", kept.join("; "));
    } else {
        element.removeAttribute("style");
    }
}

function sanitizeComposerHtml(html: string) {
    const parser = new DOMParser();
    const documentBody = parser.parseFromString(`<div>${html}</div>`, "text/html").body;
    const root = documentBody.firstElementChild as HTMLElement | null;
    if (!root) return "";

    const allowedTags = new Set(["A", "B", "BLOCKQUOTE", "BR", "DIV", "EM", "FONT", "H2", "H3", "HR", "I", "IMG", "LI", "OL", "P", "S", "SPAN", "STRONG", "U", "UL"]);

    const cleanNode = (node: Node) => {
        if (node.nodeType !== Node.ELEMENT_NODE) return;

        const element = node as HTMLElement;
        const tagName = element.tagName;

        Array.from(element.childNodes).forEach(cleanNode);

        if (!allowedTags.has(tagName)) {
            element.replaceWith(...Array.from(element.childNodes));
            return;
        }

        Array.from(element.attributes).forEach((attribute) => {
            const name = attribute.name.toLowerCase();
            if (name === "style") return;
            if (tagName === "A" && ["href", "target", "rel"].includes(name)) return;
            if (tagName === "IMG" && ["src", "alt", "width", "height"].includes(name)) return;
            if (tagName === "FONT" && ["color", "size"].includes(name)) return;
            element.removeAttribute(attribute.name);
        });

        if (tagName === "A") {
            const href = normalizeUrl(element.getAttribute("href") || "");
            if (href) {
                element.setAttribute("href", href);
                element.setAttribute("target", "_blank");
                element.setAttribute("rel", "noreferrer");
            } else {
                element.removeAttribute("href");
            }
        }

        if (tagName === "IMG") {
            const src = normalizeUrl(element.getAttribute("src") || "");
            if (!src || !/^https?:/i.test(src)) {
                element.remove();
                return;
            }
            element.setAttribute("src", src);
            element.setAttribute("alt", element.getAttribute("alt") || "Email image");
            element.setAttribute("style", "display: block; max-width: 100%; height: auto; border-radius: 8px; margin: 16px 0;");
        } else {
            sanitizeStyle(element);
        }
    };

    Array.from(root.childNodes).forEach(cleanNode);
    return root.innerHTML.trim();
}

function buildEmailHtml(subject: string, content: string) {
    const safeSubject = escapeHtml(subject.trim() || "Bhariya update");
    const cleanContent = sanitizeComposerHtml(content) || STARTER_BODY;

    return `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>${safeSubject}</title>
</head>
<body style="margin:0; padding:0; background-color: #eef0f3; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;">
<table cellpadding="0" cellspacing="0" width="100%">
    <tr>
        <td align="center" style="padding: 48px 16px;">
            <table cellpadding="0" cellspacing="0" style="max-width: 520px; background-color: #ffffff; border-radius: 14px; border: 1px solid #e5e7eb; overflow: hidden;" width="100%">
                <tr>
                    <td align="center" style="background: linear-gradient(135deg, #1f2937, #0b0d10); padding: 28px; border-bottom: 1px solid #1f2937;">
                        <table cellpadding="0" cellspacing="0" style="border: 1px solid rgba(255,255,255,0.12); border-radius: 10px;" width="100%">
                            <tr>
                                <td align="center" style="padding: 20px;">
                                    <img src="${FrontendRoute}/favicons/DarkMode.png" style="display:block;" width="120" alt="Bhariya"/>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 28px;">
                        <table cellpadding="0" cellspacing="0" style="background-color: #ffffff; border: 1px solid #e5e7eb; border-radius: 10px;" width="100%">
                            <tr>
                                <td style="padding: 28px;">
                                    <div style="font-size: 15px; color: #374151; line-height: 1.5;">
                                        ${cleanContent}
                                    </div>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>`;
}

export default function Mail() {
    const {SendNotification} = NotificationManager();
    const {SendPost} = ConnectionManager();

    const editorRef = useRef<HTMLDivElement>(null);
    const savedRangeRef = useRef<Range | null>(null);

    const [uiDisabled, setUiDisabled] = useState(false);
    const [subject, setSubject] = useState("New Login Detected");
    const [audience, setAudience] = useState<AudienceMode>("individuals");
    const [recipientInput, setRecipientInput] = useState("");
    const [recipients, setRecipients] = useState<string[]>([]);
    const [selectedGroups, setSelectedGroups] = useState<string[]>(["Admin"]);
    const [bodyHtml, setBodyHtml] = useState(STARTER_BODY);
    const [previewHtml, setPreviewHtml] = useState<string | null>(null);
    const [textColor, setTextColor] = useState("#374151");
    const [highlightColor, setHighlightColor] = useState("#fef3c7");
    const [fontSize, setFontSize] = useState("15");

    const syncBody = () => {
        setBodyHtml(editorRef.current?.innerHTML || "");
    };

    const saveSelection = () => {
        const editor = editorRef.current;
        const selection = window.getSelection();
        if (!editor || !selection || selection.rangeCount === 0) return;
        const range = selection.getRangeAt(0);
        if (!editor.contains(range.commonAncestorContainer)) return;
        savedRangeRef.current = range.cloneRange();
    };

    const setEditorHtml = (html: string) => {
        if (editorRef.current) editorRef.current.innerHTML = html;
        setBodyHtml(html);
    };

    const focusEditor = () => {
        editorRef.current?.focus();
        const selection = window.getSelection();
        if (selection && savedRangeRef.current) {
            selection.removeAllRanges();
            selection.addRange(savedRangeRef.current);
        }
    };

    const runCommand = (command: string, value?: string) => {
        focusEditor();
        document.execCommand(command, false, value);
        syncBody();
    };

    const insertHtml = (html: string) => {
        focusEditor();
        document.execCommand("insertHTML", false, html);
        syncBody();
    };

    const applyInlineStyle = (property: string, value: string) => {
        focusEditor();
        const selection = window.getSelection();
        if (!selection || selection.rangeCount === 0 || selection.isCollapsed) {
            if (property === "font-size") {
                SendNotification("Select text before applying font size");
                return;
            }
            document.execCommand(property === "color" ? "foreColor" : "backColor", false, value);
            syncBody();
            return;
        }

        const range = selection.getRangeAt(0);
        const span = document.createElement("span");
        span.style.setProperty(property, value);
        span.appendChild(range.extractContents());
        range.insertNode(span);
        selection.removeAllRanges();

        const nextRange = document.createRange();
        nextRange.selectNodeContents(span);
        nextRange.collapse(false);
        selection.addRange(nextRange);
        syncBody();
    };

    const parseRecipientInput = () => {
        const next = recipientInput
            .split(/[\s,;]+/)
            .map((email) => email.trim().toLowerCase())
            .filter(Boolean);

        const invalid = next.filter((email) => !EmailIsValid(email));
        if (invalid.length > 0) {
            SendNotification(`Invalid email: ${invalid[0]}`);
            return null;
        }

        return next;
    };

    const addRecipients = () => {
        const next = parseRecipientInput();
        if (!next) return;
        setRecipients((current) => Array.from(new Set([...current, ...next])));
        setRecipientInput("");
    };

    const handleRecipientKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === "Enter" || event.key === "," || event.key === ";") {
            event.preventDefault();
            addRecipients();
        }
    };

    const toggleGroup = (group: string) => {
        setSelectedGroups((current) => current.includes(group) ? current.filter((item) => item !== group) : [...current, group]);
    };

    const insertLink = () => {
        const url = normalizeUrl(window.prompt("Enter link URL") || "");
        if (!url) {
            SendNotification("Link must start with http, https, mailto, or tel");
            return;
        }
        runCommand("createLink", url);
    };

    const insertImageUrl = () => {
        const url = normalizeUrl(window.prompt("Enter image URL") || "");
        if (!url || !/^https?:/i.test(url)) {
            SendNotification("Image URL must start with http or https");
            return;
        }
        insertHtml(`<img src="${escapeAttribute(url)}" alt="Email image" />`);
    };

    const openPreview = () => {
        const currentBodyHtml = editorRef.current?.innerHTML || bodyHtml;
        setBodyHtml(currentBodyHtml);
        setPreviewHtml(buildEmailHtml(subject, currentBodyHtml));
    };

    const validateBeforeSend = (nextRecipients: string[], currentBodyHtml: string) => {
        if (!subject.trim()) {
            SendNotification("Subject is required");
            return false;
        }
        if (audience === "individuals" && nextRecipients.length === 0) {
            SendNotification("Add at least one recipient");
            return false;
        }
        if (audience === "groups" && selectedGroups.length === 0) {
            SendNotification("Choose at least one group");
            return false;
        }
        if (!editorRef.current?.innerText.trim() && !currentBodyHtml.includes("<img")) {
            SendNotification("Email body is required");
            return false;
        }
        return true;
    };

    const sendEmail = () => {
        let nextRecipients = audience === "individuals" ? recipients : [];
        if (audience === "individuals" && recipientInput.trim()) {
            const pendingRecipients = parseRecipientInput();
            if (!pendingRecipients) return;
            nextRecipients = Array.from(new Set([...recipients, ...pendingRecipients]));
            setRecipients(nextRecipients);
            setRecipientInput("");
        }
        const currentBodyHtml = editorRef.current?.innerHTML || bodyHtml;
        if (!validateBeforeSend(nextRecipients, currentBodyHtml)) return;
        setBodyHtml(currentBodyHtml);

        setUiDisabled(true);
        const form = new FormData();
        form.append("subject", subject.trim());
        form.append("audience", audience);
        form.append("recipients", JSON.stringify(nextRecipients));
        form.append("groups", JSON.stringify(audience === "groups" ? selectedGroups : []));
        form.append("body", editorRef.current?.innerText || "");
        form.append("html", buildEmailHtml(subject, currentBodyHtml));

        SendPost(true, true, false, APIRoute, "/mail/send", form)
            .then((data) => {
                if (data.success) {
                    SendNotification("Email queued for sending");
                }
            })
            .catch((error) => {
                console.log("Mail send stopped because:", error);
            })
            .finally(() => setUiDisabled(false));
    };

    useEffect(() => {
        document.title = "Mail - Bhariya";
        setEditorHtml(STARTER_BODY);
    }, []);

    return (
        <div className="min-h-screen p-3 sm:p-5 box-border overflow-x-hidden">
            <div className="mx-auto w-full max-w-5xl">
                <div className="rounded-2xl p-4 sm:p-5 md:p-6 flex flex-col box-border min-h-[calc(100vh-24px)] sm:min-h-[calc(100vh-40px)]"
                     style={{
                         background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                         border: "1px solid rgba(255,255,255,0.02)"
                     }}>
                    <div className="flex flex-wrap items-center gap-3 sm:gap-6 md:gap-10 mb-5 text-sm md:text-base font-medium p-3 rounded-lg border-2 border-gray-800 justify-center">
                        {[
                            {label: "Sessions", href: "/sessions"},
                            {label: "SignIn", href: "/signin"},
                            {label: "SignUp", href: "/signup"},
                            {label: "MFA", href: "/mfa"},
                            {label: "Change Password", href: "/passwordreset"}
                        ].map((item) =>
                            <Link to={item.href} key={item.label} className="relative text-gray-300 hover:text-white transition after:absolute after:left-0 after:right-0 after:-bottom-1 after:h-0.5 after:bg-indigo-500 after:scale-x-0 hover:after:scale-x-100 after:transition-transform after:origin-left">
                                {item.label}
                            </Link>
                        )}
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-5">
                        <div>
                            <h1 className="text-lg md:text-xl font-semibold text-white">
                                Email dashboard
                            </h1>
                        </div>
                        <div className="flex w-full sm:w-auto gap-2">
                            <button
                                className="flex-1 sm:flex-none px-5 py-3 rounded-md border border-gray-700 text-sm font-semibold text-gray-200 hover:bg-gray-800 disabled:opacity-60"
                                onClick={openPreview}
                                disabled={uiDisabled}>
                                Preview
                            </button>
                            <button
                                className="flex-1 sm:flex-none px-5 py-3 rounded-md font-semibold text-sm text-black bg-linear-to-r from-purple-500 to-violet-600 shadow-md transition-all duration-300 hover:brightness-125 disabled:opacity-60"
                                onClick={sendEmail}
                                disabled={uiDisabled}>
                                Send email
                            </button>
                        </div>
                    </div>

                    <div className="mx-auto w-full max-w-4xl flex-1">
                        <div className="space-y-5">
                            <div className="rounded-xl border border-gray-800 p-4">
                                <label className="text-sm text-gray-400" htmlFor="mail-subject">
                                    Subject
                                </label>
                                <input
                                    id="mail-subject"
                                    value={subject}
                                    onChange={(event) => setSubject(event.target.value)}
                                    disabled={uiDisabled}
                                    className="mt-2 w-full px-3 py-3 rounded-md bg-transparent border border-gray-700 text-sm text-white placeholder:opacity-40"
                                    placeholder="Email subject"/>
                            </div>

                            <div className="rounded-xl border border-gray-800 p-4 space-y-4">
                                <div className="text-sm text-gray-400">
                                    Recipients
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                                    {[
                                        {id: "individuals" as AudienceMode, label: "Individual"},
                                        {id: "groups" as AudienceMode, label: "Groups"},
                                        {id: "everyone" as AudienceMode, label: "Everyone"}
                                    ].map((item) =>
                                        <button
                                            key={item.id}
                                            className={`px-3 py-2 rounded-md border text-sm transition ${audience === item.id ? "border-indigo-500 text-white bg-indigo-500/15" : "border-gray-800 text-gray-400 hover:text-white"}`}
                                            onClick={() => setAudience(item.id)}
                                            disabled={uiDisabled}>
                                            {item.label}
                                        </button>
                                    )}
                                </div>

                                {audience === "individuals" &&
                                    <div className="space-y-3">
                                        <div className="flex flex-col sm:flex-row gap-2">
                                            <input
                                                value={recipientInput}
                                                onChange={(event) => setRecipientInput(event.target.value)}
                                                onKeyDown={handleRecipientKeyDown}
                                                disabled={uiDisabled}
                                                className="w-full px-3 py-3 rounded-md bg-transparent border border-gray-700 text-sm text-white placeholder:opacity-40"
                                                placeholder="name@example.com"/>
                                            <button
                                                className="px-4 py-2 text-sm bg-gray-800 hover:bg-gray-700 text-white rounded-md"
                                                onClick={addRecipients}
                                                disabled={uiDisabled}>
                                                Add
                                            </button>
                                        </div>
                                        <div className="flex flex-wrap gap-2 min-h-9">
                                            {recipients.length === 0 &&
                                                <span className="text-sm text-gray-500">
                                                    No individual recipients added.
                                                </span>
                                            }
                                            {recipients.map((email) =>
                                                <button
                                                    key={email}
                                                    className="px-3 py-1.5 rounded-full bg-gray-800 text-gray-200 text-xs hover:bg-red-600"
                                                    onClick={() => setRecipients((current) => current.filter((item) => item !== email))}
                                                    disabled={uiDisabled}>
                                                    {email}
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                }

                                {audience === "groups" &&
                                    <div className="grid grid-cols-1 sm:grid-cols-5 gap-3">
                                        {GROUP_OPTIONS.map((group) =>
                                            <button
                                                key={group.id}
                                                className={`text-left rounded-lg border p-3 transition ${selectedGroups.includes(group.id) ? "border-indigo-500 bg-indigo-500/15" : "border-gray-800 hover:border-gray-600"}`}
                                                onClick={() => toggleGroup(group.id)}
                                                disabled={uiDisabled}>
                                                <span className="block text-sm text-white">
                                                    {group.label}
                                                </span>
                                            </button>
                                        )}
                                    </div>
                                }

                                {audience === "everyone" &&
                                    <div className="rounded-lg border border-gray-800 bg-black/20 px-3 py-3 text-sm text-gray-300">
                                        All active accounts
                                    </div>
                                }
                            </div>

                            <div className="rounded-xl border border-gray-800 p-4 space-y-4">
                                <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3">
                                    <div>
                                        <div className="text-sm text-gray-400">
                                            Body
                                        </div>
                                    </div>
                                    <button
                                        className="px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-300 hover:text-white"
                                        onClick={() => setEditorHtml(STARTER_BODY)}
                                        disabled={uiDisabled}>
                                        Reset body
                                    </button>
                                </div>

                                <div className="rounded-lg border border-gray-800 bg-black/20 p-2 space-y-2">
                                    <div className="flex flex-wrap gap-2">
                                        {[
                                            {label: "B", action: () => runCommand("bold"), className: "font-bold"},
                                            {label: "I", action: () => runCommand("italic"), className: "italic"},
                                            {label: "U", action: () => runCommand("underline"), className: "underline"},
                                            {label: "S", action: () => runCommand("strikeThrough"), className: "line-through"},
                                            {label: "H2", action: () => runCommand("formatBlock", "h2"), className: ""},
                                            {label: "H3", action: () => runCommand("formatBlock", "h3"), className: ""},
                                            {label: "P", action: () => runCommand("formatBlock", "p"), className: ""},
                                            {label: "UL", action: () => runCommand("insertUnorderedList"), className: ""},
                                            {label: "OL", action: () => runCommand("insertOrderedList"), className: ""},
                                            {label: "Left", action: () => runCommand("justifyLeft"), className: ""},
                                            {label: "Center", action: () => runCommand("justifyCenter"), className: ""},
                                            {label: "Right", action: () => runCommand("justifyRight"), className: ""},
                                            {label: "Quote", action: () => runCommand("formatBlock", "blockquote"), className: ""},
                                            {label: "Line", action: () => runCommand("insertHorizontalRule"), className: ""},
                                            {label: "Clear", action: () => runCommand("removeFormat"), className: ""}
                                        ].map((tool) =>
                                            <button
                                                key={tool.label}
                                                className={`min-w-10 px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800 ${tool.className}`}
                                                onMouseDown={(event) => event.preventDefault()}
                                                onClick={tool.action}
                                                disabled={uiDisabled}>
                                                {tool.label}
                                            </button>
                                        )}
                                    </div>

                                    <div className="flex flex-wrap gap-3 items-center">
                                        <div className="grid w-full grid-cols-1 md:grid-cols-3 gap-2">
                                            <div className="flex items-center gap-2 text-xs text-gray-400 rounded-md border border-gray-800 px-2 py-2">
                                                <span className="min-w-16">Text</span>
                                                <input
                                                    type="color"
                                                    value={textColor}
                                                    onChange={(event) => setTextColor(event.target.value)}
                                                    disabled={uiDisabled}
                                                    className="h-9 w-12 rounded-md bg-transparent border border-gray-800"/>
                                                <button
                                                    type="button"
                                                    className="ml-auto px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                    onMouseDown={(event) => event.preventDefault()}
                                                    onClick={() => applyInlineStyle("color", textColor)}
                                                    disabled={uiDisabled}>
                                                    Apply
                                                </button>
                                            </div>
                                            <div className="flex items-center gap-2 text-xs text-gray-400 rounded-md border border-gray-800 px-2 py-2">
                                                <span className="min-w-16">Highlight</span>
                                                <input
                                                    type="color"
                                                    value={highlightColor}
                                                    onChange={(event) => setHighlightColor(event.target.value)}
                                                    disabled={uiDisabled}
                                                    className="h-9 w-12 rounded-md bg-transparent border border-gray-800"/>
                                                <button
                                                    type="button"
                                                    className="ml-auto px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                    onMouseDown={(event) => event.preventDefault()}
                                                    onClick={() => applyInlineStyle("background-color", highlightColor)}
                                                    disabled={uiDisabled}>
                                                    Apply
                                                </button>
                                            </div>
                                            <div className="flex items-center gap-2 text-xs text-gray-400 rounded-md border border-gray-800 px-2 py-2">
                                                <span className="min-w-16">Size</span>
                                                <select
                                                    value={fontSize}
                                                    onChange={(event) => setFontSize(event.target.value)}
                                                    disabled={uiDisabled}
                                                    className="min-w-24 px-3 py-2 rounded-md bg-black border border-gray-800 text-white">
                                                    {FONT_SIZES.map((size) =>
                                                        <option key={size} value={size}>
                                                            {size}px
                                                        </option>
                                                    )}
                                                </select>
                                                <button
                                                    type="button"
                                                    className="ml-auto px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                    onMouseDown={(event) => event.preventDefault()}
                                                    onClick={() => applyInlineStyle("font-size", `${fontSize}px`)}
                                                    disabled={uiDisabled}>
                                                    Apply
                                                </button>
                                            </div>
                                        </div>
                                        <button
                                            className="px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                            onMouseDown={(event) => event.preventDefault()}
                                            onClick={insertLink}
                                            disabled={uiDisabled}>
                                            Link
                                        </button>
                                        <button
                                            className="px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                            onMouseDown={(event) => event.preventDefault()}
                                            onClick={insertImageUrl}
                                            disabled={uiDisabled}>
                                            Image URL
                                        </button>
                                    </div>
                                </div>

                                <div
                                    ref={editorRef}
                                    contentEditable={!uiDisabled}
                                    onInput={syncBody}
                                    onKeyUp={saveSelection}
                                    onMouseUp={saveSelection}
                                    onBlur={() => {
                                        saveSelection();
                                        syncBody();
                                    }}
                                    className="other-scroll min-h-72 sm:min-h-80 max-h-[60vh] overflow-y-auto rounded-lg border border-gray-700 bg-white px-4 sm:px-5 py-5 text-gray-700 text-sm leading-6 outline-none focus:border-indigo-500"
                                    style={{fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif"}}
                                    suppressContentEditableWarning/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            {previewHtml &&
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-3 sm:p-5">
                    <div className="flex h-[90vh] w-full max-w-3xl flex-col rounded-xl border border-gray-800 bg-[#0b0d10] p-3 sm:p-4">
                        <div className="flex items-center justify-between gap-3 pb-3">
                            <div>
                                <h2 className="text-sm font-semibold text-white">
                                    Email preview
                                </h2>
                                <div className="mt-1 text-xs text-gray-400">
                                    {audience === "everyone" ? "Everyone" : audience === "groups" ? `${selectedGroups.length} groups` : `${recipients.length} recipients`}
                                </div>
                            </div>
                            <button
                                className="px-4 py-2 rounded-md border border-gray-700 text-sm text-gray-200 hover:bg-gray-800"
                                onClick={() => setPreviewHtml(null)}>
                                Close
                            </button>
                        </div>
                        <iframe
                            title="Email preview"
                            srcDoc={previewHtml}
                            sandbox=""
                            className="min-h-0 flex-1 w-full rounded-lg border border-gray-800 bg-white"/>
                    </div>
                </div>
            }
        </div>
    );
}
