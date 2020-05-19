declare namespace DRange {
    interface SubRange {
        low: number;
        high: number;
        length: number;
    }
}

/**
 * For adding/subtracting sets of range of numbers.
 */
declare class DRange {
    /**
     * Creates a new instance of DRange.
     */
    constructor(low?: number, high?: number);

    /**
     * The total length of all subranges
     */
    length: number;

    /**
     * Adds a subrange
     */
    add(low: number, high?: number): this;
    /**
     * Adds all of another DRange's subranges
     */
    add(drange: DRange): this;

    /**
     * Subtracts a subrange
     */
    subtract(low?: number, high?: number): this;
    /**
     * Subtracts all of another DRange's subranges
     */
    subtract(drange: DRange): this;

    /**
     * Keep only subranges that overlap the given subrange
     */
    intersect(low?: number, high?: number): this;
    /**
     * Intersect all of another DRange's subranges
     */
    intersect(drange: DRange): this;

    /**
     * Get the number at the specified index
     */
    index(i: number): number;

    /**
     * Clones the drange, so that changes to it are not reflected on its clone
     */
    clone(): this;

    toString(): string;

    /**
     * Get contained numbers
     */
    numbers(): number[];

    /**
     * Get copy of subranges
     */
    subranges(): DRange.SubRange[];
}

export = DRange;
