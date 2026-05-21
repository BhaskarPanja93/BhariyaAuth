const env = import.meta.env;

export const BetaFrontend = env.VITE_BETA_FRONTEND === "true";
export const BetaAPI = env.VITE_BETA_API === "true";
export const BetaWS = env.VITE_BETA_WS === "true";

export const Domain = (env.VITE_AUTH_DOMAIN as string | undefined)?.trim() || "bhariya.ddns.net";
export const Origin = (env.VITE_AUTH_ORIGIN as string | undefined)?.trim() || `https://${Domain}`;
export const Purpose = (env.VITE_AUTH_PURPOSE as string | undefined)?.trim() || "/auth";
export const PurposeFull = `${Origin}${Purpose}`;

const FrontendPrefix = "";
const FrontendSuffix = BetaFrontend ? "/beta" : "";

const APIPrefix = "/api";
const APISuffix = BetaAPI ? "/beta" : "";

const WSPrefix = "/ws";
const WSSuffix = BetaWS ? "/beta" : "";

export const FrontendRoute = `${PurposeFull}${FrontendPrefix}${FrontendSuffix}`;
export const APIRoute = `${PurposeFull}${APIPrefix}${APISuffix}`;
export const WSRoute = `${PurposeFull}${WSPrefix}${WSSuffix}`;

export const CSRFCookiePath = "csrf";
export const MFACookiePath = "mfa";

