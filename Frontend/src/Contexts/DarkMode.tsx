import {createContext, type ReactNode, useContext, useEffect, useState} from "react";
import {Favicon} from "../Utils/Favicon";
import {FrontendRoute} from "../Values/Constants";

interface DarkModeContextType {
    IsDarkMode: boolean;
}

const Context = createContext<DarkModeContextType | undefined>(undefined);

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

    return (<Context.Provider value={{IsDarkMode}}>
        {children}
    </Context.Provider>)
}

export default function DarkModeManager(): DarkModeContextType {
    const context = useContext(Context);
    if (context === undefined) {
        throw new Error('DarkModeManager() must be used within a DarkModeContext');
    }
    return context;
};


