import {createContext, type ReactNode, type RefObject, useCallback, useContext, useEffect, useMemo, useRef} from "react";
import type {AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig} from "axios";
import axios, {AxiosError} from "axios";
import Cookies from "js-cookie";
import {CurrentTime, Sleep} from "../Utils/Time";
import NotificationManager from "./Notification";
import {APIRoute, CSRFCookiePath, FrontendRoute, MFACookiePath, Origin} from "../Values/Constants";
import {useNavigate} from "react-router";

declare module "axios" {
    export interface AxiosRequestConfig {
        connection?: AxiosInstance;
        HostURL?: string;
        RemainingPath?: string;

        AttachAuth?: boolean;
        AttachCSRF?: boolean;
        AttachMFA?: boolean;
        CausedRefresh?: boolean;

        PurposePing?: boolean;
        PurposeAuthRefresh?: boolean;
        PurposeLogout?: boolean;
        SkipLoginPrompt?: boolean;

        CloseOnSuccess?: boolean;
    }
}

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
    reply: never;
};

type ConnectionRequestConfig = AxiosRequestConfig & {
    connection: AxiosInstance;
    HostURL: string;
    RemainingPath: string;
};

type SendGetT = (attachCreds: boolean, attachMFA: boolean, closeOnSuccess: boolean, host: string, remainingPath: string) => Promise<ProcessedAPIResponseT>;
type SendPostT = (attachCreds: boolean, attachMFA: boolean, closeOnSuccess: boolean, host: string, remainingPath: string, data?: FormData) => Promise<ProcessedAPIResponseT>;
type LogoutT = () => Promise<boolean>;
type CheckLoggedIn = (skipLogin:boolean) => Promise<boolean>;
type OpenPopupT = (key: string, URL: string, closeOnSuccess: boolean) => Promise<boolean>;

interface ConnectionContextType {
    SendGet: SendGetT;
    SendPost: SendPostT;
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

    const cookieEnabledConnection = useMemo(() => axios.create({withCredentials: true}), []);
    const cookieDisabledConnection = useMemo(() => axios.create({withCredentials: false}), []);

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

    const validateConnectionConfig: (config: AxiosRequestConfig) => asserts config is ConnectionRequestConfig = (config) => {
        if (!config.connection || !config.HostURL || !config.RemainingPath) {
            throw new Error();
        }
    };

    const pingServer = useCallback((host: string): Promise<boolean> => {
        if (!currentPings.current[host]) {
            const config: ConnectionRequestConfig = {
                HostURL: host,
                RemainingPath: "/status/ready",
                PurposePing: true,
                connection: cookieDisabledConnection,
            };
            currentPings.current[host] = config.connection.get(config.HostURL + config.RemainingPath, config)
                .then(() => true)
                .catch(() => false)
                .finally(() => delete currentPings.current[host]);
        }
        return currentPings.current[host];
    },[cookieDisabledConnection])

    const refreshToken = useCallback(async (skipLogin:boolean) => {
        if (!currentRefresh.current) {
            const config: ConnectionRequestConfig = {
                HostURL: APIRoute,
                RemainingPath: "/access/refresh",
                AttachCSRF: true,
                PurposeAuthRefresh: true,
                SkipLoginPrompt:skipLogin,
                connection: cookieEnabledConnection,
            };
            currentRefresh.current = Promise.resolve(config.connection.post(config.HostURL + config.RemainingPath,null, config))
                .then(() => true)
                .catch(() => false)
                .finally(() => {
                    currentRefresh.current = undefined;
                });
        }
        return currentRefresh.current;
    },[cookieEnabledConnection])

    const RetryRequest = useCallback(async (config: AxiosRequestConfig) => {
        validateConnectionConfig(config);
        try {
            return await config.connection(config);
        } catch (error) {
            return Promise.reject(error);
        }
    }, [])

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

    const Logout: LogoutT = useCallback(() => {
        if (!currentLogout.current) {
            const path = "/access/logout";
            const config: ConnectionRequestConfig = {
                HostURL: APIRoute,
                RemainingPath: path,
                AttachAuth: true,
                connection: cookieEnabledConnection,
            };
            currentLogout.current = config.connection.get(APIRoute + path, config)
                .then(() => true)
                .catch(() => false)
                .finally(() => {
                    currentLogout.current = undefined;
                });
        }
        return currentLogout.current;
    },[cookieEnabledConnection])

    const SendGet: SendGetT = useCallback(async (attachAuth, attachMFA, closeOnSuccess, host, path) => {
        const config: ConnectionRequestConfig = {
            HostURL: host,
            RemainingPath: path,
            AttachAuth: attachAuth,
            AttachMFA: attachMFA,
            CloseOnSuccess: closeOnSuccess,
            connection: cookieDisabledConnection,
        };
        return config.connection.get(host + path, config);
    },[cookieDisabledConnection])

    const SendPost: SendPostT = useCallback(async (attachAuth, attachMFA, closeOnSuccess, host, path, data) => {
        const config: ConnectionRequestConfig = {
            HostURL: host,
            RemainingPath: path,
            AttachAuth: attachAuth,
            AttachMFA: attachMFA,
            CloseOnSuccess: closeOnSuccess,
            connection: cookieDisabledConnection,
        };
        return config.connection.post(host + path, data, config);
    },[cookieDisabledConnection])

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

    const RequestFulfilledInterceptor = useCallback(async (config: InternalAxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
        validateConnectionConfig(config);
        if (!config.PurposePing) {
            while (getGatewayErrors(config.HostURL) != 0) {
                if (await pingServer(config.HostURL)) {
                    break;
                }
                await Sleep(1000);
            }
        }
        const rateLimitKey = config.HostURL + config.RemainingPath;
        if (rateLimitKey in RateLimits.current) {
            const retryAfter = RateLimits.current[rateLimitKey];
            if (CurrentTime() < retryAfter.getTime()) {
                SendNotification(`Rate limit reached. Please retry after ${Math.trunc((retryAfter.getTime() - CurrentTime()) / 1000)} seconds`);
                return Promise.reject("Rate limited");
            }
        }

        config.headers = config.headers || {};

        if (config.PurposeLogout) {
            if (!await CheckLoggedIn(true)) return Promise.reject("Not Logged in");
            if (!config.AttachAuth) config.AttachAuth = true;
        }

        if (config.AttachMFA) {
            let mfa = Cookies.get(MFACookiePath);
            if (!mfa && !await PromptMFA()) return Promise.reject("MFA incomplete");
            mfa = Cookies.get(MFACookiePath);
            if (!mfa) return Promise.reject("MFA absent");
            config.headers["mfa"] = mfa;
            if (!config.AttachAuth) config.AttachAuth = true;
        }

        if (config.AttachAuth) {
            if (!await CheckLoggedIn(false)) return Promise.reject("Access absent");
            config.headers["authorization"] = `Bearer ${AccessToken.current}`;
        }

        if (config.AttachCSRF) {
            let csrf = Cookies.get(CSRFCookiePath);
            if (!csrf && !await PromptLogin()) return Promise.reject("Login incomplete");
            csrf = Cookies.get(CSRFCookiePath);
            if (!csrf) return Promise.reject("CSRF absent");
            config.headers["csrf"] = csrf;
        }

        return config;
    }, [pingServer, SendNotification, CheckLoggedIn, PromptMFA, PromptLogin])

    const RequestRejectedInterceptor = useCallback(async (error: AxiosError) => Promise.reject(error),[])

    const ResponseFulfilledInterceptor = useCallback(async (response: AxiosResponse<RawAPIResponseT>): Promise<ProcessedAPIResponseT> => {
        const config = response.config;
        validateConnectionConfig(config);

        const data = response.data;
        const status = response.status;

        if (data.notifications) {
            data.notifications.forEach((notification) => SendNotification(notification));
        }

        if (status === 200) {
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
        }

        return {success: data.success, reply: data.reply};
    },[SendNotification, resetGatewayErrors])

    const ResponseRejectedInterceptor = useCallback(async (error: AxiosError<RawAPIResponseT>) => {
        const response = error.response;
        const config = response?.config;
        const data = response?.data;
        const status = response?.status;
        if (data?.notifications) {
            data.notifications.forEach((notification) => SendNotification(notification));
        }
        if (!config) return Promise.reject(error);
        validateConnectionConfig(config);
        if (status === 401) {
            if (!config.AttachAuth && !config.PurposeAuthRefresh) {
                SendNotification("Retrying with authentication. Please report this incident to admin");
                config.AttachAuth = true;
                return RetryRequest(config);
            }
            if (config.AttachAuth) {
                if (!config.CausedRefresh) {
                    config.CausedRefresh = true;
                    if (await refreshToken(false)) {
                        return RetryRequest(config);
                    }
                    SendNotification("You need to be logged in to a valid account to perform this action.");
                    return Promise.reject("Not logged in/Session expired/revoked");
                }
                if (await PromptLogin()) return RetryRequest(config);
                SendNotification("You do not have enough permissions to perform this action.");
                navigate("/", {replace: true});
                return Promise.reject("Invalid permissions");
            }
            if (config.PurposeAuthRefresh) {
                if (await PromptLogin()) {
                    return true;
                }
                SendNotification("You are not logged in.");
                return Promise.reject("Session expired/revoked");
            }
        } else if (status === 403) {
            if (!config.AttachMFA) {
                config.AttachMFA = true;
                return await RetryRequest(config);
            }
            if (await PromptMFA()) return await RetryRequest(config);
            SendNotification("MFA verification is required for this action.");
            return Promise.reject("MFA incomplete");
        } else if (status === 422) {
            SendNotification("Frontend has errors, please refresh and retry or report this to admin.");
            return Promise.reject("Frontend Errors");
        } else if (status === 429) {
            const retryAfterRaw = data?.["retry-after"] || 1;
            if (!retryAfterRaw) return Promise.reject("Rate limited");
            RateLimits.current[config.HostURL + config.RemainingPath] = new Date(CurrentTime() + Number(retryAfterRaw) * 1000);
            SendNotification(`Rate limit reached. Please retry after ${retryAfterRaw} seconds.`);
            return Promise.reject("Rate limited");
        } else if (status === 500) {
            SendNotification("Server error. Please try again or report it to admins.")
            return Promise.reject("Server error");
        } else if (status === 502 || status === 504 || status === 404) {
            incrementGatewayErrors(config.HostURL);
            SendNotification("Server not reachable. Retrying automatically.")
            await Sleep(1000);
            return await RetryRequest(config);
        } else {
            SendNotification(`Unknown error (${status}). Retrying automatically.`)
            await Sleep(1000);
            return await RetryRequest(config);
        }

        return Promise.reject(error);
    }, [SendNotification, RetryRequest, navigate, refreshToken, PromptLogin, PromptMFA, incrementGatewayErrors])

    const ResponseFulfilledInterceptorCompat = ResponseFulfilledInterceptor as unknown as (response: AxiosResponse) => AxiosResponse | Promise<AxiosResponse>;

    useEffect(() => {
        const refreshRequestId = cookieEnabledConnection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor);
        const refreshResponseId = cookieEnabledConnection.interceptors.response.use(ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor);
        const cookieRequestId = cookieDisabledConnection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor);
        const cookieResponseId = cookieDisabledConnection.interceptors.response.use(ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor);

        return () => {
            cookieEnabledConnection.interceptors.request.eject(refreshRequestId);
            cookieEnabledConnection.interceptors.response.eject(refreshResponseId);
            cookieDisabledConnection.interceptors.request.eject(cookieRequestId);
            cookieDisabledConnection.interceptors.response.eject(cookieResponseId);
        };
    }, [RequestFulfilledInterceptor, RequestRejectedInterceptor, ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor, cookieDisabledConnection, cookieEnabledConnection]);

    return <context.Provider value={{SendGet, SendPost, OpenPopup, Logout, CheckLoggedIn: CheckLoggedIn}}>
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


