import { useEffect, useRef } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";

interface XtermTerminalProps {
  wsUrl: string;
  onClose?: () => void;
}

function XtermTerminal({ wsUrl, onClose }: XtermTerminalProps) {
  const termRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<Terminal | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!termRef.current) return;

    const terminal = new Terminal({
      cursorBlink: true,
      theme: {
        background: "#0a0a0a",
        foreground: "#22c55e",
        cursor: "#22c55e",
      },
      fontSize: 13,
      fontFamily: "'Geist Mono', 'Fira Code', monospace",
    });

    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(termRef.current);
    fitAddon.fit();

    terminalRef.current = terminal;

    // Connect WebSocket
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      terminal.writeln("Connected to terminal...\r\n");
    };

    ws.onmessage = (event) => {
      terminal.write(event.data);
    };

    ws.onclose = () => {
      terminal.writeln("\r\n\x1b[31mConnection closed.\x1b[0m");
      onClose?.();
    };

    ws.onerror = () => {
      terminal.writeln("\r\n\x1b[31mWebSocket error.\x1b[0m");
    };

    // Send terminal input to WebSocket
    terminal.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });

    // Handle resize
    const handleResize = () => {
      fitAddon.fit();
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          JSON.stringify({
            type: "resize",
            cols: terminal.cols,
            rows: terminal.rows,
          })
        );
      }
    };

    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      ws.close();
      terminal.dispose();
    };
  }, [wsUrl, onClose]);

  return (
    <div
      ref={termRef}
      className="h-full w-full min-h-[400px] rounded-lg overflow-hidden border border-border"
    />
  );
}

export default XtermTerminal;
