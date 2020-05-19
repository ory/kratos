import * as DRange from "drange";

declare namespace RandExp {}

declare class RandExp {
    static randexp(pattern: string | RegExp, flags?: string): string;
    static sugar(): void;
    constructor(pattern: string | RegExp, flags?: string);
    gen(): string;
    defaultRange: DRange;
    randInt: (from: number, to: number) => number;
    max: number;
}

export = RandExp;
