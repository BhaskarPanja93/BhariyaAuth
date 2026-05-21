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

        ConnectivityTestPurpose?: boolean;
        AuthRefreshPurpose?: boolean;
        CloseIfPopup?: boolean;
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
    "retry-after": number | string;
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
type EnsureLoggedInT = () => Promise<boolean>;
type OpenPopupT = (key: string, URL: string, closeOnSuccess: boolean) => Promise<boolean>;

interface ConnectionContextType {
    SendGet: SendGetT;
    SendPost: SendPostT;
    OpenPopup: OpenPopupT;
    Logout: LogoutT;
    EnsureLoggedIn: EnsureLoggedInT;
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

    const refreshCredentialConnection = useMemo(() => axios.create({withCredentials: true}), []);
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
            throw new Error("Missing connection metadata on axios request config.");
        }
    };

    const pingServer = useCallback((host: string): Promise<boolean> => {
        if (!currentPings.current[host]) {
            const path = "/status/ready";
            const config: ConnectionRequestConfig = {
                HostURL: host,
                RemainingPath: path,
                ConnectivityTestPurpose: true,
                connection: cookieDisabledConnection,
            };
            currentPings.current[host] = config.connection.get(host + path, config)
                .then(() => true)
                .catch(() => false)
                .finally(() => delete currentPings.current[host]);
        }
        return currentPings.current[host];
    },[cookieDisabledConnection])

    const refreshToken = useCallback(async () => {
        if (!currentRefresh.current) {
            const path = "/access/refresh";
            const config: ConnectionRequestConfig = {
                HostURL: APIRoute,
                RemainingPath: path,
                AttachCSRF: true,
                AuthRefreshPurpose: true,
                connection: refreshCredentialConnection,
            };

            currentRefresh.current = Promise.resolve(config.connection.post(APIRoute + path, null, config))
                .then(() => true)
                .catch(() => false)
                .finally(() => {
                    currentRefresh.current = undefined;
                });
        }

        return currentRefresh.current;
    },[refreshCredentialConnection])

    const RetryRequest = useCallback(async (config: AxiosRequestConfig) => {
        validateConnectionConfig(config);
        try {
            return await config.connection(config);
        } catch (error) {
            return Promise.reject(error);
        }
    }, [])

    const OpenPopup: OpenPopupT = useCallback(async (key, URL, closeOnSuccess) => {
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
                            }

                            if (event.data["modify-auth"]) {
                                const token = event.data.token;
                                const expires = event.data.expires;
                                if (token) AccessToken.current = token;
                                if (expires) AccessExpiry.current = new Date(expires);
                                IsLoggedIn.current = !!AccessToken.current;
                            }

                            delete currentPopups.current[URL];
                            resolve(true);
                            return;
                        }
                    }
                }

                window.addEventListener("message", onMessage);
                const interval = setInterval(() => {
                    if (popup.closed) {
                        clearInterval(interval);
                        if (!finished) {
                            delete currentPopups.current[URL];
                            window.removeEventListener("message", onMessage);
                            resolve(false);
                        }
                    }
                }, 200);
            });
        }

        return currentPopups.current[URL];
    }, [SendNotification])

    const Logout: LogoutT = async () => {
        if (currentLogout.current == null) {
            const path = "/access/logout";
            const config: ConnectionRequestConfig = {
                HostURL: APIRoute,
                RemainingPath: path,
                AttachAuth: true,
                connection: cookieDisabledConnection,
            };

            currentLogout.current = config.connection.post(APIRoute + path, null, config)
                .then(() => true)
                .catch(() => false)
                .finally(() => {
                    currentLogout.current = undefined;
                });
        }

        return currentLogout.current;
    };

    const SendGet: SendGetT = useCallback(async (attachAuth, attachMFA, closeOnSuccess, host, path) => {
        const config: ConnectionRequestConfig = {
            HostURL: host,
            RemainingPath: path,
            AttachAuth: attachAuth,
            AttachMFA: attachMFA,
            CloseIfPopup: closeOnSuccess,
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
            CloseIfPopup: closeOnSuccess,
            connection: cookieDisabledConnection,
        };
        return config.connection.post(host + path, data, config);
    },[cookieDisabledConnection])

    const PromptMFA = useCallback(async () => OpenPopup("MFA", FrontendRoute + "/mfa", false),[OpenPopup])

    const PromptLogin = useCallback(async () => OpenPopup("SIGNIN", FrontendRoute + "/signin", false),[OpenPopup])

    const EnsureLoggedIn: EnsureLoggedInT = useCallback(async () => {
        const accessIsUsable = IsLoggedIn.current && AccessExpiry.current.getTime() > CurrentTime();
        if (accessIsUsable) {
            return true;
        }
        return await refreshToken();
    },[refreshToken])

    const RequestFulfilledInterceptor = useCallback(async (config: InternalAxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
        validateConnectionConfig(config);

        if (!config.ConnectivityTestPurpose) {
            while (getGatewayErrors(config.HostURL) > 0) {
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

        if (config.AttachAuth) {
            if (!await EnsureLoggedIn()) return Promise.reject("Auth absent");
            config.headers["authorization"] = `Bearer ${AccessToken.current}`;
        }

        if (config.AttachMFA) {
            if (!await EnsureLoggedIn()) return Promise.reject("Auth absent");
            let mfa = Cookies.get(MFACookiePath);
            if (!mfa && !await PromptMFA()) return Promise.reject("MFA incomplete");
            mfa = Cookies.get(MFACookiePath);
            if (!mfa) return Promise.reject("MFA absent");
            config.headers["mfa"] = mfa;
        }

        if (config.AttachCSRF) {
            let csrf = Cookies.get(CSRFCookiePath);
            if (!csrf && !await PromptLogin()) return Promise.reject("Login incomplete");
            csrf = Cookies.get(CSRFCookiePath);
            if (!csrf) return Promise.reject("CSRF absent");
            config.headers["csrf"] = csrf;
        }

        return config;
    }, [pingServer, SendNotification, EnsureLoggedIn, PromptMFA, PromptLogin])

    const RequestRejectedInterceptor = useCallback(async (error: AxiosError) => Promise.reject(error),[])

    const ResponseFulfilledInterceptor = useCallback(async (response: AxiosResponse<RawAPIResponseT>): Promise<ProcessedAPIResponseT> => {
        const config = response.config;
        validateConnectionConfig(config);

        const data = response.data;
        const status = response.status;

        if (data?.notifications) {
            data.notifications.forEach((notification) => SendNotification(notification));
        }

        if (status === 200) {
            resetGatewayErrors(config.HostURL);
            if (data["modify-auth"]) {
                AccessToken.current = data["new-token"];
                AccessExpiry.current = new Date(data.reply);
                IsLoggedIn.current = !!AccessToken.current;
            }
            if (config.CloseIfPopup && window.opener && data.success) {
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
            if (!config.AttachAuth && !config.AuthRefreshPurpose) {
                SendNotification("Retrying with authentication. Please report this incident to admin");
                config.AttachAuth = true;
                return await RetryRequest(config);
            }

            if (config.AttachAuth) {
                if (!config.CausedRefresh) {
                    IsLoggedIn.current = false;
                    if (await refreshToken()) {
                        config.CausedRefresh = true;
                        return await RetryRequest(config);
                    }
                    SendNotification("You need to be logged in to a valid account to perform this action.");
                    return Promise.reject("Not logged in/Session expired/revoked");
                }

                SendNotification("You do not have enough permissions to perform this action.");
                navigate("/", {replace: true});
                return Promise.reject("Invalid permissions");
            }

            if (config.AuthRefreshPurpose) {
                const loggedIn = await PromptLogin();
                if (loggedIn) {
                    return loggedIn;
                }
                if (!config.AttachAuth) {
                    SendNotification("You are not logged in. Please login and try again.");
                }
                return Promise.reject("Session expired/revoked");
            }
        } else if (status === 403) {
            if (!config.AttachMFA) {
                config.AttachMFA = true;
                return await RetryRequest(config);
            }
            SendNotification("MFA verification is required for this action.");
            return Promise.reject("MFA required");
        } else if (status === 422) {
            SendNotification("Frontend has errors, please refresh and retry or report this to admin.");
            return Promise.reject("Frontend Errors");
        } else if (status === 429) {
            const retryAfterRaw = data?.["retry-after"];
            if (!retryAfterRaw) return Promise.reject("Rate limited");
            const retryAfter = new Date(CurrentTime() + Number(retryAfterRaw) * 1000);
            RateLimits.current[config.HostURL + config.RemainingPath] = retryAfter;
            SendNotification(`Rate limit reached. Please retry after ${Math.trunc((retryAfter.getTime() - CurrentTime()) / 1000)} seconds.`);
            return Promise.reject("Rate limited");
        } else if (status === 500) {
            return Promise.reject("Server error");
        } else if (status === 502 || status === 504) {
            incrementGatewayErrors(config.HostURL);
            await Sleep(1000);
            return await RetryRequest(config);
        } else {
            await Sleep(1000);
            return await RetryRequest(config);
        }

        return Promise.reject(error);
    }, [SendNotification, RetryRequest, navigate, refreshToken, PromptLogin, incrementGatewayErrors])

    const ResponseFulfilledInterceptorCompat = ResponseFulfilledInterceptor as unknown as (response: AxiosResponse) => AxiosResponse | Promise<AxiosResponse>;

    useEffect(() => {
        const refreshRequestId = refreshCredentialConnection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor);
        const refreshResponseId = refreshCredentialConnection.interceptors.response.use(ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor);
        const cookieRequestId = cookieDisabledConnection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor);
        const cookieResponseId = cookieDisabledConnection.interceptors.response.use(ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor);

        return () => {
            refreshCredentialConnection.interceptors.request.eject(refreshRequestId);
            refreshCredentialConnection.interceptors.response.eject(refreshResponseId);
            cookieDisabledConnection.interceptors.request.eject(cookieRequestId);
            cookieDisabledConnection.interceptors.response.eject(cookieResponseId);
        };
    }, [RequestFulfilledInterceptor, RequestRejectedInterceptor, ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor, cookieDisabledConnection, refreshCredentialConnection]);

    return <context.Provider value={{SendGet, SendPost, OpenPopup, Logout, EnsureLoggedIn}}>
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


