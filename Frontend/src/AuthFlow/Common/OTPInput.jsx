import { InputOtp } from 'primereact/inputotp';

export default function OTPInput({value, onValueChange, disabled}) {

    return (
        <div className="flex justify-between gap-2 mt-2">
            <InputOtp value={value} onChange={(e) => onValueChange(e.value)} integerOnly disabled={disabled} length={6} />
        </div>
    );
}
