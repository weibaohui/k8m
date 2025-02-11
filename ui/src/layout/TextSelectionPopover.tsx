import React, {useState, useEffect} from "react";
import ReactDOM from "react-dom";
import {render as amisRender} from "amis";
import {Card} from "amis-ui";

// 判断选区是否在 Monaco Editor
const isInMonacoEditor = (node: Node | null): boolean => {
    while (node) {
        if (node instanceof HTMLElement && node.classList.contains("monaco-editor")) {
            return true;
        }
        node = node.parentNode;
    }
    return false;
};

// 获取 Input / Textarea 的光标位置
const getInputCaretCoords = (input: HTMLInputElement | HTMLTextAreaElement, selectionStart: number) => {
    const rect = input.getBoundingClientRect();
    const offset = selectionStart * 7; // 估算字符宽度（可根据实际情况调整）
    return {
        x: rect.left + offset + window.scrollX,
        y: rect.top + rect.height + window.scrollY
    };
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

            let x = 0, y = 0;
            const activeElement = document.activeElement as HTMLElement;

            if (activeElement && (activeElement.tagName === "INPUT" || activeElement.tagName === "TEXTAREA")) {
                // 🛠️ 处理 Input / Textarea 选区
                console.log("选中了 Input / Textarea");
                const input = activeElement as HTMLInputElement | HTMLTextAreaElement;
                const selectionStart = input.selectionStart || 0;
                const coords = getInputCaretCoords(input, selectionStart);
                x = coords.x;
                y = coords.y;
            } else {
                // 🛠️ 处理 Monaco Editor 选区
                const selectionObj = window.getSelection();
                const range = selectionObj?.rangeCount ? selectionObj.getRangeAt(0) : null;

                if (range && isInMonacoEditor(range.commonAncestorContainer)) {
                    console.log("选中了 Monaco Editor");
                    const editorElement = document.querySelector(".monaco-editor") as HTMLElement;
                    if (editorElement) {
                        const rect = editorElement.getBoundingClientRect();
                        x = rect.left + window.scrollX + 100; // 偏移以适应 Editor
                        y = rect.top + window.scrollY + 40;
                    }
                } else if (range) {
                    // 🛠️ 普通文本选区
                    const rect = range.getBoundingClientRect();
                    x = rect.left + window.scrollX;
                    y = rect.bottom + window.scrollY;
                }
            }

            setSelection({text: selectedText, x, y});
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
            <Card style={{width: '50hv', maxWidth: '500px'}}
                  title={selection.text.length > 40 ? selection.text.slice(0, 40) + "..." : selection.text}
            >
                <div style={{display: "flex", flexDirection: "column", alignItems: "center"}}>
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
            </Card>


        </div>,
        document.body
    );
};

export default GlobalTextSelector;
