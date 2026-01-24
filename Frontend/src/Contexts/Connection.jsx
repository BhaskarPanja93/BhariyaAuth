import React, {createContext, useContext, useRef} from 'react';
import axios from "axios";
import Cookies from "js-cookie";
import {Sleep} from "../Utils/Sleep.js";
import {FetchNotificationManager} from "./Notification.jsx";
import {AuthBackendURL, AuthFrontendURL, CSRFCookiePath, MFACookiePath, Origin} from "../Values/Constants.js";

/** @type {React.Context<ConnectionContextType | null>} */
const ConnectionContext = createContext(null);
export default function ConnectionProvider ({children}) {
    const {SendNotification} = FetchNotificationManager();
    const AccessToken = useRef("")
    const AccessExpiry = useRef(new Date())
    const IsLoggedIn = useRef(false)

    /** @type {RefObject<Record<string,number>>} */
    const GatewayErrors = useRef({})
    /** @type {RefObject<Record<string,Promise<boolean>>>} */
    const currentPopups = useRef({})
    /** @type {RefObject<Record<string,Promise<boolean>>>} */
    const currentPings = useRef({});
    /** @type {RefObject<Promise<boolean>>} */
    const currentRefreshes = useRef(null);
    /** @type {RefObject<Promise<boolean>>} */
    const currentLogout = useRef(null);
    /** @type {RefObject<Record<string, WebSocketWriter>>} */
    const webSockets = useRef({});

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

    /** @type SendGetT */
    const SendGet = async (attachCreds, backendURL, remainingPath, config) => {
        if (config == null)
            config = {}
        config.backendURL = backendURL;
        config.attachCreds = attachCreds;
        return connection.get(backendURL+remainingPath, config)
    }

    /** @type SendPostT */
    const SendPost = async (attachCreds, backendURL, remainingPath, data, config) => {
        if (config == null)
            config = {}
        config.backendURL = backendURL;
        config.attachCreds = attachCreds;
        return connection.post(backendURL+remainingPath, data, config)
    }

    /** @type EnsureLoggedInT */
    const EnsureLoggedIn = async () => {
        return (IsLoggedIn.current && AccessExpiry.current && (AccessExpiry.current.getTime() - Date.now()) > 0) || await RefreshToken() || await OpenPopup(AuthFrontendURL+"/login")
    }

    /** @type OpenPopupT */
    const OpenPopup = async (URL) => {
        if (currentPopups.current[URL] == null) {
            currentPopups.current[URL] = new Promise(async (resolve) => {
                const popup = window.open(
                    URL,
                    URL,
                    "width=500,height=750,popup"
                )
                if (!popup) {
                    delete currentPopups.current[URL]
                    return resolve(false);
                }
                let finished = false;
                function onMessage(event) {
                    if (event.source === popup && event.origin === Origin) {
                        if (event.data && event.data.success) {
                            window.removeEventListener("message", onMessage);
                            finished = true
                            if (event.data["token"]) AccessToken.current = event.data["token"]
                            if (event.data["token"]) AccessExpiry.current = event.data["expires"]
                            delete currentPopups.current[URL];
                            if (window.opener) {
                                window.opener.postMessage({success: true, token: event.data["token"], expires: event.data["expires"]}, window.location.origin);
                                window.close();
                            }
                            return resolve(true);
                        }
                    }
                }
                window.addEventListener("message", onMessage);
                while (!popup.closed) await Sleep(100)
                if (!finished) {
                    delete currentPopups.current[URL];
                    window.removeEventListener("message", onMessage);
                    return resolve(false);
                }
            });
        }
        return currentPopups.current[URL];
    }

    /** @type LogoutT */
    const Logout = async () => {
        if (currentLogout.current == null) {
            currentLogout.current = new Promise((resolve) => {
                SendPost(false, AuthBackendURL, "/account/logout", null, {
                        requiresCSRF: true,
                        changesAuth: true,
                    })
                    .then(() => {
                        SendNotification("Logged out");
                        resolve(true);
                    })
                    .catch(() => {
                        resolve(false);
                    })
                    .finally(() => {
                        currentLogout.current = null;
                    });
            })
        }
        return currentLogout.current
    }

    const IsServerOnline = async (host) => {
        if (currentPings.current[host] == null) {
            currentPings.current[host] = new Promise((resolve) => {
                SendGet(false, host, "/status/ping", {
                    forConnectivityCheck: true
                }).then(() => {
                    resolve(true)
                }).catch(() => {
                    resolve(false)
                }).finally(async () => {
                    delete currentPings.current[host];
                })
            })
        }
        return currentPings.current[host];
    }

    const RefreshToken = async () => {
        if (currentRefreshes.current == null) {
            currentRefreshes.current = new Promise((resolve) => {
                SendPost(true, AuthBackendURL, "/account/refresh", null, {
                    requiresCSRF: true,
                    forTokenRefresh: true,
                    allowAccessChange: true,
                }).then(() => {
                    resolve(true)
                }).catch(() => {
                    resolve(false)
                }).finally(() => {
                    currentRefreshes.current = null;
                });
            })
        }
        return currentRefreshes.current;
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
        if (!config.forConnectivityCheck) {
            while (gatewayFailures > 0) {
                await IsServerOnline(config.backendURL)
            }
        }
        if (config.attachCreds) {
            if (AccessToken.current !== "")
                config.headers["authorization"] = AccessToken.current;
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
            if (config.forMFA && window.opener && data["success"]) {
                window.opener.postMessage({success: true}, window.location.origin);
                window.close();
            } else if (config.allowAccessChange && data["modify-auth"]) {
                AccessToken.current = data["new-token"]
                AccessExpiry.current = new Date(data["reply"])
                Sleep(AccessExpiry.current.getTime() - Date.now() - 30*1000).then(()=>{
                    if (Object.values(webSockets.current).find(webSocket => webSocket.state() === WebSocket.OPEN && webSocket.authRequired))
                        RefreshToken();
                })
                if (!config.forTokenRefresh && window.opener) {
                    window.opener.postMessage({success: true, token: AccessToken.current, expiry: AccessExpiry.current}, window.location.origin);
                    window.close();
                }
                IsLoggedIn.current = !!AccessToken.current;
                Object.values(webSockets.current).forEach(websocket => {
                    websocket.authenticateIfRequired()
                });
            }
        }

        return Promise.resolve({success: data.success, reply: data.reply})
    }

    const ResponseRejectedInterceptor = async (error) => {
        const response = error.response;
        const config = response.config;
        const data = response.data;
        const status = response.status;
        if (data && data["notifications"]) data["notifications"].forEach((notification) => SendNotification(notification))

        // Not authenticated
        if (status === 401) {
            if (config.attachCreds)
                return Promise.reject("Authentication required")
            IsLoggedIn.current = false
            if (!config.forTokenRefresh) {
                if (await EnsureLoggedIn())
                    return await RetryRequest(connection, config)
                else
                    return Promise.reject("Authentication stopped")
            } else
                return Promise.reject("Not authenticated")
        }

        // Incomplete authentication (Mfa required)
        else if (status === 403) {
            if (!config.attachCreds)
                return Promise.reject("Authentication required")
            if (!config.requiresMFA) {
                config.requiresMFA = true;
                return await RetryRequest(connection, config)
            } else {
                if (await OpenPopup(AuthFrontendURL+"/mfa"))
                    return await RetryRequest(connection, config)
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
            SendNotification(`Too many requests, retrying automatically after ${retryAfter} seconds`)
            if (!retryAfter || isNaN(retryAfter)) retryAfter = 1
            await Sleep(retryAfter * 1000)
            return await RetryRequest(connection, config)
        }

        // Server internal error
        else if (status === 500) {
            if (isNaN(config.serverErrorCount))
                config.serverErrorCount = 0
            config.serverErrorCount += 1
            if (config.serverErrorCount < 3) {
                await Sleep(config.serverErrorCount * 1000)
                return await RetryRequest(connection, config)
            }
            return Promise.reject("Server error")
        }

        // Server unreachable
        else if (status === 502 || status === 504) {
            IncrementGatewayErrors(config.host)
            return await RetryRequest(connection, config)
        }

        // Anything else
        else {
            await Sleep(1000)
            return await RetryRequest(connection, config)
        }
    }

    const connection = axios.create({withCredentials: true});
    connection.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor)
    connection.interceptors.response.use(ResponseFulfilledInterceptor, ResponseRejectedInterceptor)


    /** @type GetWebSocketT */
    const GetWebSocket = async (URL, withCredentials) => {
        const existing = webSockets.current[URL];
        if (existing) {
            const state = existing.state();
            if (state === WebSocket.OPEN) {
                return existing;
            }
            if (state === WebSocket.CONNECTING) {
                return existing.openPromise;
            }
            delete webSockets.current[URL];
        }
        webSockets.current[URL] = new WebSocketWriter(URL, withCredentials ? AccessToken : null, () => TryCloseWebSocket(URL))
        return webSockets.current[URL].openPromise;
    }

    /** @type TryCloseWebSocketT */
    const TryCloseWebSocket = async (URL) => {
        const existing = webSockets.current[URL];
        if (existing) {
            let lastUsed = existing.lastUsed
            while (true) {
                if (existing.lastUsed !== lastUsed)
                    return false
                let idleFor = Date.now() - existing.lastUsed
                if (idleFor < 5000) {
                    await Sleep(5000 - idleFor)
                    continue
                }
                delete webSockets.current[URL]
                existing.ws.close()
                return true
            }
        }
        return true
    }

    return (<ConnectionContext.Provider value={{SendGet, SendPost, GetWebSocket, TryCloseWebSocket, OpenPopup, Logout, EnsureLoggedIn}}>
        {children}
    </ConnectionContext.Provider>)
}

export const FetchConnectionManager = () => {
    const context = useContext(ConnectionContext);
    if (context === undefined) {
        throw new Error('FetchConnectionManager() must be used within a ConnectionProvider');
    }
    return context
}
