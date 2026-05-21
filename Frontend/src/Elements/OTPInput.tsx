import {InputOtp} from "primereact/inputotp";
import React from "react";

type InputTemplateOptions = {
    props: React.InputHTMLAttributes<HTMLInputElement>;
    events: React.InputHTMLAttributes<HTMLInputElement>;
};

export default function OTPInput(
    {value, onValueChange, disabled}:
        {
            value: string;
            onValueChange: React.Dispatch<React.SetStateAction<string>>;
            disabled: boolean;
        },
) {
    return <div className="card flex justify-content-center">
        <InputOtp
            value={value}
            onChange={(event) => onValueChange(String(event.value ?? ""))}
            inputTemplate={(options) => {
                const {events, props} = options as unknown as InputTemplateOptions;

                return <input
                    {...props}
                    {...events}
                    type="text"
                    className={`${props.className ?? ""} custom-otp-input`.trim()}
                    name="otp-input"/>;
            }}
            length={6}
            disabled={disabled}/>
    </div>;
}


