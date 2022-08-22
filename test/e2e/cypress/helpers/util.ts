export function extractRecoveryCode(body: string): string | null{
    const codeRegex = /(\d{8})/;
    const result = codeRegex.exec(body);
    if (result != null && result.length > 0) {
        return result[0];
    }
    return null;
}