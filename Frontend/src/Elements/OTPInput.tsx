import {InputOtp} from 'primereact/inputotp';
import React from "react";

type InputTemplateOptions = {
    props: React.InputHTMLAttributes<HTMLInputElement>;
    events: React.InputHTMLAttributes<HTMLInputElement>;
};

export default function OTPInput(
    {value, onValueChange, disabled}:
    {
        value:string,
        onValueChange:React.Dispatch<React.SetStateAction<string>>,
        disabled:boolean
    }) {
    return (<div className="card flex justify-content-center">
            <style scoped>
                {`
                    .custom-otp-input {
                        width: 3rem;                 
                        height: 3rem;                
                        text-align: center;
                        border-radius: 0.375rem;     
                        background-color: #0b0f14;
                        border: 1px solid #4b5563;   
                        color: white;
                        font-size: 1.125rem;         
                        letter-spacing: 0.1em;       
                        outline: none;
                        transition: box-shadow 0.2s;
                        margin-right: 0.3rem;
                    }

                    .custom-otp-input:focus {
                        box-shadow: 0 0 0 2px #6366f1; 
                        border-color: #6366f1;
                    }

                `}
            </style>

            <InputOtp
                value={value}
                onChange={(e) => onValueChange(String(e.value ?? ""))}
                inputTemplate={(options) => {
                    const {events, props} = options as unknown as InputTemplateOptions;

                    return (
                        <input
                            {...props}
                            {...events}
                            type="text"
                            className={`${props.className ?? ""} custom-otp-input`.trim()}
                            name="otp-input"
                        />
                    );
                }}
                length={6}
                disabled={disabled}
            />
        </div>);
}
