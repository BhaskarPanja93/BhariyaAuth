import {type Context, createContext, type ReactNode, useContext, useRef, useState} from "react";
import {Sleep} from "../Utils/Time.ts";

type SendNotificationT = (message: string) => void

interface NotificationContextType {
    SendNotification: SendNotificationT;
}

const Context = createContext<NotificationContextType | undefined>(undefined)

export function NotificationContext({children}: { children: ReactNode }) {
    const [notifications, setNotifications] = useState<Array<{ id: number, message: string }>>([]);
    const nextId = useRef(0);

    const SendNotification: SendNotificationT = (message) => {
        const id = nextId.current++;
        setNotifications(prev => {
            let updated = [...prev, {id, message}];
            if (updated.length > 10) {
                updated = updated.slice(updated.length - 10);
            }
            return updated;
        });
        Sleep(7000).then(() => removeNotification(id));
    };

    const removeNotification = (id: number) => {
        setNotifications(prev => prev.filter(n => n.id !== id));
    };

    return (<Context.Provider value={{SendNotification}}>
        <div className="fixed top-4 space-y-2 flex flex-col"
             style={{zIndex: 50}}>
            {notifications.map((notification) => (<div
                key={notification.id}
                onClick={() => removeNotification(notification.id)}
                className="select-none text-sm bg-white border shadow-xl rounded-xl px-4 py-2 text-gray-800 transition-opacity duration-300 w-fit inline-block cursor-pointer"
                style={{border: "4px solid #a855f7"}}
            >
                {notification.message}
            </div>))}
        </div>
        {children}
    </Context.Provider>);
}

export default function NotificationManager() {
    const context = useContext(Context);
    if (context === undefined) {
        throw new Error('NotificationManager() must be used within a NotificationContext');
    }
    return context;
};
