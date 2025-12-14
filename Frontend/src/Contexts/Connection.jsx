import React, {createContext, useContext, useRef} from 'react';
import axios from "axios";
import Cookies from "js-cookie";
import {Sleep} from "../Utils/Sleep.js";
import {FetchNotificationManager} from "./Notification.jsx";
import {BackendURL, CSRFCookiePath, FrontendDomain, FrontendURL, MFACookiePath} from "../Values/Constants.js";

/**
 * @typedef {Object} ConnectionContextType
 * @property {import("axios").AxiosInstance} publicAPI
 * @property {import("axios").AxiosInstance} privateAPI
 * @property {(string) => Promise<boolean>} OpenPopup
 * @property {() => Promise<boolean>} Logout
 * @property {() => Promise<boolean>} EnsureLoggedIn
 */

/** @type {import('react').Context<ConnectionContextType | null>} */
const ConnectionContext = createContext(null);
export default function ConnectionProvider ({children}) {
    const {SendNotification} = FetchNotificationManager();
    const AccessToken = useRef("")
    const IsLoggedIn = useRef(false)

    const GatewayErrors = useRef({})
    const GetGatewayErrors = (host) => {
        return GatewayErrors.current[host] || 0
    }
    const ResetGatewayErrors = (host) => {
        if (GatewayErrors.current[host] !== 0 && GatewayErrors.current[host] !== undefined) {
            SendNotification("Server is back online")
        }
        GatewayErrors.current[host] = 0
    }
    const IncrementGatewayErrors = (host) => {
        if (isNaN(GatewayErrors.current[host]))
            GatewayErrors.current[host] = 0

        if (GatewayErrors.current[host] === 0)
            SendNotification("Server unreachable. Retrying..")
        GatewayErrors.current[host] = 1
    }

    const currentPopup = useRef({})
    const OpenPopup = (URL) => {
        if (currentPopup.current[URL]) return currentPopup.current[URL];
        currentPopup.current[URL] = new Promise(async (resolve, _) => {
            const popup = window.open(
                URL,
                URL,
                "width=500,height=750,menubar=no,toolbar=no,location=no,status=no,resizable=no,scrollbars=no"
            )
            if (!popup) {
                currentPopup.current[URL] = null
                return resolve(false);
            }
            let finished = false;
            function onMessage(event) {
                if (event.source === popup && event.origin === FrontendDomain) {
                    if (event.data && event.data.success) {
                        window.removeEventListener("message", onMessage);
                        finished = true
                        if (event.data["token"]) AccessToken.current = event.data["token"]
                        currentPopup.current[URL] = null;
                        if (window.opener) {
                            window.opener.postMessage({success: true, token: event.data["token"]}, window.location.origin);
                            window.close();
                        }
                        return resolve(true);
                    }
                }
            }
            window.addEventListener("message", onMessage);
            while (!popup.closed) await Sleep(200)
            if (!finished) {
                currentPopup.current[URL] = null;
                window.removeEventListener("message", onMessage);
                return resolve(false);
            }
        });
        return currentPopup.current[URL];
    }

    const currentPings = useRef({});
    const IsServerOnline = (host) => {
        if (currentPings.current[host] != null) return currentPings.current[host];
        currentPings.current[host] = new Promise((resolve, _) => {
            privateAPI.get(BackendURL+"/status/ping", {
                forServerConnectionCheck: true
            }).then(() => {
                resolve(true)
            }).catch(() => {
                resolve(false)
            }).finally(() => {
                currentPings.current[host] = null;
            });
        });
        return currentPings.current[host];
    }

    const currentLogout = useRef(null);
    const Logout = () => {
        if (currentLogout.current) return currentLogout.current
        currentLogout.current = new Promise((resolve, _) => {
            privateAPI.post(BackendURL + "/account/logout", null, {
                requiresCSRF: true,
                forAccessFetch: true,
            }).then(()=>{
                SendNotification("Logged out")
                resolve(true)
            }).catch(() => {
                resolve(false)
            }).finally(() => {
                currentLogout.current = null;
            });
        });
        return currentLogout.current;
    }

    const currentRefresh = useRef(null);
    const RefreshToken = () => { // Create and return a new promise that resolves when the token is refreshed or fails resolving to a boolean.
        if (currentRefresh.current) return currentRefresh.current
        currentRefresh.current = new Promise((resolve, _) => {
            privateAPI.post(BackendURL + "/account/refresh", null, {
                requiresCSRF: true,
                forTokenRefresh: true,
                forAccessFetch: true,
            }).then(()=>{
                resolve(true)
            }).catch(() => {
                resolve(false)
            }).finally(() => {
                currentRefresh.current = null;
            });
        });
        return currentRefresh.current;
    }

    const EnsureLoggedIn = async () => {
        return IsLoggedIn.current || await RefreshToken() || await OpenPopup(FrontendURL+"/login")
    }

    const RetryRequest = async (connection, config) => {
        try {
            return await connection(config);
        } catch (error) {
            return Promise.reject(error);
        }
    };

    const RequestFulfilledInterceptor = async (config) => {
        const url = new URL(config.url, config.baseURL)
        config.host = url.host;
        config.pathname = url.pathname;
        const gatewayFailures = GetGatewayErrors(config.host)
        if (gatewayFailures > 0) await Sleep(Math.min(1000 * gatewayFailures, 3000))
        if (!config.forServerConnectionCheck) {
            while (gatewayFailures > 0) {
                await IsServerOnline(config.host)
            }
        }
        if (AccessToken.current !== "") config.headers["authorization"] = AccessToken.current;
        if (config.requiresMFA) {
            let mfa = Cookies.get(MFACookiePath);
            if (!mfa) return Promise.reject(config)
            config.headers["mfa"] = mfa
        }
        if (config.requiresCSRF) {
            let csrf = Cookies.get(CSRFCookiePath);
            if (!csrf) return Promise.reject(config)
            config.headers["csrf"] = csrf
        }
        return config
    }

    const RequestRejectedInterceptor = async (config) => {
        return Promise.reject(config)
    }

    const ResponseFulfilledInterceptor = async (response) => {
        const config = response.config;
        const data = response.data;
        const status = response.status;
        if (data && data["notifications"]) data["notifications"].forEach((notification) => SendNotification(notification))

        if (status === 200) {
            ResetGatewayErrors(config.host)
            if (config.forAccessFetch && data["modify-auth"]) {
                AccessToken.current = data["new-token"]
                IsLoggedIn.current = !!AccessToken.current;
                if (!config.forTokenRefresh && window.opener) {
                    window.opener.postMessage({success: true, token: AccessToken.current}, window.location.origin);
                    window.close();
                }
            }

            if (config.forMFA && window.opener && data["success"]) {
                window.opener.postMessage({success: true}, window.location.origin);
                window.close();
            }
            return Promise.resolve({success: data.success, reply: data.reply})
        }
    }

    const ResponseRejectedInterceptor = async (error) => {
        const response = error.response;
        const config = response.config;
        const data = response.data;
        const status = response.status;
        if (data && data["notifications"]) data["notifications"].forEach((notification) => SendNotification(notification))

        // Not authenticated
        if (status === 401) {
            IsLoggedIn.current = false
            if (!config.forTokenRefresh) {
                if (await RefreshToken()) {
                    return await RetryRequest(privateAPI, config)
                } else if (!config.skipLogin || AccessToken.current) {
                    if (await OpenPopup(FrontendURL+"/login"))
                        return await RetryRequest(privateAPI, config)
                    else
                        return Promise.reject("Authentication stopped")
                }
            }
            return Promise.reject("Not authenticated")
        }

        // Incomplete authentication (Mfa required)
        else if (status === 403) {
            if (!config.requiresMFA) {
                config.requiresMFA = true;
                return await RetryRequest(privateAPI, config)
            } else {
                if (await OpenPopup(FrontendURL+"/mfa"))
                    return await RetryRequest(privateAPI, config)
                else
                    return Promise.reject("Mfa stopped")
            }
        }

        // Incomplete form/parameters
        else if (status === 422) {
            SendNotification("Frontend has errors, please refresh and retry or report this to admin.")
            return Promise.reject("Frontend Errors")
        }

        // Rate limited
        else if (status === 429) {
            let retryAfter = data["retry-after"]
            if (!retryAfter || isNaN(retryAfter)) retryAfter = 1
            await Sleep(retryAfter * 1000)
            return await RetryRequest(privateAPI, config)
        }

        // Server internal error
        else if (status === 500) {
            if (isNaN(config.serverErrorCount))
                config.serverErrorCount = 0
            config.serverErrorCount += 1
            if (config.serverErrorCount < 3) {
                await Sleep(config.serverErrorCount * 1000)
                return await RetryRequest(privateAPI, config)
            }
            return Promise.reject("Server error")
        }

        // Server unreachable
        else if (status === 502 || status === 504) {
            IncrementGatewayErrors(config.host)
            return await RetryRequest(privateAPI, config)
        }

        // Anything else
        else {
            await Sleep(1000)
            return await RetryRequest(privateAPI, config)
        }
    }

    const publicAPI = axios.create();
    const privateAPI = axios.create({withCredentials: true});
    privateAPI.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor)
    privateAPI.interceptors.response.use(ResponseFulfilledInterceptor, ResponseRejectedInterceptor)

    return (<ConnectionContext.Provider value={{publicAPI, privateAPI, OpenPopup, Logout, EnsureLoggedIn}}>
        {children}
    </ConnectionContext.Provider>);
};

export const FetchConnectionManager = () => {
    const context = useContext(ConnectionContext);
    if (context === undefined) {
        throw new Error('FetchConnectionManager() must be used within a ConnectionProvider');
    }
    return context;
};
