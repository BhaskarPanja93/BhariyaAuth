import React, {createContext, useContext, useEffect, useState} from "react";
import {SetFavicon} from "../Utils/Favicon.js";
import FaviconDarkMode from "../Favicon/DarkMode.png"
import FaviconLightMode from "../Favicon/LightMode.png"

/**
 * @typedef {Object} DarkModeContextType
 * @property {boolean} isDarkMode
 */

/**@type {import('react').Context<DarkModeContextType | null>} */
const DarkModeContext = createContext(null)

export default function DarkModeProvider({children}) {
    const [isDarkMode, setIsDarkMode] = useState(false);
    useEffect(() => {
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        const handler = (event) => {
            setIsDarkMode(event.matches);
            SetFavicon(event.matches ? FaviconDarkMode : FaviconLightMode)
        };
        mediaQuery.addEventListener('change', handler);
        handler(mediaQuery);
        return () => mediaQuery.removeEventListener('change', handler);
    }, []);

    return (
        <DarkModeContext.Provider value={{isDarkMode}}>
            {children}
        </DarkModeContext.Provider>
    );
}

export const FetchDarkModeManager = () => {
    const context = useContext(DarkModeContext);
    if (context === undefined) {
        throw new Error('FetchDarkModeManager() must be used within a DarkModeProvider');
    }
    return context;
};
