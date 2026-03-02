"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.BackendManager = void 0;
const vscode = __importStar(require("vscode"));
const path = __importStar(require("path"));
const child_process_1 = require("child_process");
const axios_1 = __importDefault(require("axios"));
class BackendManager {
    constructor() {
        this.backendProcess = null;
        const config = vscode.workspace.getConfiguration("localCopilot");
        this.backendPort = config.get("backendPort", 8089);
        this.backendPath = config.get("backendPath", "");
    }
    getBaseUrl() {
        return `http://localhost:${this.backendPort}`;
    }
    async start(workspacePath) {
        if (this.backendProcess) {
            vscode.window.showInformationMessage("Backend is already  running");
            return true;
        }
        //Find backend binary
        const binaryPath = this.findBackendBinary(workspacePath);
        if (!binaryPath) {
            vscode.window.showErrorMessage("Backend binary not found.");
            return false;
        }
        try {
            //Start backend process
            this.backendProcess = (0, child_process_1.spawn)(binaryPath, ["--workspace", workspacePath, "--port", this.backendPort.toString()], {
                cwd: path.dirname(binaryPath),
            });
            //handle stdout
            this.backendProcess.stdout?.on("data", (data) => {
                console.log(`[Backend Error] ${data}`);
            });
            //handle stderr
            this.backendProcess.stderr?.on("data", (data) => {
                console.log(`[Backend Error] ${data}`);
            });
            //Handle exit
            this.backendProcess.on("exit", (code) => {
                console.log(`Backend existed with code ${code}`);
                this.backendProcess = null;
            });
            //wait for backend to be ready
            await this.waitForBacckend();
            vscode.window.showInformationMessage("Local Copilot backend started");
            return true;
        }
        catch (err) {
            vscode.window.showErrorMessage(`Failed to start backend: ${err}`);
            return false;
        }
    }
    async stop() {
        if (this.backendProcess) {
            this.backendProcess.kill();
            this.backendProcess = null;
            vscode.window.showInformationMessage("Local Copilot backend stopped");
        }
    }
    async isRunning() {
        try {
            const response = await axios_1.default.get(`${this.getBaseUrl()}/health`, {
                timeout: 1000,
            });
            return response.status === 200;
        }
        catch {
            return false;
        }
    }
    findBackendBinary(workspacePath) {
        //If user specified path in settings, use that
        if (this.backendPath) {
            return this.backendPath;
        }
        //Try to find in workspace
        const possiblePaths = [
            path.join(workspacePath, "backend", "copilot"),
            path.join(workspacePath, "..", "backend", "copilot"),
            path.join(workspacePath, "copilot"),
        ];
        for (const p of possiblePaths) {
            try {
                if (require("fs").existsSync(p)) {
                    return p;
                }
            }
            catch {
                continue;
            }
        }
        return null;
    }
    async waitForBacckend(maxAttemps = 10) {
        for (let i = 0; i < maxAttemps; i++) {
            if (await this.isRunning()) {
                return;
            }
            await new Promise((resolve) => setTimeout(resolve, 500));
        }
        throw new Error("Backend failed to start in time");
    }
}
exports.BackendManager = BackendManager;
//# sourceMappingURL=backendManager.js.map