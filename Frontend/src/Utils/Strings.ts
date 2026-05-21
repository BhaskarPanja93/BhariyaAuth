const emailRegex = /^(?:[a-zA-Z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_`{|}~-]+)*|"(?:[\x20-\x7E]|\\[\x20-\x7E])*")@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;

export function PasswordIsStrong(password: string | never): boolean {
    if (typeof password != "string" || password.length < 8 || password.length > 72) {
        return false;
    }
    let hasUpper = false;
    let hasLower = false;
    let hasDigit = false;
    for (const ch of password) {
        if (/[A-Z]/.test(ch)) {
            hasUpper = true;
        } else if (/[a-z]/.test(ch)) {
            hasLower = true;
        } else if (/[0-9]/.test(ch)) {
            hasDigit = true;
        }
        if (hasUpper && hasLower && hasDigit) {
            return true;
        }
    }
    return hasUpper && hasLower && hasDigit;
}

export function EmailIsValid(email: string | never) {
    return typeof email == "string" && emailRegex.test(email);
}

export function NameIsValid(name: string | never) {
    return typeof name == "string" && name.length > 2 && name.length <= 50;
}

export function OTPIsValid(otp: string | never) {
    return typeof otp == "string" && otp.length === 6;
}


