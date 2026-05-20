import {BrowserRouter} from 'react-router';
import {createRoot} from 'react-dom/client'
import {DarkModeContext} from './Contexts/DarkMode'
import {NotificationContext} from './Contexts/Notification'
import {ConnectionContext} from './Contexts/Connection'
import Router from "./Structures/Router.tsx";

createRoot(document.getElementById('root')!)
    .render(
        <BrowserRouter basename="/auth">
            <DarkModeContext>
                <NotificationContext>
                    <ConnectionContext>
                        <Router/>
                    </ConnectionContext>
                </NotificationContext>
            </DarkModeContext>
        </BrowserRouter>
    );


