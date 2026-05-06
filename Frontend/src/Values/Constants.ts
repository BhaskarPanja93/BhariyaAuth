export const BetaFrontend = false;
export const BetaAPI = false;
export const BetaWS = false;

export const Domain = "bhariya.ddns.net";
export const Origin = `https://${Domain}`;
export const Purpose = "/auth";
export const PurposeFull = `${Origin}${Purpose}`;

export const FrontendPrefix = "";
export const FrontendSuffix = BetaFrontend ? "/beta" : "";

export const APIPrefix = "/api";
export const APISuffix = BetaAPI ? "/beta" : "";

export const WSPrefix = "/ws";
export const WSSuffix = BetaWS ? "/beta" : "";

export const FrontendRoute =
    `${PurposeFull}${FrontendPrefix}${FrontendSuffix}`;

export const APIRoute =
    `${PurposeFull}${APIPrefix}${APISuffix}`;

export const WSRoute =
    `${PurposeFull}${WSPrefix}${WSSuffix}`;

export const CSRFCookiePath = "csrf"
export const MFACookiePath = "mfa"
