export interface ConsolaLogObject {
  level?: number,
  tag?: string,
  type?: string,
  message?: string,
  additional?: string | string[],
  args?: any[],
}

type ConsolaMock = (...args: any) => void

type ConsolaMockFn = (type: string, defaults: ConsolaLogObject) => ConsolaMock

export interface ConsolaReporterArgs {
  async: boolean,
  stdout: any,
  stderr: any,
}

export interface ConsolaReporter {
  log: (logObj: ConsolaLogObject, args: ConsolaReporterArgs) => void
}

export interface ConsolaOptions {
  reporters?: ConsolaReporter[],
  types?: { [type: string]: ConsolaLogObject },
  level?: number,
  defaults?: ConsolaLogObject,
  async?: boolean,
  stdout?: any,
  stderr?: any,
  mockFn?: ConsolaMockFn,
  throttle?: number,
}

export declare class Consola {
  constructor(options: ConsolaOptions)

  level: number
  readonly stdout: any
  readonly stderr: any

  // Built-in log levels
  fatal(message: ConsolaLogObject | any, ...args: any[]): void
  error(message: ConsolaLogObject | any, ...args: any[]): void
  warn(message: ConsolaLogObject | any, ...args: any[]): void
  log(message: ConsolaLogObject | any, ...args: any[]): void
  info(message: ConsolaLogObject | any, ...args: any[]): void
  start(message: ConsolaLogObject | any, ...args: any[]): void
  success(message: ConsolaLogObject | any, ...args: any[]): void
  ready(message: ConsolaLogObject | any, ...args: any[]): void
  debug(message: ConsolaLogObject | any, ...args: any[]): void
  trace(message: ConsolaLogObject | any, ...args: any[]): void

  // Create
  create(options: ConsolaOptions): Consola
  withDefaults(defaults: ConsolaLogObject): Consola

  withTag(tag: string): Consola
  withScope(tag: string): Consola

  // Reporter
  addReporter(reporter: ConsolaReporter): Consola
  setReporters(reporters: Array<ConsolaReporter>): Consola

  removeReporter(reporter?: ConsolaReporter): Consola
  remove(reporter?: ConsolaReporter): Consola
  clear(reporter?: ConsolaReporter): Consola

  // Wrappers
  wrapAll(): void
  restoreAll(): void
  wrapConsole(): void
  restoreConsole(): void
  wrapStd(): void
  restoreStd(): void

  // Pause/Resume
  pauseLogs(): void
  pause(): void

  resumeLogs(): void
  resume(): void

  // Mock
  mockTypes(mockFn: ConsolaMockFn): any
  mock(mockFn: ConsolaMockFn): any
}

export declare class BrowserReporter implements ConsolaReporter {
  log: (logObj: ConsolaLogObject, args: ConsolaReporterArgs) => void
}

declare const consolaGlobalInstance: Consola;

export default consolaGlobalInstance

