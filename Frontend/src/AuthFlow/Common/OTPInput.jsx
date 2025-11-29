import React, {useRef, useState, useEffect} from "react";


export default function OTPInput({value = "", onValueChange, disabled}) {
    const inputsRef = useRef([]);
    const skipClearOnFocus = useRef(false); // prevents clearing on programmatic focus

    const [buffer, setBuffer] = useState(() => {
        const digits = (value || "").replace(/\D/g, "").slice(0, 6).split("");
        return Array.from({length: 6}, (_, i) => digits[i] || "");
    });

    useEffect(() => {
        const digits = (value || "").replace(/\D/g, "").slice(0, 6).split("");
        setBuffer((cur) => {
            if (cur.join("") === digits.join("")) return cur;
            return Array.from({length: 6}, (_, i) => digits[i] || "");
        });
    }, [value]);

    const emit = (buf) => {
        const compact = buf.join("").replace(/\s/g, "");
        console.log(compact);
        onValueChange && onValueChange(compact);
    };

    const focusAt = (i, programmatic = false) => {
        const el = inputsRef.current[i];
        if (el) {
            if (programmatic) skipClearOnFocus.current = true; // mark next focus as programmatic
            el.focus();
            el.select?.();
        }
    };

    const handleChange = (index, e) => {
        const digit = e.target.value.replace(/\D/g, "").slice(0, 1);

        setBuffer((prev) => {
            const next = [...prev];
            next[index] = digit || "";
            emit(next);
            return next;
        });

        // If user typed a digit, move focus forward BUT mark it as programmatic
        if (digit && index < 5) {
            focusAt(index + 1, true);
        }
    };

    const handleKeyDown = (index, e) => {
        if (e.ctrlKey || e.metaKey) return;

        if (e.key === "Backspace") {
            e.preventDefault();
            setBuffer((prev) => {
                const next = [...prev];
                if (next[index]) {
                    next[index] = "";
                    emit(next);
                    focusAt(index, true);
                } else if (index > 0) {
                    next[index - 1] = "";
                    emit(next);
                    focusAt(index - 1, true);
                }
                return next;
            });
            return;
        }


        if (/^[0-9]$/.test(e.key)) {
            const cur = inputsRef.current[index]?.value || "";
            if (cur === e.key) {
                e.preventDefault();
                setBuffer((prev) => {
                    const next = [...prev];
                    next[index] = e.key;
                    emit(next);
                    return next;
                });
                if (index < 5) focusAt(index + 1, true);
                return;
            }
        }


        if (e.key === "ArrowLeft" && index > 0) {
            e.preventDefault();
            focusAt(index - 1, true);
            return;
        }

        if (e.key === "ArrowRight" && index < 5) {
            e.preventDefault();
            focusAt(index + 1, true);
            return;
        }

        if (!/^[0-9]$/.test(e.key) && e.key.length === 1) {
            e.preventDefault();
        }
    };

    const handlePaste = (index, e) => {
        e.preventDefault();
        skipClearOnFocus.current = true;

        const paste = (e.clipboardData || window.clipboardData).getData("text");
        const digits = (paste || "").replace(/\D/g, "").slice(0, 6).split("");
        if (!digits.length) return;

        setBuffer((prev) => {
            const next = [...prev];
            for (let i = 0; i < digits.length && index + i < 6; i++) {
                next[index + i] = digits[i];
            }
            emit(next);

            const end = Math.min(index + digits.length - 1, 5);
            // focus the last pasted position programmatically
            setTimeout(() => focusAt(end, true), 0);
            return next;
        });
    };

    const handleFocus = (index) => {
        // if this focus was programmatic, skip clearing and reset flag
        if (skipClearOnFocus.current) {
            skipClearOnFocus.current = false;
            return;
        }

        // manual focus (click): clear that slot only
        setBuffer((prev) => {
            const next = [...prev];
            next[index] = "";
            emit(next);
            return next;
        });

        setTimeout(() => inputsRef.current[index]?.select?.(), 0);
    };

    return (
        <div className="flex justify-between gap-2 mt-2">
            {Array.from({length: 6}).map((_, i) => (
                <input
                    key={i}
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]*"
                    maxLength={1}
                    disabled={disabled}
                    value={buffer[i] || ""}
                    ref={(el) => (inputsRef.current[i] = el)}
                    onChange={(e) => handleChange(i, e)}
                    onKeyDown={(e) => handleKeyDown(i, e)}
                    onPaste={(e) => handlePaste(i, e)}
                    onFocus={() => handleFocus(i)}
                    className="
            w-12 h-12 text-center rounded-md
            bg-[#0b0f14] border border-gray-700
            text-white text-lg tracking-widest
            focus:outline-none focus:ring-2 focus:ring-indigo-500
          "
                    aria-label={`OTP digit ${i + 1}`}
                />
            ))}
        </div>
    );
}
