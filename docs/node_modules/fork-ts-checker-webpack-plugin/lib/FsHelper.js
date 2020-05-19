"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs = require("fs");
function fileExistsSync(filePath) {
    try {
        fs.statSync(filePath);
    }
    catch (err) {
        if (err.code === 'ENOENT') {
            return false;
        }
        else {
            throw err;
        }
    }
    return true;
}
exports.fileExistsSync = fileExistsSync;
function throwIfIsInvalidSourceFileError(filepath, error) {
    if (fileExistsSync(filepath) &&
        // check the error type due to file system lag
        !(error instanceof Error) &&
        !(error.constructor.name === 'FatalError') &&
        !(error.message && error.message.trim().startsWith('Invalid source file'))) {
        // it's not because file doesn't exist - throw error
        throw error;
    }
}
exports.throwIfIsInvalidSourceFileError = throwIfIsInvalidSourceFileError;
//# sourceMappingURL=FsHelper.js.map