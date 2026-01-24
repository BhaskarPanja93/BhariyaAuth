/*
  ==========================================
  Connection.jsx
  ==========================================
*/

/**
 * @typedef {Object} ConnectionContextType
 * @property {SendGetT} SendGet
 * @property {SendPostT} SendPost
 * @property {GetWebSocketT} GetWebSocket
 * @property {TryCloseWebSocketT} TryCloseWebSocket
 * @property {OpenPopupT} OpenPopup
 * @property {LogoutT} Logout
 * @property {EnsureLoggedInT} EnsureLoggedIn
 */

/**
 * @typedef {
 * (
 *     URL: string,
 *     withCredentials: boolean
 * ) =>
 *     Promise<WebSocketWriter>
 * } GetWebSocketT
 */

/**
 * @typedef {
 * (
 *     URL: string,
 * ) =>
 *     Promise<boolean>
 * } TryCloseWebSocketT
 */

/**
 * @typedef {
 * (
 *     attachCreds: boolean,
 *     backendURL: string,
 *     remainingPath: string,
 *     config: any,
 * ) =>
 *     Promise<{success: boolean, reply: any}|any>
 * } SendGetT
 */

/**
 * @typedef {
 * (
 *     attachCreds: boolean,
 *     backendURL: string,
 *     remainingPath: string,
 *     data: FormData,
 *     config: any,
 * ) =>
 *     Promise<{success: boolean, reply: any}|any>
 * } SendPostT
 */

/**
 * @typedef {
 * (
 * ) =>
 *     Promise<boolean>
 * } LogoutT
 */

/**
 * @typedef {
 * (
 * ) =>
 *     Promise<boolean>
 * } EnsureLoggedInT
 */

/**
 * @typedef {
 * (
 *     URL: string
 * ) =>
 *     Promise<boolean>
 * } OpenPopupT
 */

/*
  ==========================================
  DarkMode.jsx
  ==========================================
*/

/**
 * @typedef {Object} DarkModeContextType
 * @property {boolean} isDarkMode
 */

/*
  ==========================================
  Notification.jsx
  ==========================================
*/

/**
 * @typedef {Object} NotificationContextType
 * @property {SendNotificationT} SendNotification
 */

/**
 * @typedef {
 * (
 *     message: string
 * ) =>
 *     void
 * } SendNotificationT
 */

/*
  ==========================================
  Countdown.js
  ==========================================
*/

/**
 * @typedef {
 * (
 *     durationS: number,
 *     intervalS: number,
 *     currentCountdownIDRef:  RefObject<any>,
 *     setter: (value: number) => void
 * ) =>
 *     Promise<void>
 * } CountdownT
 */

export {};
