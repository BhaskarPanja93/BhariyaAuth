import {Sleep} from "./Sleep.js";

export const Countdown = async (durationS, intervalS, currentCountdownIDRef, setter) => {
    if (isNaN(durationS)) return
    let dummyValue = durationS
    const newID = ++currentCountdownIDRef.current
    await Sleep(intervalS*1000)
    setter(durationS)
    while (newID === currentCountdownIDRef.current && dummyValue !== 0) {
        let newTime = dummyValue-intervalS
        if (newTime < 0)
            newTime = 0
        setter(newTime)
        dummyValue = newTime
        await Sleep(intervalS*1000)
    }
}