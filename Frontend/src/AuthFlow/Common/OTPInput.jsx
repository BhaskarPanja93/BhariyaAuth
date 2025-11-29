import React, { useState } from 'react';
import { InputOtp } from 'primereact/inputotp';

export default function TemplateDemo() {
    const [token, setTokens] = useState();

    const customInput = ({events, props}) => <input {...events} {...props} type="text" className="custom-otp-input" />;

    return (
        <div className="card flex justify-content-center">
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
                value={token}
                onChange={(e) => setTokens(e.value)}
                inputTemplate={customInput}
                length={6}

            />
        </div>
    );
}
