export function Sleep(milliseconds: number) {
    return new Promise<void>(resolve =>
        setTimeout(resolve, Math.max(0, milliseconds))
    );
}

export function CurrentTime() {
    return Date.now()
}

