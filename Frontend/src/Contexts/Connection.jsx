import React, {createContext, useContext} from 'react';
import axios, {AxiosHeaders} from "axios";
import Cookies from "js-cookie";
import {Sleep} from "../Utils/Sleep.js";
import {FetchNotificationManager} from "./Notification.jsx";

/**
 * @typedef {Object} ConnectionContextType
 * @property {import("axios").AxiosInstance} publicAPI
 * @property {import("axios").AxiosInstance} privateAPI
 */

/** @type {import('react').Context<ConnectionContextType | null>} */
const ConnectionContext = createContext(null);
export const ConnectionProvider = ({children}) => {
    let AccessToken = ""
    const {SendNotification} = FetchNotificationManager();
    const GatewayErrors = {}
    const ServerInternalErrors = {}

    const GetGatewayErrors = (host) => {
        return GatewayErrors[host] || 0
    }
    const GetServerInternalErrors = (host) => {
        return ServerInternalErrors[host] || 0
    }
    const ResetGatewayErrors = (host) => {
        GatewayErrors[host] = 0
    }
    const ResetServerInternalErrors = (host) => {
        ServerInternalErrors[host] = 0
    }
    const IncrementGatewayErrors = (host) => {
        if (GatewayErrors[host] != null) GatewayErrors[host]++
        else GatewayErrors[host] = 1
    }
    const IncrementServerInternalErrors = (host) => {
        if (ServerInternalErrors[host] != null) ServerInternalErrors[host]++
        else ServerInternalErrors[host] = 1
    }

    const RetryRequest = async (connection, config) => {
        try {
            return await connection(config);
        } catch (error) {
            return Promise.reject(error);
        }
    };

    const currentPings = {};
    const IsServerOnline = (host) => {
        if (currentPings[host] != null) return currentPings[host];
        SendNotification("Pinging:" + host)
        currentPings[host] = new Promise((resolve, _) => {
            privateAPI.get(BackendURL+"/status/ping", {
                forServerConnectionCheck: true
            }).then(() => {
                resolve(true)
            }).catch(() => {
                resolve(false)
            }).finally(() => {
                currentPings[host] = null;
            });
        });
        return currentPings[host];
    }

    let currentRefresh = null;
    const RefreshToken = (skipLogin) => { // Create and return a new promise that resolves when the token is refreshed or fails resolving to a boolean.
        if (currentRefresh) return currentRefresh
        const currentCSRF = Cookies.get(CSRFPath)
        if (!currentCSRF) return Promise.reject()
        currentRefresh = new Promise((resolve, _) => {
            axios.get(BackendURL + "/account/refresh", {
                requiresCSRF: true,
                forTokenRefresh: true,
                skipLogin: skipLogin,
                headers: new AxiosHeaders({ CSRF: currentCSRF })
            }).then(()=>{
                SendNotification("Access refreshed..")
                resolve(true)
            }).catch(() => {
                if (!skipLogin) SendNotification("Unable to authenticate you. Please refresh tab")
                resolve(false)
            }).finally(() => {
                currentRefresh = null;
            });
        });
        return currentRefresh;
    }

    const RequestFulfilledInterceptor = async (config) => {
        const url = new URL(config.url, config.baseURL)
        config.host = url.host;
        config.pathname = url.pathname;

        if (isNaN(config.gatewayErrorCount)) config.gatewayErrorCount = 0
        if (isNaN(config.serverInternalErrorCount)) config.serverInternalErrorCount = 0
        if (isNaN(config.serverInternalErrorRetryLimit)) config.serverInternalErrorRetryLimit = 3

        // repeat requests speed limiter
        const gatewayFailures = GetGatewayErrors(config.host)
        const serverInternalFailures = GetServerInternalErrors(config.pathname)
        if (gatewayFailures > 0) await Sleep(Math.min(1000 * gatewayFailures, 3000))
        if (serverInternalFailures > 0) await Sleep(Math.min(1000 * serverInternalFailures, 3000))

        // for server gateway failures, wait for the server to be back online, except serverActiveCheck requests
        if (!config.forServerConnectionCheck) {
            while (gatewayFailures > 0) {
                SendNotification("Unable to reach server, waiting for reconnection..")
                await IsServerOnline(config.host)
            }
            // Attach access token to request if it exists
            if (AccessToken !== "") config.headers["Authorization"] = AccessToken;
            if (config.requiresCSRF) config.headers["CSRF"] = Cookies.get(CSRFPath);
            if (config.requiresMFA) config.headers["MFA"] = Cookies.get(MFAPath);
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
                    AccessToken = data["new-token"]
                }
            }
        }
        return Promise.resolve(response) // TODO: check if it needs only response or Promise.resolve(response)
    }

    const ResponseRejectedInterceptor = async (response) => {
        const config = response.config;
        const data = response.data;
        const status = response.status;
        if (data["notifications"]) data["notifications"].forEach((notification) => SendNotification(notification))

        if (status === 401) {  }
        else if (status === 403) {  }
        else if (status === 422) {  }
        else if (status === 429) {  }
        else if (status === 500) {  }
        else {  }
    }


    const publicAPI = axios.create();
    const privateAPI = axios.create({withCredentials: true});
    privateAPI.interceptors.request.use(RequestFulfilledInterceptor, RequestRejectedInterceptor)
    privateAPI.interceptors.response.use(ResponseFulfilledInterceptor, ResponseRejectedInterceptor)

    return (<ConnectionContext.Provider value={{publicAPI, privateAPI}}>
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
