const emailRegex = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/;
const passwordRegex = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d).{7,20}$/;

export function EmailIsValid(email) {
    return emailRegex.test(email);
}

export function PasswordIsStrong(password) {
    return passwordRegex.test(password);
}
