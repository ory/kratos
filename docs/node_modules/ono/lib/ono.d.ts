interface ErrorPOJO {
  message?: string;
  stack?: string;
  name?: string;
  [key: string]: unknown;
}

type ErrorLike = Error | ErrorPOJO;

declare const ono: {
  (error: ErrorLike): Error;
  (error: ErrorLike, props: object): Error;
  (error: ErrorLike, message: string, ...params: any[]): Error;
  (error: ErrorLike, props: object, message: string, ...params: any[]): Error;
  (message: string, ...params: any[]): Error;
  (props: object): Error;
  (props: object, message: string, ...params: any[]): Error;


  error(error: ErrorLike): Error;
  error(error: ErrorLike, props: object): Error;
  error(error: ErrorLike, message: string, ...params: any[]): Error;
  error(error: ErrorLike, props: object, message: string, ...params: any[]): Error;
  error(message: string, ...params: any[]): Error;
  error(props: object): Error;
  error(props: object, message: string, ...params: any[]): Error;


  eval(error: ErrorLike): EvalError;
  eval(error: ErrorLike, props: object): EvalError;
  eval(error: ErrorLike, message: string, ...params: any[]): EvalError;
  eval(error: ErrorLike, props: object, message: string, ...params: any[]): EvalError;
  eval(message: string, ...params: any[]): EvalError;
  eval(props: object): EvalError;
  eval(props: object, message: string, ...params: any[]): EvalError;


  range(error: ErrorLike): RangeError;
  range(error: ErrorLike, props: object): RangeError;
  range(error: ErrorLike, message: string, ...params: any[]): RangeError;
  range(error: ErrorLike, props: object, message: string, ...params: any[]): RangeError;
  range(message: string, ...params: any[]): RangeError;
  range(props: object): RangeError;
  range(props: object, message: string, ...params: any[]): RangeError;


  reference(error: ErrorLike): ReferenceError;
  reference(error: ErrorLike, props: object): ReferenceError;
  reference(error: ErrorLike, message: string, ...params: any[]): ReferenceError;
  reference(error: ErrorLike, props: object, message: string, ...params: any[]): ReferenceError;
  reference(message: string, ...params: any[]): ReferenceError;
  reference(props: object): ReferenceError;
  reference(props: object, message: string, ...params: any[]): ReferenceError;


  syntax(error: ErrorLike): SyntaxError;
  syntax(error: ErrorLike, props: object): SyntaxError;
  syntax(error: ErrorLike, message: string, ...params: any[]): SyntaxError;
  syntax(error: ErrorLike, props: object, message: string, ...params: any[]): SyntaxError;
  syntax(message: string, ...params: any[]): SyntaxError;
  syntax(props: object): SyntaxError;
  syntax(props: object, message: string, ...params: any[]): SyntaxError;


  type(error: ErrorLike): TypeError;
  type(error: ErrorLike, props: object): TypeError;
  type(error: ErrorLike, message: string, ...params: any[]): TypeError;
  type(error: ErrorLike, props: object, message: string, ...params: any[]): TypeError;
  type(message: string, ...params: any[]): TypeError;
  type(props: object): TypeError;
  type(props: object, message: string, ...params: any[]): TypeError;


  uri(error: ErrorLike): URIError;
  uri(error: ErrorLike, props: object): URIError;
  uri(error: ErrorLike, message: string, ...params: any[]): URIError;
  uri(error: ErrorLike, props: object, message: string, ...params: any[]): URIError;
  uri(message: string, ...params: any[]): URIError;
  uri(props: object): URIError;
  uri(props: object, message: string, ...params: any[]): URIError;


  formatter(message: string, ...params: any[]): string;
}

export = ono;
