import React, {createContext, useContext, useEffect, useRef, useState} from "react";

/**
 * @typedef {Object} NotificationContextType
 * @property {(message: string) => void} SendNotification
 */

/**@type {import('react').Context<NotificationContextType | null>} */
const NotificationContext = createContext(null)

export default function NotificationProvider({children}) {
    const [notifications, setNotifications] = useState([]);
    const nextId = useRef(0);

    const SendNotification = (message) => {
        const id = nextId.current++;
        setNotifications(prev => {
            let updated = [...prev, { id, message }];
            if (updated.length > 10) {
                updated = updated.slice(updated.length - 10);
            }
            return updated;
        });
        setTimeout(() => {
            setNotifications(prev => prev.filter(n => n.id !== id));
        }, 7000);
    };

    const removeNotification = (id) => {
        setNotifications(prev => prev.filter(n => n.id !== id));
    };

    return (
        <NotificationContext.Provider value={{SendNotification}}>
            <div className="fixed top-4 space-y-2 flex flex-col"
                 style={{ zIndex: 50 }}>
                {notifications.map((notification) => (
                    <div
                        key={notification.id}
                        onClick={() => removeNotification(notification.id)}
                        className="text-sm bg-white border shadow-xl rounded-xl px-4 py-2 text-gray-800 transition-opacity duration-300 w-fit inline-block cursor-pointer"
                        style={{ border: "4px solid #a855f7" }}
                    >
                        {notification.message}
                    </div>
                ))}
            </div>

            {children}
        </NotificationContext.Provider>
    );
}


export const FetchNotificationManager = () => {
    const context = useContext(NotificationContext);
    if (context === undefined) {
        throw new Error('FetchNotificationManager() must be used within a NotificationProvider');
    }
    return context;
};
