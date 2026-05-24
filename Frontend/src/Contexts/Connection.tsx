import {createContext, type ReactNode, type RefObject, useCallback, useContext, useRef} from "react";
import Cookies from "js-cookie";
import {CurrentTime, Sleep} from "../Utils/Time";
import NotificationManager from "./Notification";
import {APIRoute, CSRFCookiePath, FrontendRoute, MFACookiePath, Origin} from "../Values/Constants";
import {useNavigate} from "react-router";

type PopupResponseT = {
    success: boolean;
    "modify-auth"?: boolean;
    token?: string;
    expires?: string;
    state?: string;
};

type RawAPIResponseT = {
    success: boolean;
    reply: never;
    notifications: string[];
    "modify-auth": boolean;
    "new-token": string;
    "retry-after": number;
};

type ProcessedAPIResponseT = {
    success: boolean;
    reply: unknown;
};

type SendAPIRequestT = (method: string, attachAuth: boolean, attachMFA: boolean, allowCookies: boolean, closeOnSuccess: boolean, host: string, remainingPath: string, body?: BodyInit) => Promise<ProcessedAPIResponseT>;
type LogoutT = () => Promise<boolean>;
type CheckLoggedIn = (skipLogin:boolean) => Promise<boolean>;
type OpenPopupT = (key: string, URL: string, closeOnSuccess: boolean) => Promise<boolean>;

type RequestConfig = {
    Trial:number;

    Method: string;
    HostURL: string;
    RemainingPath: string;
    Body:BodyInit;

    AllowCookies: boolean;
    AttachAuth: boolean;
    AttachCSRF: boolean;
    AttachMFA: boolean;

    PurposePing: boolean;
    PurposeAuthRefresh: boolean;
    PurposeLogout: boolean;
    SkipLoginPrompt: boolean;

    CausedRefresh: boolean;
    CloseOnSuccess: boolean;
}

const createEmptyConfig = () => {
    const requestConfig:RequestConfig = {
        Trial:0,

        Method: "",
        HostURL: "",
        RemainingPath: "",
        Body:"",

        AllowCookies: false,
        AttachAuth: false,
        AttachCSRF: false,
        AttachMFA: false,


        PurposePing: false,
        PurposeAuthRefresh: false,
        PurposeLogout: false,
        SkipLoginPrompt: false,

        CausedRefresh: false,
        CloseOnSuccess: false,
    }
    return requestConfig;
}

interface ConnectionContextType {
    SendAPIRequest: SendAPIRequestT;
    OpenPopup: OpenPopupT;
    Logout: LogoutT;
    CheckLoggedIn: CheckLoggedIn;
}


const context = createContext<ConnectionContextType | undefined>(undefined);

export function ConnectionContext({children}: { children: ReactNode }) {
    const navigate = useNavigate();
    const {SendNotification} = NotificationManager();

    const AccessToken = useRef("");
    const AccessExpiry = useRef(new Date(0));
    const IsLoggedIn = useRef(false);

    const RateLimits: RefObject<Record<string, Date>> = useRef({});
    const GatewayErrors: RefObject<Record<string, number>> = useRef({});
    const currentPopups: RefObject<Record<string, Promise<boolean>>> = useRef({});
    const currentPings: RefObject<Record<string, Promise<boolean>>> = useRef({});
    const currentRefresh: RefObject<Promise<boolean> | undefined> = useRef(undefined);
    const currentLogout: RefObject<Promise<boolean> | undefined> = useRef(undefined);

    const getGatewayErrors = (host: string) => GatewayErrors.current[host] || 0;

    const resetGatewayErrors = useCallback((host: string) => {
        if (getGatewayErrors(host) !== 0) {
            SendNotification("Server is back online");
            GatewayErrors.current[host] = 0;
        }
    },[SendNotification])

    const incrementGatewayErrors = useCallback((host: string) => {
        const current = getGatewayErrors(host);
        GatewayErrors.current[host] = current + 1;
        if (current === 0) {
            SendNotification("Server unreachable. Retrying..");
        }
    },[SendNotification])

    const OpenPopup: OpenPopupT = useCallback((key, URL, closeOnSuccess) => {
        if (!currentPopups.current[URL]) {
            currentPopups.current[URL] = new Promise<boolean>((resolve) => {
                const popup = window.open(URL, key, "width=500,height=750,popup");
                if (!popup) {
                    SendNotification("Popup blocked, please allow popups for this site.");
                    delete currentPopups.current[URL];
                    resolve(false);
                    return;
                }
                let finished = false;
                function onMessage(event: MessageEvent<PopupResponseT>) {
                    if (event.source === popup && event.origin === Origin) {
                        if (event.data && event.data.success) {
                            window.removeEventListener("message", onMessage);
                            finished = true;

                            if (closeOnSuccess && window.opener) {
                                window.opener.postMessage(event.data, window.location.origin);
                                window.close();
                                return;
                            }

                            if (event.data["modify-auth"]) {
                                if (event.data.token) AccessToken.current = event.data.token;
                                if (event.data.expires) AccessExpiry.current = new Date(event.data.expires);
                                IsLoggedIn.current = !!AccessToken.current;
                            }

                            resolve(true);
                            delete currentPopups.current[URL];
                            return;
                        }
                    }
                }
                window.addEventListener("message", onMessage);
                const interval = setInterval(() => {
                    if (popup.closed) {
                        clearInterval(interval);
                        if (!finished) {
                            window.removeEventListener("message", onMessage);
                            resolve(false);
                            delete currentPopups.current[URL];
                        }
                    }
                }, 200);
            });
        }
        return currentPopups.current[URL];
    }, [SendNotification])

    const pingServer = useCallback((host: string): Promise<boolean> => {
        if (!currentPings.current[host]) {
            const config = createEmptyConfig()
            config.PurposePing = true;
            config.HostURL = host
            config.RemainingPath = "/status/ready"
            config.Method = "GET"
            currentPings.current[host] = sendRequestInternal(config)
                .then(() => true)
                .catch(() => false)
                .finally(() => delete currentPings.current[host]);
        }
        return currentPings.current[host];
    },[])

    const refreshToken = useCallback(async (skipLogin:boolean) => {
        if (!currentRefresh.current) {
            const config = createEmptyConfig()
            config.AllowCookies = true;
            config.AttachCSRF = true;
            config.PurposeAuthRefresh = true;
            config.SkipLoginPrompt = skipLogin
            config.HostURL = APIRoute
            config.RemainingPath = "/access/refresh"
            config.Method = "POST"
            currentRefresh.current = sendRequestInternal(config)
                .then(() => true)
                .catch(() => false)
                .finally(() => currentRefresh.current = undefined);
        }
        return currentRefresh.current;
    },[])

    const Logout: LogoutT = useCallback(() => {
        if (!currentLogout.current) {
            const config = createEmptyConfig()
            config.AttachAuth = true;
            config.PurposeLogout = true;
            config.HostURL = APIRoute
            config.RemainingPath = "/access/logout"
            config.Method = "POST"
            currentLogout.current = sendRequestInternal(config)
                .then(() => true)
                .catch(() => false)
                .finally(() => currentLogout.current = undefined);
        }
        return currentLogout.current;
    },[])

    const SendAPIRequest:SendAPIRequestT = useCallback((method, attachAuth, attachMFA, allowCookies, closeOnSuccess, host, path, body?:BodyInit) => {
        const config = createEmptyConfig()
        config.AttachAuth = attachAuth
        config.AttachMFA = attachMFA
        config.AllowCookies = allowCookies
        config.Method = method
        config.CloseOnSuccess = closeOnSuccess
        config.HostURL = host
        config.RemainingPath = path
        config.Body = body || ""
        return sendRequestInternal(config)
    },[])

    const PromptMFA = useCallback(async () => OpenPopup("MFA", FrontendRoute + "/mfa", false),[OpenPopup])

    const PromptLogin = useCallback(async () => OpenPopup("SIGNIN", FrontendRoute + "/signin", false),[OpenPopup])

    const CheckLoggedIn: CheckLoggedIn = useCallback(async (skipLogin:boolean) => {
        const accessIsUsable = IsLoggedIn.current && AccessExpiry.current.getTime() > CurrentTime();
        if (accessIsUsable) {
            return true;
        }
        if (!skipLogin) return await refreshToken(false);
        return false;
    },[refreshToken])

    const sendRequestInternal:(config:RequestConfig)=>Promise<ProcessedAPIResponseT> = useCallback(async function sendReq(config:RequestConfig){
        if (config.Trial++ > 3) {
            SendNotification("Too many failed retries")
            return Promise.reject("Too many retries")
        }
        const headers: Record<string, string> = {};

        // Wait if server unresponsive (allow ping requests only)
        if (!config.PurposePing) {
            while (getGatewayErrors(config.HostURL) != 0) {
                if (await pingServer(config.HostURL)) {
                    break;
                }
                await Sleep(1000);
            }
        }
        // Wait if route is rate limited
        const rateLimitKey = config.HostURL + config.RemainingPath;
        if (rateLimitKey in RateLimits.current) {
            const retryAfter = RateLimits.current[rateLimitKey];
            if (CurrentTime() < retryAfter.getTime()) {
                SendNotification(`Rate limit reached. Please retry after ${Math.trunc((retryAfter.getTime() - CurrentTime()) / 1000)} seconds`);
                return Promise.reject("Rate limited");
            }
        }
        // Logout needs pre-logged in and auth attached
        if (config.PurposeLogout) {
            if (!await CheckLoggedIn(true)) return Promise.reject("Not Logged in");
            if (!config.AttachAuth) config.AttachAuth = true;
            if (!config.AllowCookies) config.AllowCookies = true;
        }
        // Refresh needs cookies
        if (config.PurposeAuthRefresh) {
            if (!config.AllowCookies) config.AllowCookies = true;
            if (!config.AttachCSRF) config.AttachCSRF = true;
        }
        // MFA requires auth attached and MFA cookie present, else user will be prompted for MFA
        if (config.AttachMFA) {
            if (!Cookies.get(MFACookiePath) && !await PromptMFA()) return Promise.reject("MFA incomplete");
            headers["mfa"] = Cookies.get(MFACookiePath)||"";
            if (!config.AttachAuth) config.AttachAuth = true;
        }
        // CSRF requires CSRF cookie present, else user will be prompted for Login
        if (config.AttachCSRF) {
            if (!Cookies.get(CSRFCookiePath) && !await PromptLogin()) return Promise.reject("Login incomplete");
            headers["csrf"] = Cookies.get(CSRFCookiePath)||"";
        }
        // Attach auth requires valid access token else will trigger a refresh, which on failure will trigger a login prompt
        if (config.AttachAuth) {
            if (!await CheckLoggedIn(false)) return Promise.reject("Access absent");
            headers["authorization"] = `Bearer ${AccessToken.current}`;
        }
        const requestInit:RequestInit = {
            method:config.Method,
            credentials:config.AllowCookies?"include":"omit",
            headers:headers,
            ...(config.Method.toUpperCase() !== "GET" &&
                config.Method.toUpperCase() !== "HEAD" && {
                    body: config.Body
                })
        }
        try {
            const result = await fetch(config.HostURL+config.RemainingPath, requestInit)
            let data: RawAPIResponseT = {} as RawAPIResponseT;
            try {
                data = await result.json()
                if (data.notifications) {
                    data.notifications.forEach((notification) => SendNotification(notification));
                }
            } catch (error) {
                console.log("fetch data json parse failed", error)
            }

            switch (result.status) {
                case 200: {
                    resetGatewayErrors(config.HostURL);
                    if (data["modify-auth"]) {
                        AccessToken.current = data["new-token"];
                        AccessExpiry.current = new Date(data.reply);
                        IsLoggedIn.current = !!AccessToken.current;
                    }
                    if (config.CloseOnSuccess && window.opener && data.success) {
                        const popupResponse: PopupResponseT = {
                            success: data.success,
                            "modify-auth": data["modify-auth"],
                            token: AccessToken.current,
                            expires: AccessExpiry.current.toISOString(),
                        };
                        window.opener.postMessage(popupResponse, window.location.origin);
                        window.close();
                    }
                    return {success: data.success, reply: data.reply};
                }

                case 401: {
                    if (!config.AttachAuth && !config.PurposeAuthRefresh) {
                        SendNotification("Retrying with authentication. Please report this incident to admin");
                        config.AttachAuth = true;
                        return sendReq(config);
                    }
                    if (config.AttachAuth) {
                        if (!config.CausedRefresh) {
                            config.CausedRefresh = true;
                            if (await refreshToken(config.SkipLoginPrompt)) {
                                return sendReq(config);
                            }
                            SendNotification("You need to be logged in to a valid account to perform this action.");
                            return Promise.reject("Not logged in/Session expired/revoked");
                        }
                        if (!config.SkipLoginPrompt && await PromptLogin()) return sendReq(config);
                        SendNotification("You do not have enough permissions to perform this action.");
                        navigate("/", {replace: true});
                        return Promise.reject("Invalid permissions");
                    }
                    if (config.PurposeAuthRefresh) {
                        if (await PromptLogin()) {
                            const response: ProcessedAPIResponseT = {success: data.success, reply: ""};
                            return response
                        }
                        SendNotification("You are not logged in.");
                        return Promise.reject("Session expired/revoked");
                    }
                    return Promise.reject("Not authenticated");
                }

                case 403: {
                    if (!config.AttachMFA) {
                        config.AttachMFA = true;
                        return await sendReq(config);
                    }
                    if (await PromptMFA()) return await sendReq(config);
                    SendNotification("MFA verification is required for this action.");
                    return Promise.reject("MFA incomplete");
                }

                case 422: {
                    SendNotification("Frontend has errors, please refresh and retry or report this to admin.");
                    return Promise.reject("Frontend Errors");
                }

                case 429: {
                    const retryAfterRaw = data["retry-after"] || 1;
                    if (!retryAfterRaw) return Promise.reject("Rate limited");
                    RateLimits.current[config.HostURL + config.RemainingPath] = new Date(CurrentTime() + Number(retryAfterRaw) * 1000);
                    SendNotification(`Rate limit reached. Please retry after ${retryAfterRaw} seconds.`);
                    return Promise.reject("Rate limited");
                }
                
                case 500: {
                    SendNotification("Server error. Please try again or report it to admins.")
                    return Promise.reject("Server error");
                }

                case 404:
                case 502:
                case 504: {
                    incrementGatewayErrors(config.HostURL);
                    SendNotification("Server not reachable. Retrying automatically.")
                    await Sleep(1000);
                    return await sendReq(config);
                }

                default:
                    SendNotification(`Unknown error (${result.status}). Retrying automatically.`)
                    await Sleep(1000);
                    return await sendReq(config);
            }
        } catch (error) {
            console.error(error);
            SendNotification("Unable to send request. Check console for more information.");
            return Promise.reject(error);
        }
    },[CheckLoggedIn, PromptLogin, PromptMFA, SendNotification, incrementGatewayErrors, navigate, pingServer, refreshToken, resetGatewayErrors])

    return <context.Provider value={{SendAPIRequest, OpenPopup, Logout, CheckLoggedIn}}>
        {children}
    </context.Provider>;
}

export default function ConnectionManager() {
    const ctx = useContext(context);
    if (ctx === undefined) {
        throw new Error("ConnectionManager() must be used within a ConnectionContext");
    }
    return ctx;
}


