import {createContext, type ReactNode, useContext, useEffect, useState} from "react";
import {Favicon} from "../Utils/Favicon";
import {FrontendRoute} from "../Values/Constants";

interface DarkModeContextType {
    IsDarkMode: boolean;
}

const context = createContext<DarkModeContextType | undefined>(undefined);

export function DarkModeContext({children}: { children: ReactNode }) {
    const [IsDarkMode, setIsDarkMode] = useState<boolean>(false);

    useEffect(() => {
        const applyMode = (isDarkMode: boolean): void => {
            setIsDarkMode(isDarkMode);
            Favicon(isDarkMode ? FrontendRoute + "/favicons/DarkMode.png" : FrontendRoute + "/favicons/LightMode.png");
        };
        const handleChange = (event: MediaQueryListEvent): void => {
            applyMode(event.matches);
        };
        const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
        applyMode(mediaQuery.matches);
        mediaQuery.addEventListener("change", handleChange);
        return () => mediaQuery.removeEventListener("change", handleChange);
    }, []);

    return (<context.Provider value={{IsDarkMode}}>
        {children}
    </context.Provider>)
}

export default function DarkModeManager(): DarkModeContextType {
    const ctx = useContext(context);
    if (ctx === undefined) {
        throw new Error('DarkModeManager() must be used within a DarkModeContext');
    }
    return ctx;
};


