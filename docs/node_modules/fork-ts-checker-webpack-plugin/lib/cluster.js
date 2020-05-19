"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const childProcess = require("child_process");
const path = require("path");
const process = require("process");
const worker_rpc_1 = require("worker-rpc");
const NormalizedMessage_1 = require("./NormalizedMessage");
const RpcTypes_1 = require("./RpcTypes");
// fork workers...
const division = parseInt(process.env.WORK_DIVISION || '', 10);
const workers = [];
for (let num = 0; num < division; num++) {
    workers.push(childProcess.fork(path.resolve(__dirname, './service.js'), [], {
        execArgv: ['--max-old-space-size=' + process.env.MEMORY_LIMIT],
        env: Object.assign({}, process.env, { WORK_NUMBER: num.toString() }),
        stdio: ['inherit', 'inherit', 'inherit', 'ipc']
    }));
}
// communication with parent process
const parentRpc = new worker_rpc_1.RpcProvider(message => {
    try {
        process.send(message);
    }
    catch (e) {
        // channel closed...
        process.exit();
    }
});
process.on('message', message => parentRpc.dispatch(message));
// communication with worker processes
const workerRpcs = workers.map(worker => {
    const rpc = new worker_rpc_1.RpcProvider(message => {
        try {
            worker.send(message);
        }
        catch (e) {
            // channel closed - something went wrong - close cluster...
            process.exit();
        }
    });
    worker.on('message', message => rpc.dispatch(message));
    return rpc;
});
parentRpc.registerRpcHandler(RpcTypes_1.RUN, (message) => __awaiter(this, void 0, void 0, function* () {
    const workerResults = yield Promise.all(workerRpcs.map(workerRpc => workerRpc.rpc(RpcTypes_1.RUN, message)));
    function workerFinished(workerResult) {
        return workerResult.every(result => typeof result !== 'undefined');
    }
    if (!workerFinished(workerResults)) {
        return undefined;
    }
    const merged = workerResults.reduce((innerMerged, innerResult) => ({
        diagnostics: innerMerged.diagnostics.concat(innerResult.diagnostics.map(NormalizedMessage_1.NormalizedMessage.createFromJSON)),
        lints: innerMerged.lints.concat(innerResult.lints.map(NormalizedMessage_1.NormalizedMessage.createFromJSON))
    }), { diagnostics: [], lints: [] });
    merged.diagnostics = NormalizedMessage_1.NormalizedMessage.deduplicate(merged.diagnostics);
    merged.lints = NormalizedMessage_1.NormalizedMessage.deduplicate(merged.lints);
    return merged;
}));
process.on('SIGINT', () => {
    process.exit();
});
process.on('exit', () => {
    workers.forEach(worker => {
        try {
            worker.kill();
        }
        catch (e) {
            // do nothing...
        }
    });
});
//# sourceMappingURL=cluster.js.map