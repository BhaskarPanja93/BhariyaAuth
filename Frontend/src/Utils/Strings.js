const emailRegex = /^(?=.{1,50}$)[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/;
const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,19}$/


export function NameIsValid(name) {
    return name.length > 0 && name.length <= 50;
}

export function EmailIsValid(email) {
    return emailRegex.test(email);
}

export function PasswordIsStrong(password) {
    return passwordRegex.test(password);
}

export function OTPIsValid(otp) {
    return otp.length === 6;
}
