import type {Dispatch, SetStateAction} from "react";
import {Sleep} from "./Time";

type CountdownSetter = Dispatch<SetStateAction<number>>;

export default class Countdown {
    private context: number
    private durationSeconds: number
    private intervalSeconds: number
    private readonly setter: CountdownSetter

    public constructor(durationSeconds: number, intervalSeconds: number, setter: CountdownSetter) {
        this.context = 0
        this.durationSeconds = Countdown.getSafeDuration(durationSeconds)
        this.intervalSeconds = Countdown.getSafeInterval(intervalSeconds)
        this.setter = setter
    }

    private static getSafeDuration(durationSeconds: number): number {
        if (!Number.isFinite(durationSeconds) || durationSeconds < 0) {
            return 0
        }
        return durationSeconds
    }

    private static getSafeInterval(intervalSeconds: number): number {
        if (!Number.isFinite(intervalSeconds) || intervalSeconds <= 0) {
            return 0.1
        }
        return intervalSeconds
    }

    public updateInterval(intervalSeconds: number) {
        this.intervalSeconds = Countdown.getSafeInterval(intervalSeconds)
        return this
    }

    public resetDuration(durationSeconds: number) {
        this.durationSeconds = Countdown.getSafeDuration(durationSeconds)
        return this.start()
    }

    public cancel() {
        ++this.context
        return this
    }

    public start() {
        const context = ++this.context
        void this.startCountdown(context)
        return this
    }

    private async startCountdown(context: number) {
        const startTime = performance.now()
        while (context === this.context) {
            const elapsedSeconds = (performance.now() - startTime) / 1000
            const remainingSeconds = Math.max(0, this.durationSeconds - elapsedSeconds)
            this.setter(remainingSeconds)

            if (remainingSeconds <= 0) {
                return
            }

            await Sleep(1000 * this.intervalSeconds)
        }
    }
}
