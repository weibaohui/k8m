import React, { useState, useEffect } from "react";
import ReactDOM from "react-dom";
import { Popover, Button } from "@arco-design/web-react";
import { render as amisRender } from "amis";

const GlobalTextSelector: React.FC = () => {
    const [selection, setSelection] = useState<{ text: string; x: number; y: number } | null>(null);

    useEffect(() => {
        const handleMouseUp = () => {
            const selectedText = window.getSelection()?.toString().trim();
            console.log("选中文本:", selectedText); // ✅ 调试信息

            if (selectedText) {
                const range = window.getSelection()?.getRangeAt(0);
                if (range) {
                    const rect = range.getBoundingClientRect();
                    console.log("选中文本位置:", rect); // ✅ 位置日志
                    setSelection({
                        text: selectedText,
                        x: rect.left + window.scrollX,
                        y: rect.bottom + window.scrollY
                    });
                }
            } else {
                setSelection(null);
            }
        };

        document.addEventListener("mouseup", handleMouseUp);
        return () => {
            document.removeEventListener("mouseup", handleMouseUp);
        };
    }, []);

    if (!selection) return null;

    return ReactDOM.createPortal(
        <div
            style={{
                position: "absolute",
                top: selection.y + 5,
                left: selection.x,
                zIndex: 100000000,
            }}
        >
            <Popover
                defaultPopupVisible
                trigger="manual"
                position="bottom"
                style={{
                    zIndex: 100000000,
                }}
                content={
                    <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
                        <div><strong>{selection.text}</strong></div>
                        {

                            amisRender({
                                "type": "websocketMarkdownViewer",
                                "url": "/k8s/chat/any_selection",
                                "params": {
                                    "question": selection.text
                                }
                            })
                        }
                    </div>
                }
            >
                <span style={{ width: 1, height: 1, display: "inline-block" }} />
            </Popover >
        </div >,
        document.body
    );
};

export default GlobalTextSelector;
