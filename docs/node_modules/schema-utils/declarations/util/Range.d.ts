export = Range;
/**
 * @typedef {[number, boolean]} RangeValue
 */
/**
 * @callback RangeValueCallback
 * @param {RangeValue} rangeValue
 * @returns {boolean}
 */
declare class Range {
  /** @type {Array<RangeValue>} */
  _left: Array<RangeValue>;
  /** @type {Array<RangeValue>} */
  _right: Array<RangeValue>;
  /**
   * @param {number} value
   * @param {boolean=} exclusive
   */
  left(value: number, exclusive?: boolean | undefined): void;
  /**
   * @param {number} value
   * @param {boolean=} exclusive
   */
  right(value: number, exclusive?: boolean | undefined): void;
  /**
   * @param {boolean} logic is not logic applied
   * @return {string} "smart" range string representation
   */
  format(logic?: boolean): string;
}
declare namespace Range {
  export {
    getOperator,
    formatRight,
    formatLeft,
    formatRange,
    getRangeValue,
    RangeValue,
    RangeValueCallback,
  };
}
type RangeValue = [number, boolean];
/**
 * @param {"left" | "right"} side
 * @param {boolean} exclusive
 * @returns {">" | ">=" | "<" | "<="}
 */
declare function getOperator(
  side: 'left' | 'right',
  exclusive: boolean
): '>' | '>=' | '<' | '<=';
/**
 * @param {number} value
 * @param {boolean} logic is not logic applied
 * @param {boolean} exclusive is range exclusive
 * @returns {string}
 */
declare function formatRight(
  value: number,
  logic: boolean,
  exclusive: boolean
): string;
/**
 * @param {number} value
 * @param {boolean} logic is not logic applied
 * @param {boolean} exclusive is range exclusive
 * @returns {string}
 */
declare function formatLeft(
  value: number,
  logic: boolean,
  exclusive: boolean
): string;
/**
 * @param {number} start left side value
 * @param {number} end right side value
 * @param {boolean} startExclusive is range exclusive from left side
 * @param {boolean} endExclusive is range exclusive from right side
 * @param {boolean} logic is not logic applied
 * @returns {string}
 */
declare function formatRange(
  start: number,
  end: number,
  startExclusive: boolean,
  endExclusive: boolean,
  logic: boolean
): string;
/**
 * @param {Array<RangeValue>} values
 * @param {boolean} logic is not logic applied
 * @return {RangeValue} computed value and it's exclusive flag
 */
declare function getRangeValue(
  values: [number, boolean][],
  logic: boolean
): [number, boolean];
type RangeValueCallback = (rangeValue: [number, boolean]) => boolean;
