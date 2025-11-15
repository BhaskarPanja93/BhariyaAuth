import {Suspense} from "react";
import {Outlet} from "react-router-dom";

export default function PageStructure() {
    return (
        <Suspense fallback={null}>
            <Outlet/>
        </Suspense>
    )
}