import React, { useState, useEffect } from "react";
import ReactDOM from "react-dom";
import { Popover } from "@arco-design/web-react";
import { render as amisRender } from "amis";

// 检测选区是否在 Monaco Editor
const isInMonacoEditor = (node: Node | null): boolean => {
    while (node) {
        if (node instanceof HTMLElement && node.classList.contains("monaco-editor")) {
            return true;
        }
        node = node.parentNode;
    }
    return false;
};

const GlobalTextSelector: React.FC = () => {
    const [selection, setSelection] = useState<{ text: string; x: number; y: number } | null>(null);

    useEffect(() => {
        const handleMouseUp = () => {
            const selectedText = window.getSelection()?.toString().trim();
            console.log("选中文本:", selectedText); // ✅ 调试信息

            if (!selectedText) {
                setSelection(null);
                return;
            }

            const selectionObj = window.getSelection();
            const range = selectionObj?.rangeCount ? selectionObj.getRangeAt(0) : null;
            let x = 0, y = 0;

            // 🛠️ 处理 Monaco Editor 选中文字的情况
            if (range && isInMonacoEditor(range.commonAncestorContainer)) {
                console.log("选中了 Monaco Editor 内的文本");
                const editorElement = document.querySelector(".monaco-editor") as HTMLElement;
                if (editorElement) {
                    const rect = editorElement.getBoundingClientRect();
                    x = rect.left + window.scrollX + 100; // 手动调整 X 偏移
                    y = rect.top + window.scrollY + 40; // 手动调整 Y 偏移
                }
            } else if (range) {
                // 普通文本选区
                const rect = range.getBoundingClientRect();
                x = rect.left + window.scrollX;
                y = rect.bottom + window.scrollY;
            }

            setSelection({ text: selectedText, x, y });
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
                    width: "100%",
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
            </Popover>
        </div>,
        document.body
    );
};

export default GlobalTextSelector;
