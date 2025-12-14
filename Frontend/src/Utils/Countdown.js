import {Sleep} from "./Sleep.js";

export const Countdown = async (duration, interval, currentCountdownIDRef, setter) => {
    let dummyValue = duration
    if (isNaN(duration)) return
    const newID = currentCountdownIDRef.current + 1
    currentCountdownIDRef.current = newID
    await Sleep(interval)
    setter(duration)
    while (newID === currentCountdownIDRef.current && dummyValue !== 0) {
        let newTime = dummyValue-interval
        if (newTime < 0)
            newTime = 0
        setter(newTime)
        dummyValue = newTime
        await Sleep(interval)
    }
}