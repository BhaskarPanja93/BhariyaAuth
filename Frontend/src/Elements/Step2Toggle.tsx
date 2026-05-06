export default function Step2Toggle(
    { usingOTP, toggleUsingOTP, disabled }:
    {
        usingOTP: boolean,
        toggleUsingOTP:()=>void,
        disabled:boolean
    }){
    return (
        <button
            disabled={disabled}
            onClick={toggleUsingOTP}
            type="button" className="text-xs text-indigo-400 hover:underline">
            {usingOTP ? 'Use Password' : 'Use OTP'}
        </button>
    )
}