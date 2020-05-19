# openapi-generator

[![NPM version][npm-image]][npm-url]
[![build status][travis-image]][travis-url]
[![Test coverage][codecov-image]][codecov-url]
[![David deps][david-image]][david-url]
[![Known Vulnerabilities][snyk-image]][snyk-url]
[![npm download][download-image]][download-url]

[npm-image]: https://img.shields.io/npm/v/openapi-generator.svg?style=flat-square
[npm-url]: https://npmjs.org/package/openapi-generator
[travis-image]: https://img.shields.io/travis/zhang740/openapi-generator.svg?style=flat-square
[travis-url]: https://travis-ci.org/zhang740/openapi-generator
[codecov-image]: https://codecov.io/github/zhang740/openapi-generator/coverage.svg?branch=master
[codecov-url]: https://codecov.io/github/zhang740/openapi-generator?branch=master
[david-image]: https://img.shields.io/david/zhang740/openapi-generator.svg?style=flat-square
[david-url]: https://david-dm.org/zhang740/openapi-generator
[snyk-image]: https://snyk.io/test/npm/openapi-generator/badge.svg?style=flat-square
[snyk-url]: https://snyk.io/test/npm/openapi-generator
[download-image]: https://img.shields.io/npm/dm/openapi-generator.svg?style=flat-square
[download-url]: https://npmjs.org/package/openapi-generator

# Quick View

openapi-generator from swagger 2.0 or OpenAPI 3.0:

## Simple

`openapi-generator url http://xxx/v2/api-docs -c true`

## Use Config

`openapi-generator config ./xxx.js` or `openapi-generator config ./xxx.json`

Config interface:

```ts
interface CliConfig {
  api: string;

  /** dir for openapi-generator */
  sdkDir: string;
  /** path of service template */
  templatePath?: string;
  /** path of interface template */
  interfaceTemplatePath?: string;
  /** request lib */
  requestLib = true;
  /** filename style, true 为大驼峰，lower 为小驼峰 */
  camelCase?: boolean | 'lower' = false;
  /** gen type */
  type?: 'ts' | 'js' = 'ts';
  /** service type */
  serviceType?: 'function' | 'class' = 'function';
  /** namespace of typings */
  namespace?: string = 'API';
  /** 自动清除旧文件时忽略列表 */
  ignoreDelete: string[] = [];
}
```

### genAPISDK

`function genAPISDK(data: RouteMetadataType[], config: GenConfig) => void`
