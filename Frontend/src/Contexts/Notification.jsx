import React, {createContext, useContext, useEffect, useRef, useState} from "react";

/**
 * @typedef {Object} NotificationContextType
 * @property {(message: string) => void} SendNotification
 */

/**@type {import('react').Context<NotificationContextType | null>} */
const NotificationContext = createContext(null)

export const NotificationProvider = ({children}) => {
    const [notifications, setNotifications] = useState([]);
    const nextId = useRef(0);

    const SendNotification = (message) => {
        setNotifications(prev => {
            let updated = [...prev, {id: nextId.current++, message: message}];
            if (updated.length > 10) {
                updated = updated.slice(updated.length - 10);
            }
            return updated;
        });
    };

    useEffect(() => {
        const timers = notifications.map((notification) => setTimeout(() => {
            setNotifications(prev => prev.filter(n => n.id !== notification.id));
        }, 7000));
        return () => {
            timers.forEach(clearTimeout);
        };
    }, [notifications]);

    return (<NotificationContext.Provider value={{SendNotification}}>
        <div className="fixed top-4 space-y-2"
             style={{zIndex: 50}}>
            {notifications.map((notification) => (
                <div key={notification.id}
                     style={{border: "4px solid #ffc300"}}
                     className="bg-white border shadow-xl rounded-xl px-4 py-2 text-gray-800 transition-opacity duration-300">
                    {notification.message}
                </div>)
            )}
        </div>
        {children}
    </NotificationContext.Provider>);
};

export const FetchNotificationManager = () => {
    const context = useContext(NotificationContext);
    if (context === undefined) {
        throw new Error('FetchNotificationManager() must be used within a NotificationProvider');
    }
    return context;
};
