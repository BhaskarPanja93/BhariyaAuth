export const BetaFrontend = false;
export const BetaAPI = false;
export const BetaWS = false;

export const Domain = "bhariya.ddns.net";
export const Origin = `https://${Domain}`;
export const Purpose = "/auth";
export const PurposeFull = `${Origin}${Purpose}`;

const FrontendPrefix = "";
const FrontendSuffix = BetaFrontend ? "/beta" : "";

const APIPrefix = "/api";
const APISuffix = BetaAPI ? "/beta" : "";

const WSPrefix = "/ws";
const WSSuffix = BetaWS ? "/beta" : "";

export const FrontendRoute =
    `${PurposeFull}${FrontendPrefix}${FrontendSuffix}`;

export const APIRoute =
    `${PurposeFull}${APIPrefix}${APISuffix}`;

export const WSRoute =
    `${PurposeFull}${WSPrefix}${WSSuffix}`;

export const CSRFCookiePath = "csrf"
export const MFACookiePath = "mfa"
