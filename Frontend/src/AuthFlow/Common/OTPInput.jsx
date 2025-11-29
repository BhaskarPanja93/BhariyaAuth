import { InputOtp } from 'primereact/inputotp';

export default function OTPInput({value, onValueChange, disabled}) {

    return (
        <div className="card flex justify-content-center">
            <InputOtp value={value} onChange={(e) => onValueChange(e.value)} integerOnly disabled={disabled} length={6} />
        </div>
    );
}
