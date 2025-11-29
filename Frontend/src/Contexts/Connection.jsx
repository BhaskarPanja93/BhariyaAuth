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
 * @property {(URL: string) => Promise} OpenAuthPopup
 * @property {() => Promise} Logout
 */

/** @type {import('react').Context<ConnectionContextType | null>} */
const ConnectionContext = createContext(null);
export const ConnectionProvider = ({children}) => {
    const {SendNotification} = FetchNotificationManager();
    const AccessToken = useRef("")

    const GatewayErrors = useRef({})
    const GetGatewayErrors = (host) => {
        return GatewayErrors.current[host] || 0
    }
    const ResetGatewayErrors = (host) => {
        GatewayErrors.current[host] = 0
    }
    const IncrementGatewayErrors = (host) => {
        if (GatewayErrors.current[host] != null) GatewayErrors.current[host]++
        else GatewayErrors.current[host] = 1
    }

    const currentAuthPopups = useRef({});
    const OpenAuthPopup = (URL) => {
        if (currentAuthPopups.current[URL]) return currentAuthPopups.current[URL];
        currentAuthPopups.current[URL] = new Promise((resolve, _) => {
            const popup = window.open(URL, "_blank", "width=500,height=600");
            if (!popup) {
                SendNotification("Please allow popups to proceed..")
                return resolve(false);
            }
            let finished = false;
            function onMessage(event) {
                if (event.source === popup && event.origin === FrontendDomain) {
                    finished = true;
                    window.removeEventListener("message", onMessage);
                    popup.close();
                    if (event.data && event.data && event.data.success) {
                        if (event.data["token"]) {
                            AccessToken.current = event.data["token"]
                            if (window.opener) {
                                window.opener.postMessage({ success: true, token: AccessToken.current}, window.location.origin);
                                window.close();
                            }
                            return resolve(true);
                        }
                    }
                }
            }
            window.addEventListener("message", onMessage);
            const timer = setInterval(() => {
                if (popup.closed) {
                    clearInterval(timer);
                    if (!finished) {
                        window.removeEventListener("message", onMessage);
                        if (window.opener) {
                            window.opener.postMessage({ success: false }, window.location.origin);
                            window.close();
                        }
                        resolve(false);
                    }
                }
            }, 300);
        });
        return currentAuthPopups.current[URL];
    }

    const currentMFAPopup = useRef(null);
    const OpenMFAPopup = () => {
        if (currentMFAPopup.current) return currentMFAPopup.current;
        currentMFAPopup.current = new Promise((resolve, _) => {
            const popup = window.open(FrontendURL+"/mfa", "_blank", "width=500,height=600");
            if (!popup) {
                SendNotification("Please allow popups to proceed..")
                return resolve(false);
            }
            let finished = false;
            function onMessage(event) {
                if (event.source === popup && event.origin === FrontendURL) {
                    finished = true;
                    window.removeEventListener("message", onMessage);
                    popup.close();
                    if (event.data && event.data && event.data.success) {
                        return resolve(true);
                    }
                }
            }
            window.addEventListener("message", onMessage);
            const timer = setInterval(() => {
                if (popup.closed) {
                    clearInterval(timer);
                    if (!finished) {
                        window.removeEventListener("message", onMessage);
                        resolve(false);
                    }
                }
            }, 300);
        });
        return currentMFAPopup.current;
    }

    const RetryRequest = async (connection, config) => {
        try {
            return await connection(config);
        } catch (error) {
            return Promise.reject();
        }
    };

    const currentPings = useRef({});
    const IsServerOnline = (host) => {
        if (currentPings.current[host] != null) return currentPings.current[host];
        SendNotification("Pinging:" + host)
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
    const Logout = async () => {
        if (currentLogout.current) return await currentLogout.current
        const currentCSRF = Cookies.get(CSRFCookiePath)
        if (!currentCSRF) return Promise.resolve(false)
        currentLogout.current = new Promise((resolve, _) => {
            axios.post(BackendURL + "/account/logout", {
                requiresCSRF: true,
                forLogout: true,
            }).then(()=>{
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
    const RefreshToken = async (skipLogin) => { // Create and return a new promise that resolves when the token is refreshed or fails resolving to a boolean.
        if (currentRefresh.current) return await currentRefresh.current
        const currentCSRF = Cookies.get(CSRFCookiePath)
        if (!currentCSRF) return Promise.resolve(false)
        currentRefresh.current = new Promise((resolve, _) => {
            axios.get(BackendURL + "/account/refresh", {
                requiresCSRF: true,
                forTokenRefresh: true,
                skipLogin: skipLogin
            }).then(()=>{
                SendNotification("Access refreshed..")
                resolve(true)
            }).catch(() => {
                if (!skipLogin) SendNotification("Unable to authenticate you. Please refresh tab")
                resolve(false)
            }).finally(() => {
                currentRefresh.current = null;
            });
        });
        return currentRefresh.current;
    }

    const RequestFulfilledInterceptor = async (config) => {
        const url = new URL(config.url, config.baseURL)
        config.host = url.host;
        config.pathname = url.pathname;

        // repeat requests speed limiter
        const gatewayFailures = GetGatewayErrors(config.host)
        if (gatewayFailures > 0) await Sleep(Math.min(1000 * gatewayFailures, 3000))

        // for server gateway failures, wait for the server to be back online, except serverActiveCheck requests
        if (!config.forServerConnectionCheck) {
            while (gatewayFailures > 0) {
                SendNotification("Unable to reach server, waiting for reconnection..")
                await IsServerOnline(config.host)
            }
            // Attach access token to request if it exists
            if (AccessToken.current !== "") config.headers["Authorization"] = AccessToken.current;
            if (config.requiresCSRF) {
                config.headers["CSRF"] = Cookies.get(CSRFCookiePath);
                if (!config.headers["CSRF"]) return Promise.reject(config)
            }
            if (config.requiresMFA) {
                config.headers["MFA"] = Cookies.get(MFACookiePath);
                if (!config.headers["MFA"]) return Promise.reject(config)
            }
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
        if (data["notifications"]) data["notifications"].forEach((notification) => SendNotification(notification))

        if (status === 200) {
            ResetGatewayErrors(config.host)
            if (response.forServerConnectionCheck) {
                SendNotification("Reconnected to server..")
            }
            if (response.forTokenRefresh || response.forLogout) {
                if (data["auth-modified"]) {
                    AccessToken.current = data["new-token"]
                }
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
            if (config.forTokenRefresh) {
                if (!config.skipLogin) {
                    if (OpenAuthPopup(FrontendURL+"/login"))
                        return Promise.resolve({success: data.success, reply: data.reply})
                    else
                        return Promise.reject()
                }
            } else {
                if (await RefreshToken(true)) {
                    return await RetryRequest(privateAPI, config)
                } else if (!config.skipLogin) {
                    if (OpenAuthPopup(FrontendURL+"/login"))
                        return await RetryRequest(privateAPI, config)
                    else
                        return Promise.reject()
                }
            }
        }

        // Incomplete authentication (MFA)
        else if (status === 403) {
            if (!config.requiresMFA) {
                config.requiresMFA = true;
                return await RetryRequest(privateAPI, config)
            } else {
                if (await OpenMFAPopup())
                    return await RetryRequest(privateAPI, config)
                else
                    return Promise.reject()
            }
        }

        // Incomplete form/parameters
        else if (status === 422) {
            SendNotification("Frontend has errors, please report this to admin.")
            return Promise.reject()
        }

        // Rate limited
        else if (status === 429) {
            const retryAfter = data["retry-after"]
            if (!retryAfter || isNaN(retryAfter)) retryAfter = 1
            await Sleep(retryAfter * 1000)
            return await RetryRequest(privateAPI, config)
        }

        // Server internal error
        else if (status === 500) {
            if (isNaN(config.serverInternalErrorCount)) {
                config.serverInternalErrorCount = 1
            } else {
                config.serverInternalErrorCount += 1
            }
            if (config.serverInternalErrorCount < 3) {
                await Sleep(config.serverInternalErrorCount * 1000)
                return await RetryRequest(privateAPI, config)
            }
            return Promise.reject()
        }

        // Server unreachable
        else if (status === 502 || status === 504) {
            IncrementGatewayErrors(config.host)
            config.gatewayErrorCount += 1
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

    return (<ConnectionContext.Provider value={{publicAPI, privateAPI, OpenAuthPopup, Logout}}>
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
