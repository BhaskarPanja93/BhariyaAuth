import {type KeyboardEvent, useEffect, useMemo, useRef, useState} from "react";
import ConnectionManager from "../Contexts/Connection.tsx";
import NotificationManager from "../Contexts/Notification.tsx";
import {EmailIsValid} from "../Utils/Strings.ts";
import {APIRoute, FrontendRoute} from "../Values/Constants.ts";

type AudienceMode = "individuals" | "groups" | "everyone";

const GROUP_OPTIONS = [
    {id: "V", label: "Viewers"},
    {id: "M", label: "Moderators"},
    {id: "A", label: "Admins"},
    {id: "O", label: "Owners"},
];

const STARTER_BODY = `Write your content here`;

const FONT_SIZES = [12, 14, 16, 18, 20, 24, 28, 32];
const ALLOWED_STYLE_PROPS = new Set([
    "background",
    "background-color",
    "border",
    "border-bottom",
    "border-collapse",
    "border-left",
    "border-radius",
    "border-right",
    "border-top",
    "color",
    "display",
    "font-family",
    "font-size",
    "font-style",
    "font-weight",
    "height",
    "line-height",
    "margin",
    "margin-bottom",
    "margin-left",
    "margin-right",
    "margin-top",
    "max-height",
    "max-width",
    "min-height",
    "min-width",
    "padding",
    "padding-bottom",
    "padding-left",
    "padding-right",
    "padding-top",
    "text-align",
    "text-decoration",
    "vertical-align",
    "white-space",
    "width",
]);
const VOID_HTML_TAGS = new Set(["area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr"]);
const INLINE_TEXT_TAGS = new Set(["a", "b", "code", "del", "em", "i", "mark", "s", "small", "span", "strong", "sub", "sup", "u"]);
const FORMAT_CLEAR_TAGS = new Set(["b", "strong", "em", "i", "u", "s", "del", "mark", "span", "small", "sub", "sup", "font"]);
const STRUCTURAL_PRESERVED_ATTRS = new Set(["border", "cellpadding", "cellspacing", "colspan", "rowspan"]);
const PER_TAG_PRESERVED_ATTRS: Record<string, Set<string>> = {
    a: new Set(["href", "target", "rel"]),
    img: new Set(["src", "alt", "width", "height"]),
};

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

function stripFormattingFromFragment(htmlFragment: string) {
    const parser = new DOMParser();
    const body = parser.parseFromString(`<div>${htmlFragment}</div>`, "text/html").body;
    const root = body.firstElementChild as HTMLElement | null;
    if (!root) return htmlFragment;

    const cleanNode = (node: Node) => {
        if (node.nodeType !== Node.ELEMENT_NODE) return;

        const element = node as HTMLElement;
        const tagName = element.tagName.toLowerCase();

        Array.from(element.childNodes).forEach(cleanNode);

        if (FORMAT_CLEAR_TAGS.has(tagName)) {
            element.replaceWith(...Array.from(element.childNodes));
            return;
        }

        const allowedAttrs = PER_TAG_PRESERVED_ATTRS[tagName] || new Set<string>();
        Array.from(element.attributes).forEach((attribute) => {
            const attrName = attribute.name.toLowerCase();
            if (attrName.startsWith("on")) {
                element.removeAttribute(attribute.name);
                return;
            }
            if (allowedAttrs.has(attrName) || STRUCTURAL_PRESERVED_ATTRS.has(attrName)) {
                return;
            }
            element.removeAttribute(attribute.name);
        });
    };

    Array.from(root.childNodes).forEach(cleanNode);
    return root.innerHTML;
}

function formatHtmlSource(rawHtml: string) {
    const parser = new DOMParser();
    const body = parser.parseFromString(`<div>${rawHtml}</div>`, "text/html").body;
    const root = body.firstElementChild as HTMLElement | null;
    if (!root) return rawHtml;

    const indentUnit = "  ";
    const indent = (depth: number) => indentUnit.repeat(depth);

    const serializeNode = (node: Node, depth: number, preserveWhitespace: boolean): string[] => {
        if (node.nodeType === Node.TEXT_NODE) {
            const raw = node.textContent ?? "";
            if (preserveWhitespace) {
                if (!raw.trim()) return [];
                return raw.replace(/\r\n?/g, "\n").split("\n").map((line) => `${indent(depth)}${escapeHtml(line)}`);
            }

            const compact = raw.replace(/\s+/g, " ").trim();
            if (!compact) return [];
            return [`${indent(depth)}${escapeHtml(compact)}`];
        }

        if (node.nodeType === Node.COMMENT_NODE) {
            const comment = (node.textContent || "").trim();
            return comment ? [`${indent(depth)}<!--${comment}-->`] : [];
        }

        if (node.nodeType !== Node.ELEMENT_NODE) return [];

        const element = node as HTMLElement;
        const tagName = element.tagName.toLowerCase();
        const nextPreserveWhitespace = preserveWhitespace || tagName === "pre";
        const attrs = Array.from(element.attributes)
            .map((attribute) => `${attribute.name}="${escapeAttribute(attribute.value)}"`)
            .join(" ");
        const startTag = attrs ? `<${tagName} ${attrs}>` : `<${tagName}>`;

        if (VOID_HTML_TAGS.has(tagName)) {
            return [`${indent(depth)}${startTag}`];
        }

        const childLines = Array.from(element.childNodes).flatMap((child) => serializeNode(child, depth + 1, nextPreserveWhitespace));

        if (childLines.length === 0) {
            return [`${indent(depth)}${startTag}</${tagName}>`];
        }

        const isInline = INLINE_TEXT_TAGS.has(tagName);
        if (isInline && childLines.length === 1 && !childLines[0].includes("\n")) {
            const compactChild = childLines[0].trimStart();
            return [`${indent(depth)}${startTag}${compactChild}</${tagName}>`];
        }

        return [
            `${indent(depth)}${startTag}`,
            ...childLines,
            `${indent(depth)}</${tagName}>`
        ];
    };

    const lines = Array.from(root.childNodes).flatMap((node) => serializeNode(node, 0, false));
    return lines.join("\n").trim();
}

function buildSyntaxHighlightedHtml(source: string) {
    const escaped = escapeHtml(source);

    const colorTag = (token: string) => {
        const parsed = token.match(/^(&lt;\/?)([a-zA-Z][\w:-]*)([\s\S]*?)(&gt;)$/);
        if (!parsed) {
            return `<span style="color:#94a3b8;">${token}</span>`;
        }

        const [, openPart, tagName, attrPart, closePart] = parsed;
        const coloredAttrs = attrPart.replace(
            /([a-zA-Z_:][-a-zA-Z0-9_:.]*)(\s*=\s*)("[^"]*"|'[^']*'|[^\s"'=<>`]+)/g,
            `<span style="color:#86efac;">$1</span>$2<span style="color:#fbbf24;">$3</span>`
        );

        return `<span style="color:#94a3b8;">${openPart}</span><span style="color:#67e8f9;">${tagName}</span>${coloredAttrs}<span style="color:#94a3b8;">${closePart}</span>`;
    };

    let result = "";
    let cursor = 0;
    const tokenRegex = /(&lt;!--[\s\S]*?--&gt;)|(&lt;[\s\S]*?&gt;)/g;

    for (const match of escaped.matchAll(tokenRegex)) {
        const full = match[0];
        const index = match.index ?? 0;

        result += escaped.slice(cursor, index);

        if (full.startsWith("&lt;!--")) {
            result += `<span style="color:#64748b;">${full}</span>`;
        } else {
            result += colorTag(full);
        }

        cursor = index + full.length;
    }

    result += escaped.slice(cursor);
    return result || "&nbsp;";
}

function sanitizeHref(raw: string) {
    const trimmed = raw.trim();
    if (!trimmed) return "";
    if (trimmed.startsWith("#")) return trimmed;

    try {
        const parsed = new URL(trimmed);
        if (["http:", "https:", "mailto:", "tel:"].includes(parsed.protocol)) {
            return parsed.toString();
        }
    } catch {
        if (EmailIsValid(trimmed)) return `mailto:${trimmed}`;
    }

    return "";
}

function sanitizeImageSource(raw: string) {
    const trimmed = raw.trim();
    if (!trimmed) return "";

    if (/^data:image\/[a-z0-9+.-]+;base64,[a-z0-9+/=\s]+$/i.test(trimmed)) {
        return trimmed;
    }

    if (/^cid:[a-z0-9._%+\-@]+$/i.test(trimmed)) {
        return trimmed;
    }

    try {
        const parsed = new URL(trimmed);
        if (["http:", "https:"].includes(parsed.protocol)) {
            return parsed.toString();
        }
    } catch {
        return "";
    }

    return "";
}

function sanitizeInlineStyle(rawStyle: string) {
    const safeDeclarations: string[] = [];

    rawStyle.split(";").forEach((declaration) => {
        const colonIndex = declaration.indexOf(":");
        if (colonIndex <= 0) return;

        const property = declaration.slice(0, colonIndex).trim().toLowerCase();
        const value = declaration.slice(colonIndex + 1).trim();

        if (!property || !value || !ALLOWED_STYLE_PROPS.has(property)) return;
        if (/url\s*\(|expression\s*\(|javascript:|vbscript:/i.test(value)) return;

        safeDeclarations.push(`${property}: ${value}`);
    });

    return safeDeclarations.join("; ");
}

function sanitizeComposerHtml(html: string) {
    const parser = new DOMParser();
    const body = parser.parseFromString(`<div>${html}</div>`, "text/html").body;
    const root = body.firstElementChild as HTMLElement | null;
    if (!root) return "";

    const blockedTags = new Set(["SCRIPT", "STYLE", "IFRAME", "EMBED", "OBJECT", "META", "BASE", "FORM", "INPUT", "SELECT", "TEXTAREA", "BUTTON"]);
    const allowedTags = new Set([
        "A", "B", "BLOCKQUOTE", "BR", "CAPTION", "CODE", "DEL", "DIV", "EM", "H1", "H2", "H3", "H4", "H5", "H6", "HR",
        "I", "IMG", "LI", "MARK", "OL", "P", "PRE", "S", "SMALL", "SPAN", "STRONG", "SUB", "SUP", "TABLE", "TBODY",
        "TD", "TH", "THEAD", "TR", "U", "UL"
    ]);
    const sharedAttrs = new Set(["align", "class", "colspan", "height", "id", "rowspan", "style", "title", "valign", "width"]);
    const perTagAttrs: Record<string, Set<string>> = {
        A: new Set(["href", "target", "rel"]),
        IMG: new Set(["src", "alt"]),
        TABLE: new Set(["cellpadding", "cellspacing", "border"]),
        TD: new Set(["cellpadding", "cellspacing"]),
        TH: new Set(["cellpadding", "cellspacing"]),
    };

    const cleanNode = (node: Node) => {
        if (node.nodeType !== Node.ELEMENT_NODE) return;

        const element = node as HTMLElement;
        const tagName = element.tagName;

        Array.from(element.childNodes).forEach(cleanNode);

        if (blockedTags.has(tagName)) {
            element.remove();
            return;
        }

        if (!allowedTags.has(tagName)) {
            element.replaceWith(...Array.from(element.childNodes));
            return;
        }

        Array.from(element.attributes).forEach((attribute) => {
            const attrName = attribute.name.toLowerCase();
            if (attrName.startsWith("on")) {
                element.removeAttribute(attribute.name);
                return;
            }

            const tagAllowed = perTagAttrs[tagName]?.has(attrName) ?? false;
            if (!sharedAttrs.has(attrName) && !tagAllowed) {
                element.removeAttribute(attribute.name);
                return;
            }

            if (attrName === "style") {
                const style = sanitizeInlineStyle(attribute.value);
                if (style) element.setAttribute("style", style);
                else element.removeAttribute("style");
            }
        });

        if (tagName === "A") {
            const href = sanitizeHref(element.getAttribute("href") || "");
            if (!href) {
                element.removeAttribute("href");
            } else {
                element.setAttribute("href", href);
                if (element.getAttribute("target") === "_blank") {
                    element.setAttribute("rel", "noopener noreferrer");
                } else {
                    element.removeAttribute("target");
                    element.removeAttribute("rel");
                }
            }
        }

        if (tagName === "IMG") {
            const src = sanitizeImageSource(element.getAttribute("src") || "");
            if (!src) {
                element.remove();
                return;
            }

            element.setAttribute("src", src);
            element.setAttribute("alt", element.getAttribute("alt") || "Email image");

            const width = element.getAttribute("width");
            const height = element.getAttribute("height");
            if (width && !/^\d{1,4}(px|%)?$/i.test(width.trim())) element.removeAttribute("width");
            if (height && !/^\d{1,4}(px|%)?$/i.test(height.trim())) element.removeAttribute("height");
        }
    };

    Array.from(root.childNodes).forEach(cleanNode);
    return root.innerHTML.trim();
}

function getHtmlText(html: string) {
    const parser = new DOMParser();
    return parser.parseFromString(`<div>${html}</div>`, "text/html").body.textContent?.trim() || "";
}

function hasVisualContent(html: string) {
    const text = getHtmlText(html);
    if (text) return true;
    return /<(img|hr|table|ul|ol|li|blockquote|h[1-6]|br)\b/i.test(html);
}

function buildEmailHtml(subject: string, rawContent: string) {
    const safeSubject = escapeHtml(subject.trim());
    const cleanContent = sanitizeComposerHtml(rawContent) || "<p>Type your content here.</p>";

    return `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>${safeSubject}</title>
</head>
<body style="margin:0; padding:0; background-color:#eef0f3; font-family:-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;">
<table cellpadding="0" cellspacing="0" width="100%" style="background-color:#eef0f3;">
    <tr>
        <td align="center" style="padding:32px 12px;">
            <table cellpadding="0" cellspacing="0" width="100%" style="max-width:640px; background-color:#ffffff; border:1px solid #e5e7eb; border-radius:12px; overflow:hidden;">
                <tr>
                    <td align="center" style="background:linear-gradient(135deg, #1f2937, #0b0d10); padding:24px;">
                        <img src="${FrontendRoute}/favicons/DarkMode.png" width="120" alt="Bhariya" style="display:block;" />
                    </td>
                </tr>
                <tr>
                    <td style="padding:28px 24px;">
                        <div style="color:#1f2937; font-size:15px; line-height:1.6;">
                            ${cleanContent}
                        </div>
                    </td>
                </tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>`;
}

export default function MailStructure() {
    const {SendNotification} = NotificationManager();
    const {SendAPIRequest} = ConnectionManager();

    const sourceRef = useRef<HTMLTextAreaElement>(null);
    const highlightLayerRef = useRef<HTMLPreElement>(null);

    const [uiDisabled, setUiDisabled] = useState(false);
    const [previewMode, setPreviewMode] = useState<boolean>(false);
    const [subject, setSubject] = useState("");
    const [audience, setAudience] = useState<AudienceMode>("individuals");
    const [recipients, setRecipients] = useState<string[]>([]);
    const [bodyHtml, setBodyHtml] = useState(STARTER_BODY);
    const [recipientInput, setRecipientInput] = useState("");

    const [textColor, setTextColor] = useState("#1f2937");
    const [highlightColor, setHighlightColor] = useState("#fff3bf");
    const [fontSize, setFontSize] = useState(16);

    const [linkHref, setLinkHref] = useState("https://");
    const [linkText, setLinkText] = useState("Click here");
    const [openLinkInNewTab, setOpenLinkInNewTab] = useState(true);

    const [imageUrl, setImageUrl] = useState("https://");
    const [imageAlt, setImageAlt] = useState("Email image");
    const [imageWidth, setImageWidth] = useState("100%");

    const [tableRows, setTableRows] = useState(2);
    const [tableCols, setTableCols] = useState(2);

    const previewHtml = useMemo(() => buildEmailHtml(subject, bodyHtml), [subject, bodyHtml]);
    const cleanedBody = useMemo(() => sanitizeComposerHtml(bodyHtml), [bodyHtml]);
    const syntaxHighlightedBody = useMemo(() => buildSyntaxHighlightedHtml(bodyHtml), [bodyHtml]);

    const setSourceValue = (nextValue: string, selectionStart?: number, selectionEnd?: number) => {
        setBodyHtml(nextValue);

        if (selectionStart === undefined || selectionEnd === undefined) return;

        requestAnimationFrame(() => {
            const textarea = sourceRef.current;
            if (!textarea) return;
            textarea.focus();
            textarea.setSelectionRange(selectionStart, selectionEnd);
        });
    };

    const wrapSelection = (startTag: string, endTag: string, placeholder: string) => {
        const textarea = sourceRef.current;
        if (!textarea) return;

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const selected = bodyHtml.slice(start, end);
        const content = selected || placeholder;
        const nextValue = `${bodyHtml.slice(0, start)}${startTag}${content}${endTag}${bodyHtml.slice(end)}`;
        const nextStart = start + startTag.length;
        const nextEnd = nextStart + content.length;
        setSourceValue(nextValue, nextStart, nextEnd);
    };

    const insertAtCursor = (snippet: string, selectSnippet = false) => {
        const textarea = sourceRef.current;
        if (!textarea) return;

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const prefix = start > 0 && bodyHtml[start - 1] !== "\n" ? "\n" : "";
        const suffix = end < bodyHtml.length && bodyHtml[end] !== "\n" ? "\n" : "";
        const insertion = `${prefix}${snippet}${suffix}`;
        const nextValue = `${bodyHtml.slice(0, start)}${insertion}${bodyHtml.slice(end)}`;

        if (selectSnippet) {
            const base = start + prefix.length;
            setSourceValue(nextValue, base, base + snippet.length);
            return;
        }

        const cursor = start + insertion.length;
        setSourceValue(nextValue, cursor, cursor);
    };


    const syncHighlightScroll = () => {
        const textarea = sourceRef.current;
        const layer = highlightLayerRef.current;
        if (!textarea || !layer) return;
        layer.scrollTop = textarea.scrollTop;
        layer.scrollLeft = textarea.scrollLeft;
    };

    const createTableSnippet = () => {
        const rows = Math.max(1, Math.min(tableRows, 10));
        const cols = Math.max(1, Math.min(tableCols, 8));
        const headers = Array.from({length: cols}, (_, index) => `      <th>Header ${index + 1}</th>`).join("\n");
        const bodyRows = Array.from({length: rows}, (_, rowIndex) =>
            `    <tr>\n${Array.from({length: cols}, (_, colIndex) => `      <td>Row ${rowIndex + 1}, Col ${colIndex + 1}</td>`).join("\n")}\n    </tr>`
        ).join("\n");

        return `<table border="1" cellpadding="8" cellspacing="0" style="border-collapse:collapse; width:100%;">
  <thead>
    <tr>
${headers}
    </tr>
  </thead>
  <tbody>
${bodyRows}
  </tbody>
</table>`;
    };

    const insertLinkSnippet = () => {
        const href = sanitizeHref(linkHref);
        if (!href) {
            SendNotification("Link URL must use http, https, mailto, tel, or #anchor");
            return;
        }

        const target = openLinkInNewTab ? ` target="_blank"` : "";
        wrapSelection(`<a href="${escapeAttribute(href)}"${target}>`, "</a>", linkText.trim() || "Link text");
    };

    const insertImageSnippet = () => {
        const source = sanitizeImageSource(imageUrl);
        if (!source) {
            SendNotification("Image URL must be https/http, cid:, or a data:image base64 value");
            return;
        }

        const width = imageWidth.trim();
        const widthAttr = width ? ` width="${escapeAttribute(width)}"` : "";
        const alt = escapeAttribute(imageAlt.trim() || "Email image");
        insertAtCursor(`<img src="${escapeAttribute(source)}" alt="${alt}"${widthAttr} style="max-width:100%; height:auto;" />`);
    };

    const applyColorTag = () => {
        wrapSelection(`<span style="color:${textColor};">`, "</span>", "Colored text");
    };

    const applyHighlightTag = () => {
        wrapSelection(`<span style="background-color:${highlightColor};">`, "</span>", "Highlighted text");
    };

    const applyFontSizeTag = () => {
        wrapSelection(`<span style="font-size:${fontSize}px;">`, "</span>", "Sized text");
    };

    const clearSelectionFormatting = () => {
        const textarea = sourceRef.current;
        if (!textarea) return;

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        if (start === end) {
            SendNotification("Select a section in the editor first");
            return;
        }

        const selected = bodyHtml.slice(start, end);
        const cleanedSelection = stripFormattingFromFragment(selected);
        const nextValue = `${bodyHtml.slice(0, start)}${cleanedSelection}${bodyHtml.slice(end)}`;
        setSourceValue(nextValue, start, start + cleanedSelection.length);
    };

    const organizeHtmlEditor = () => {
        const formatted = formatHtmlSource(bodyHtml);
        if (!formatted) {
            return;
        }
        setSourceValue(formatted);
        requestAnimationFrame(() => {
            if (sourceRef.current) sourceRef.current.scrollTop = 0;
            if (highlightLayerRef.current) highlightLayerRef.current.scrollTop = 0;
        });
    };

    const parseRecipientInput = () => {
        const next = recipientInput
            .split(/[\s, ]+/)
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
        if (event.key === "Enter" || event.key === "," || event.key === " ") {
            event.preventDefault();
            addRecipients();
        }
    };

    const toggleGroup = (group: string) => {
        setRecipients((current) => current.includes(group) ? current.filter((item) => item !== group) : [...current, group]);
    };

    const validateBeforeSend = (nextRecipients: string[], sanitizedBody: string) => {
        if (!subject.trim()) {
            SendNotification("Subject is required");
            return false;
        }
        if (audience === "individuals" && nextRecipients.length === 0) {
            SendNotification("Add at least one recipient");
            return false;
        }
        if (audience === "groups" && nextRecipients.length === 0) {
            SendNotification("Choose at least one group");
            return false;
        }
        if (!hasVisualContent(sanitizedBody)) {
            SendNotification("Email body is empty");
            return false;
        }
        return true;
    };

    const sendEmail = () => {
        if (audience === "individuals" && recipientInput.trim()) {
            SendNotification("Pending recipients, please press Add or remove them")
            return;
        }

        if (!validateBeforeSend(recipients, sanitizeComposerHtml(bodyHtml))) return;

        setUiDisabled(true);
        const form = new FormData();
        form.append("subject", subject);
        form.append("audience", audience);
        form.append("body", buildEmailHtml(subject, bodyHtml));
        recipients.forEach(r => {
            form.append("recipients", r)
        })
        SendAPIRequest("POST", true, true, false, false, APIRoute, "/mail/send", form)
            .then()
            .catch((error) => {
                console.log("Mail send stopped because:", error);
            })
            .finally(() => setUiDisabled(false));
    };

    useEffect(() => {
        document.title = "Mail - Bhariya";
    }, []);

    useEffect(() => {
        syncHighlightScroll();
    }, [bodyHtml]);

    const modeLabel = previewMode ? "Preview" : "Edit HTML";

    return (
        <div className="min-h-screen p-3 sm:p-5 box-border overflow-x-hidden">
            <div className="mx-auto w-full max-w-5xl">
                <div
                    className="rounded-2xl p-4 sm:p-5 md:p-6 flex flex-col box-border min-h-[calc(100vh-24px)] sm:min-h-[calc(100vh-40px)]"
                    style={{
                        background: "linear-gradient(180deg, rgba(12,14,18,0.9), rgba(7,8,10,0.85))",
                        border: "1px solid rgba(255,255,255,0.02)"
                    }}>

                    <div className="mx-auto w-full max-w-4xl flex-1">
                        <div className="space-y-5">
                            <div className="rounded-xl border border-gray-800 p-4">
                                <label className="text-sm text-gray-400" htmlFor="mail-subject">
                                    Subject
                                </label>
                                <input
                                    id="mail-subject"
                                    value={subject}
                                    onChange={(event) => setSubject(event.target.value.trim())}
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
                                            onClick={() => {
                                                setRecipients([])
                                                setAudience(item.id)
                                            }}
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
                                    <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
                                        {GROUP_OPTIONS.map((group) =>
                                            <button
                                                key={group.id}
                                                className={`text-left rounded-lg border p-3 transition ${recipients.includes(group.id) ? "border-indigo-500 bg-indigo-500/15" : "border-gray-800 hover:border-gray-600"}`}
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
                                    <div className="text-sm text-gray-400">
                                        Body ({modeLabel})
                                    </div>
                                </div>

                                <div className="rounded-lg border border-gray-800 bg-black/20 p-3 space-y-3">
                                    <div className="text-xs text-gray-500">
                                        Quick tags
                                    </div>
                                    <div className="flex flex-wrap gap-2">
                                        {[
                                            {label: "<strong>", action: () => wrapSelection("<strong>", "</strong>", "Bold text")},
                                            {label: "<em>", action: () => wrapSelection("<em>", "</em>", "Italic text")},
                                            {label: "<u>", action: () => wrapSelection("<u>", "</u>", "Underlined text")},
                                            {label: "<s>", action: () => wrapSelection("<s>", "</s>", "Struck text")},
                                            {label: "<p>", action: () => wrapSelection("<p>", "</p>", "Paragraph")},
                                            {label: "<h1>", action: () => wrapSelection("<h1>", "</h1>", "Heading 1")},
                                            {label: "<h2>", action: () => wrapSelection("<h2>", "</h2>", "Heading 2")},
                                            {label: "<h3>", action: () => wrapSelection("<h3>", "</h3>", "Heading 3")},
                                            {label: "<blockquote>", action: () => wrapSelection("<blockquote>", "</blockquote>", "Quoted text")},
                                            {label: "<code>", action: () => wrapSelection("<code>", "</code>", "inline_code")},
                                            {label: "<pre>", action: () => wrapSelection("<pre>", "</pre>", "multi-line code")},
                                            {label: "<hr />", action: () => insertAtCursor("<hr />")},
                                            {label: "<ul>", action: () => insertAtCursor("<ul>\n  <li>Item 1</li>\n  <li>Item 2</li>\n</ul>", true)},
                                            {label: "<ol>", action: () => insertAtCursor("<ol>\n  <li>Item 1</li>\n  <li>Item 2</li>\n</ol>", true)},
                                            {label: "align-left", action: () => wrapSelection(`<div style="text-align:left;">`, "</div>", "Left aligned")},
                                            {label: "align-center", action: () => wrapSelection(`<div style="text-align:center;">`, "</div>", "Center aligned")},
                                            {label: "align-right", action: () => wrapSelection(`<div style="text-align:right;">`, "</div>", "Right aligned")},
                                        ].map((tool) =>
                                            <button
                                                key={tool.label}
                                                className="px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={tool.action}
                                                disabled={uiDisabled || previewMode}>
                                                {tool.label}
                                            </button>
                                        )}
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
                                        <div className="flex items-center gap-2 text-xs text-gray-400 rounded-md border border-gray-800 px-2 py-2">
                                            <span className="min-w-12">Text</span>
                                            <input
                                                type="color"
                                                value={textColor}
                                                onChange={(event) => setTextColor(event.target.value)}
                                                disabled={uiDisabled || previewMode}
                                                className="h-8 w-10 rounded-md bg-transparent border border-gray-800"/>
                                            <button
                                                type="button"
                                                className="ml-auto px-3 py-1.5 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={applyColorTag}
                                                disabled={uiDisabled || previewMode}>
                                                Apply
                                            </button>
                                        </div>
                                        <div className="flex items-center gap-2 text-xs text-gray-400 rounded-md border border-gray-800 px-2 py-2">
                                            <span className="min-w-12">Mark</span>
                                            <input
                                                type="color"
                                                value={highlightColor}
                                                onChange={(event) => setHighlightColor(event.target.value)}
                                                disabled={uiDisabled || previewMode}
                                                className="h-8 w-10 rounded-md bg-transparent border border-gray-800"/>
                                            <button
                                                type="button"
                                                className="ml-auto px-3 py-1.5 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={applyHighlightTag}
                                                disabled={uiDisabled || previewMode}>
                                                Apply
                                            </button>
                                        </div>
                                        <div className="flex items-center gap-2 text-xs text-gray-400 rounded-md border border-gray-800 px-2 py-2">
                                            <span className="min-w-12">Size</span>
                                            <select
                                                value={fontSize}
                                                onChange={(event) => setFontSize(Number(event.target.value))}
                                                disabled={uiDisabled || previewMode}
                                                className="min-w-20 px-2 py-1.5 rounded-md bg-black border border-gray-800 text-white">
                                                {FONT_SIZES.map((size) =>
                                                    <option key={size} value={size}>
                                                        {size}px
                                                    </option>
                                                )}
                                            </select>
                                            <button
                                                type="button"
                                                className="ml-auto px-3 py-1.5 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={applyFontSizeTag}
                                                disabled={uiDisabled || previewMode}>
                                                Apply
                                            </button>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-2">
                                        <div className="rounded-md border border-gray-800 p-2 space-y-2">
                                            <div className="text-xs text-gray-500">
                                                Anchor
                                            </div>
                                            <input
                                                value={linkHref}
                                                onChange={(event) => setLinkHref(event.target.value)}
                                                disabled={uiDisabled || previewMode}
                                                className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                placeholder="https://bhariya.ddns.net/auth"/>
                                            <input
                                                value={linkText}
                                                onChange={(event) => setLinkText(event.target.value)}
                                                disabled={uiDisabled || previewMode}
                                                className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                placeholder="Link text"/>
                                            <label className="flex items-center gap-2 text-xs text-gray-300">
                                                <input
                                                    type="checkbox"
                                                    checked={openLinkInNewTab}
                                                    onChange={(event) => setOpenLinkInNewTab(event.target.checked)}
                                                    disabled={uiDisabled || previewMode}/>
                                                Open in new tab
                                            </label>
                                            <button
                                                className="w-full px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={insertLinkSnippet}
                                                disabled={uiDisabled || previewMode}>
                                                Insert &lt;a&gt;
                                            </button>
                                        </div>

                                        <div className="rounded-md border border-gray-800 p-2 space-y-2">
                                            <div className="text-xs text-gray-500">
                                                Image
                                            </div>
                                            <input
                                                value={imageUrl}
                                                onChange={(event) => setImageUrl(event.target.value)}
                                                disabled={uiDisabled || previewMode}
                                                className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                placeholder="https://bhariya.ddns.net/auth/favicons/DarkMode.png"/>
                                            <div className="grid grid-cols-2 gap-2">
                                                <input
                                                    value={imageAlt}
                                                    onChange={(event) => setImageAlt(event.target.value)}
                                                    disabled={uiDisabled || previewMode}
                                                    className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                    placeholder="Alt text"/>
                                                <input
                                                    value={imageWidth}
                                                    onChange={(event) => setImageWidth(event.target.value)}
                                                    disabled={uiDisabled || previewMode}
                                                    className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                    placeholder="100% or 640"/>
                                            </div>
                                            <button
                                                className="w-full px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={insertImageSnippet}
                                                disabled={uiDisabled || previewMode}>
                                                Insert &lt;img&gt;
                                            </button>
                                        </div>
                                    </div>

                                    <div className="rounded-md border border-gray-800 p-2 space-y-2">
                                        <div className="text-xs text-gray-500">
                                            Table
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-4 gap-2">
                                            <input
                                                type="number"
                                                min={1}
                                                max={10}
                                                value={tableRows}
                                                onChange={(event) => setTableRows(Number(event.target.value) || 1)}
                                                disabled={uiDisabled || previewMode}
                                                className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                placeholder="Rows"/>
                                            <input
                                                type="number"
                                                min={1}
                                                max={8}
                                                value={tableCols}
                                                onChange={(event) => setTableCols(Number(event.target.value) || 1)}
                                                disabled={uiDisabled || previewMode}
                                                className="w-full px-2 py-1.5 rounded-md bg-transparent border border-gray-700 text-xs text-white"
                                                placeholder="Columns"/>
                                            <button
                                                className="sm:col-span-2 px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-200 hover:bg-gray-800"
                                                onClick={() => insertAtCursor(createTableSnippet(), true)}
                                                disabled={uiDisabled || previewMode}>
                                                Insert &lt;table&gt;
                                            </button>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-5">
                                    <div className="flex w-full sm:w-auto gap-2">
                                        <div className="flex rounded-md border border-gray-700 overflow-hidden">
                                            {[
                                                {id: "edit", state: false, label: "Edit HTML"},
                                                {id: "preview", state: true, label: "Preview"}
                                            ].map((item) =>
                                                <button
                                                    key={item.id}
                                                    className={`px-4 py-2 text-sm font-medium ${previewMode === item.state ? "bg-indigo-600 text-white" : "bg-transparent text-gray-300 hover:bg-gray-800"}`}
                                                    onClick={() => setPreviewMode(item.state)}
                                                    disabled={uiDisabled}>
                                                    {item.label}
                                                </button>
                                            )}
                                        </div>
                                        <button
                                            className="px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-300 hover:text-white"
                                            onClick={clearSelectionFormatting}
                                            disabled={uiDisabled || previewMode}>
                                            Clear selected formatting
                                        </button>
                                        <button
                                            className="px-3 py-2 rounded-md border border-gray-800 text-xs text-gray-300 hover:text-white"
                                            onClick={organizeHtmlEditor}
                                            disabled={uiDisabled}>
                                            Organise HTML
                                        </button>
                                        <button
                                            className="px-5 py-2 rounded-md font-semibold text-sm text-black bg-linear-to-r from-purple-500 to-violet-600 shadow-md transition-all duration-300 hover:brightness-125 disabled:opacity-60"
                                            onClick={sendEmail}
                                            disabled={uiDisabled}>
                                            Send email
                                        </button>
                                    </div>
                                </div>
                                {!previewMode &&
                                    <div className="relative h-[65vh] min-h-80 max-h-[65vh] w-full rounded-lg border border-gray-700 bg-white overflow-hidden">
                                        <pre
                                            ref={highlightLayerRef}
                                            aria-hidden
                                            className="other-scroll pointer-events-none absolute inset-0 m-0 overflow-auto px-4 py-4 text-sm leading-6 whitespace-pre-wrap break-words"
                                            style={{
                                                fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace"
                                            }}
                                            dangerouslySetInnerHTML={{__html: syntaxHighlightedBody}}/>
                                        <textarea
                                            ref={sourceRef}
                                            value={bodyHtml}
                                            onChange={(event) => setBodyHtml(event.target.value)}
                                            onScroll={syncHighlightScroll}
                                            disabled={uiDisabled}
                                            spellCheck={false}
                                            className="other-scroll relative z-10 h-full w-full resize-none overflow-auto bg-transparent px-4 py-4 text-sm leading-6 text-transparent outline-none focus:ring-0 placeholder:text-slate-500"
                                            style={{
                                                fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
                                                caretColor: "#7900b8"
                                            }}
                                            placeholder="Write raw HTML for email body here..."
                                        />
                                    </div>
                                }


                                {previewMode &&
                                    <div className="rounded-lg border border-gray-700 overflow-hidden">
                                        <iframe
                                            title="Email preview"
                                            srcDoc={previewHtml}
                                            sandbox=""
                                            className="h-[70vh] w-full bg-white"/>
                                    </div>
                                }

                                <div className="text-xs text-gray-500">
                                    Unsafe tags/attributes are filtered before preview and send. Sanitized body length: {cleanedBody.length}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}