import {createContext, type ReactNode, type RefObject, useContext, useRef} from 'react';
import type {AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig} from "axios";
import axios, {AxiosError, type AxiosResponse} from "axios";
import Cookies from "js-cookie";
import {CurrentTime, Sleep} from "../Utils/Time";
import NotificationManager from "./Notification.tsx";
import {APIRoute, FrontendRoute, CSRFCookiePath, MFACookiePath, Origin} from "../Values/Constants";
import {useNavigate} from "react-router";

declare module 'axios' {
    export interface AxiosRequestConfig {

        connection?:AxiosInstance;
        HostURL?:string;
        RemainingPath?:string;

        AttachAuth?:boolean;
        AttachCSRF?:boolean;
        AttachMFA?:boolean;
        AttachCookies?:boolean;
        CausedRefresh?:boolean;

        ConnectivityTestPurpose?:boolean;
        AuthRefreshPurpose?:boolean;

        CloseIfPopup?:boolean;
    }
}

type PopupResponseT = {
    success: boolean
    "modify-auth"?: boolean
    token?: string
    expires?: string
    state?: string
}

type RawAPIResponseT = {
    success: boolean,
    reply: never,
    notifications:string[],
    "modify-auth": boolean,
    "new-token": string,
    "retry-after":string,
}

type ProcessedAPIResponseT = {
    success: boolean,
    reply: never
}

type ConnectionRequestConfig = AxiosRequestConfig & {
    connection: AxiosInstance;
    HostURL: string;
    RemainingPath: string;
}

type SendGetT = (attachCreds: boolean, attachMFA: boolean, closeOnSuccess:boolean, host: string, remainingPath: string) => Promise<ProcessedAPIResponseT>
type SendPostT = (attachCreds: boolean, attachMFA: boolean, closeOnSuccess:boolean,  host: string, remainingPath: string, data?: FormData) => Promise<ProcessedAPIResponseT>
type LogoutT = () => Promise<boolean>;
type EnsureLoggedInT = () => Promise<boolean>;
type OpenPopupT = (key:string, URL: string, closeOnSuccess: boolean) => Promise<boolean>;

interface ConnectionContextType {
    SendGet: SendGetT;
    SendPost: SendPostT;
    OpenPopup: OpenPopupT;
    Logout: LogoutT;
    EnsureLoggedIn: EnsureLoggedInT;
}

const Context = createContext<ConnectionContextType | undefined>(undefined);

export function ConnectionContext ({children}: { children: ReactNode }) {
    const navigate = useNavigate();

    const {SendNotification} = NotificationManager();
    const AccessToken = useRef("")
    const AccessExpiry = useRef(new Date())
    const IsLoggedIn = useRef(false)

    const RateLimits: RefObject<Record<string, Date>> = useRef({})
    const GatewayErrors: RefObject<Record<string, number>> = useRef({})
    const currentPopups: RefObject<Record<string, Promise<boolean>>> = useRef({})
    const currentPings: RefObject<Record<string, Promise<boolean>>> = useRef({});
    const currentRefresh: RefObject<Promise<boolean>|undefined> = useRef(undefined);
    const currentLogout: RefObject<Promise<boolean>|undefined> = useRef(undefined);

    const GetGatewayErrors = (host:string) => {
        return GatewayErrors.current[host] || 0
    }

    const ResetGatewayErrors = (host:string) => {
        if (GetGatewayErrors(host) !== 0) {
            SendNotification("Server is back online")
        }
        GatewayErrors.current[host] = 0
    }

    const IncrementGatewayErrors = (host:string) => {
        const current = GetGatewayErrors(host)
        if (current == 0) {
            SendNotification("Server unreachable. Retrying..")
            GatewayErrors.current[host] = 1
        }
        GatewayErrors.current[host] = current+1
    }

    const ValidateConnectionConfig: (config: AxiosRequestConfig) => asserts config is ConnectionRequestConfig = (config) => {
        if (!config.connection || !config.HostURL || !config.RemainingPath) {
            throw new Error("Missing connection metadata on axios request config.")
        }
    }

    const SendGet:SendGetT = async (attachAuth, attachMFA, closeOnSuccess, host, path) => {
        const config: ConnectionRequestConfig = {
            HostURL: host,
            RemainingPath: path,
            AttachAuth: attachAuth,
            AttachMFA: attachMFA,
            CloseIfPopup: closeOnSuccess,
            connection: cookieDisabledConnection
        }
        return config.connection.get(path, config)
    }

    const SendPost:SendPostT = async (attachAuth, attachMFA, closeOnSuccess, host, path, data) => {
        const config: ConnectionRequestConfig = {
            HostURL: host,
            RemainingPath: path,
            AttachAuth: attachAuth,
            AttachMFA: attachMFA,
            CloseIfPopup: closeOnSuccess,
            connection: cookieDisabledConnection
        }
        return config.connection.post(host+path, data, config)
    }

    const IsServerOnline = (host: string): Promise<boolean> => {
        if (!currentPings.current[host]) {
            const path = "/status/ready"
            const config: ConnectionRequestConfig = {
                HostURL: host,
                RemainingPath: path,
                ConnectivityTestPurpose: true,
                connection: cookieDisabledConnection
            }
            currentPings.current[host] = config.connection.get(host+path, config)
                    .then(() => true)
                    .catch(() => false)
                    .finally(() => delete currentPings.current[host])
        }
        return currentPings.current[host]
    }

    const RefreshToken = async () => {
        if (!currentRefresh.current) {
            const path = "/access/refresh"
            const config: ConnectionRequestConfig = {
                HostURL: APIRoute,
                RemainingPath: path,
                AttachCSRF: true,
                AuthRefreshPurpose: true,
                connection: refreshCredentialConnection
            }
            currentRefresh.current = Promise.resolve(config.connection.post(APIRoute + path, null, config))
                .then(() => true)
                .catch(() => false)
                .finally(() => {
                    currentRefresh.current = undefined
                })
        }
        return currentRefresh.current;
    }

    const Logout:LogoutT = async () => {
        if (currentLogout.current == null) {
            const path = "/access/logout"
            const config: ConnectionRequestConfig = {
                HostURL: APIRoute,
                RemainingPath: path,
                AttachAuth: true,
                connection: cookieDisabledConnection
            }
            currentLogout.current = config.connection.post(APIRoute+path, null, config)
                    .then(() => true )
                    .catch(() => false )
                    .finally(() => { currentLogout.current = undefined });
        }
        return currentLogout.current
    }

    const PromptMFA = async () => {
        return OpenPopup("MFA", FrontendRoute+"/mfa", false)
    }

    const PromptLogin = async () => {
        return OpenPopup("SIGNIN", FrontendRoute+"/signin", false)
    }

    const EnsureLoggedIn:EnsureLoggedInT = async () => {
        return IsLoggedIn.current && AccessExpiry.current && AccessExpiry.current.getTime() > CurrentTime() ||
            await RefreshToken()
    }

    const RetryRequest = async (config:AxiosRequestConfig) => {
        ValidateConnectionConfig(config)
        try {
            return await config.connection(config)
        } catch (error) {
            return Promise.reject(error);
        }
    };

    const OpenPopup:OpenPopupT = async (key, URL, closeOnSuccess) => {
        if (!currentPopups.current[URL]) {
            currentPopups.current[URL] = new Promise<boolean>((resolve) => {
                const popup = window.open(URL, key, "width=500,height=750,popup")
                if (!popup) {
                    SendNotification("Popup blocked, please allow popups for this site.")
                    delete currentPopups.current[URL]
                    resolve(false);
                    return
                }
                let finished = false;
                function onMessage(event: MessageEvent<PopupResponseT>) {
                    if (event.source === popup && event.origin === Origin) {
                        if (event.data && event.data.success) {
                            window.removeEventListener("message", onMessage);
                            finished = true
                            if (closeOnSuccess && window.opener) { // only when current window is auth(signin/signup) popup and that opens sso popup
                                window.opener.postMessage(event.data, window.location.origin);
                                window.close();
                            }
                            if (event.data["modify-auth"]) {
                                const token = event.data.token;
                                const expires = event.data.expires;
                                if (token) AccessToken.current = token
                                if (expires) AccessExpiry.current = new Date(expires)
                                IsLoggedIn.current = !!AccessToken.current
                            }
                            delete currentPopups.current[URL];
                            resolve(true);
                            return
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
    }

    const RequestFulfilledInterceptor = async (config: InternalAxiosRequestConfig): Promise<InternalAxiosRequestConfig> => {
        ValidateConnectionConfig(config)
        // Server not responding
        const gatewayFailures = GetGatewayErrors(config.HostURL)
        if (gatewayFailures > 0) await Sleep(Math.min(1000 * gatewayFailures, 3000))
        if (!config.ConnectivityTestPurpose) {
            while (gatewayFailures > 0) {
                await IsServerOnline(config.HostURL)
            }
        }
        // Rate limited
        if (config.HostURL+config.RemainingPath in RateLimits.current) {
            const retryAfter = RateLimits.current[config.HostURL+config.RemainingPath]
            if (CurrentTime() < retryAfter.getTime()) {
                SendNotification(`Rate limit reached. Please retry after ${Math.trunc((retryAfter.getTime() - CurrentTime()) / 1000)} seconds`)
                return Promise.reject("Rate limited")
            }
        }
        config.headers = config.headers || {};
        if (config.AttachAuth) {
            if (!await EnsureLoggedIn()) return Promise.reject("Auth absent")
            config.headers["authorization"] = AccessToken.current;
        }
        if (config.AttachMFA) {
            const mfa = Cookies.get(MFACookiePath);
            if (!await EnsureLoggedIn() && !mfa && !await PromptMFA()) return Promise.reject("MFA incomplete")
            config.headers["mfa"] = mfa
        }
        if (config.AttachCSRF) {
            const csrf = Cookies.get(CSRFCookiePath);
            if (!csrf && !await PromptLogin()) return Promise.reject("Login incomplete")
            config.headers["csrf"] = csrf
        }
        return config
    }

    const RequestRejectedInterceptor = async (error: AxiosError) => {
        return Promise.reject(error)
    }

    const ResponseFulfilledInterceptor = async (response: AxiosResponse<RawAPIResponseT>): Promise<ProcessedAPIResponseT> => {
        const config = response.config;
        ValidateConnectionConfig(config)
        const data = response.data;
        const status = response.status;

        if (data?.notifications) {
            data.notifications.forEach((notification) =>
                SendNotification(notification)
            );
        }

        if (status === 200) {
            ResetGatewayErrors(config.HostURL);
            if (data["modify-auth"]) {
                AccessToken.current = data["new-token"];
                AccessExpiry.current = new Date(data.reply);
                IsLoggedIn.current = !!AccessToken.current;
            }
            if (config.CloseIfPopup && window.opener && data.success) { // current window can be any site and that opens login / MFA popup
                const re:PopupResponseT = {success: data.success, "modify-auth": data["modify-auth"], token: AccessToken.current, expires: AccessExpiry.current.toISOString()};
                window.opener.postMessage(re, window.location.origin);
                window.close();
            }
        }
        return {success: data.success, reply: data.reply}
    };

    const ResponseRejectedInterceptor = async (error: AxiosError<RawAPIResponseT>) => {
        const response = error.response;
        const config = response?.config;
        const data = response?.data;
        const status = response?.status;
        if (data && data.notifications) data.notifications.forEach((notification) => SendNotification(notification))
        if (!config) return Promise.reject(error)
        ValidateConnectionConfig(config)

        // Not logged in or
        // Action not allowed (lacks permission)
        if (status === 401) {
            if (!config.AttachAuth && !config.AuthRefreshPurpose) { // Auth was not attached and is not for refresh
                SendNotification("Retrying with authentication. Please report this incident to admin")
                config.AttachAuth = true
                return await RetryRequest(config)
            }
            if (config.AttachAuth) { // Auth was attached and still failed
                if (!config.CausedRefresh) { // Server rejected current access token
                    IsLoggedIn.current = false
                    if (await RefreshToken()) { // Retry after refresh
                        config.CausedRefresh = true
                        return await RetryRequest(config)
                    }
                    SendNotification("You need to be logged in to a valid account to perform this action.")
                    return Promise.reject("Not logged in/Session expired/revoked")
                }
                // Server rejected even after refresh
                SendNotification("You do not have enough permissions to perform this action.")
                navigate("/", {replace:true})
                return Promise.reject("Invalid permissions")
            }
            if (config.AuthRefreshPurpose) {
                const loggedIn = await PromptLogin()
                if (loggedIn) {
                    return loggedIn
                }
                if (!config.AttachAuth) {
                    SendNotification("You are not logged in. Please login and try again.")
                }
                return Promise.reject("Session expired/revoked")
            }
        }

        // Mfa required
        else if (status === 403) {
            if (!config.AttachMFA) {
                config.AttachMFA = true;
                return await RetryRequest(config)
            }
        }

        // Incomplete form/parameters
        else if (status === 422) {
            SendNotification("Frontend has errors, please refresh and retry or report this to admin.")
            return Promise.reject("Frontend Errors")
        }

        // Rate limited
        else if (status === 429) {
            const retryAfterRaw = data?.["retry-after"]
            if (!retryAfterRaw) return Promise.reject("Rate limited")
            const retryAfter = new Date(retryAfterRaw)
            RateLimits.current[config.HostURL+config.RemainingPath] = retryAfter
            SendNotification(`Rate limit reached. Please retry after ${Math.trunc((retryAfter.getTime() - CurrentTime()) / 1000)} seconds.`)
            return Promise.reject("Rate limited")
        }

        // Server internal error
        else if (status === 500) {
            return Promise.reject("Server error")
        }

        // Server unreachable
        else if (status === 502 || status === 504) {
            IncrementGatewayErrors(config.HostURL)
            return await RetryRequest(config)
        }

        // Anything else
        else {
            await Sleep(1000)
            return await RetryRequest(config)
        }
    }

    const ResponseFulfilledInterceptorCompat = ResponseFulfilledInterceptor as unknown as (response: AxiosResponse) => AxiosResponse | Promise<AxiosResponse>

    const refreshCredentialConnection = axios.create({withCredentials:true})
    refreshCredentialConnection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor)
    refreshCredentialConnection.interceptors.response.use(ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor)

    const cookieDisabledConnection = axios.create({withCredentials:false})
    cookieDisabledConnection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor)
    cookieDisabledConnection.interceptors.response.use(ResponseFulfilledInterceptorCompat, ResponseRejectedInterceptor)

    return (<Context.Provider value={{SendGet, SendPost, OpenPopup, Logout, EnsureLoggedIn}}>
        {children}
    </Context.Provider>)
}

export default function ConnectionManager() {
    const context = useContext(Context);
    if (context === undefined) {
        throw new Error('ConnectionManager() must be used within a ConnectionContext');
    }
    return context
}
