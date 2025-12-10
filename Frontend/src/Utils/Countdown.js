import {Sleep} from "./Sleep.js";

export const Countdown = async (seconds, currentCountdownIDRef, setValue) => {
    let dummyValue = seconds
    if (isNaN(seconds)) return
    const newID = currentCountdownIDRef.current + 1
    currentCountdownIDRef.current = newID
    await Sleep(100)
    setValue(seconds)
    while (newID === currentCountdownIDRef.current && dummyValue !== 0) {
        let newTime = dummyValue-0.1
        if (newTime < 0)
            newTime = 0
        setValue(newTime)
        dummyValue = newTime
        await Sleep(100)
    }
}