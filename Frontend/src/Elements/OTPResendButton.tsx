export default function OTPResendButton(
    { delay, onClick, disabled }:
    {
        delay:number,
        onClick:()=>void,
        disabled:boolean
    }) {
    return (
        <button
            onClick={onClick}
            disabled={disabled || delay !== 0}
            className="text-xs text-indigo-400 hover:underline"
            type="button">
            {delay <= 0 ? "Resend OTP" : "Resend in "+delay.toFixed(1)}
        </button>
    )
}

