<h1 align="center"><img src="https://raw.githubusercontent.com/ory/meta/master/static/banners/kratos.svg" alt="ORY Kratos - Cloud native Identity and User Management"></h1>

<h4 align="center">
    <a href="https://www.ory.sh/chat">Chat</a> |
    <a href="https://community.ory.sh/">Forums</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/kratos/docs/">Guide</a> |
    <a href="https://www.ory.sh/kratos/docs/sdk/api">API Docs</a> |
    <a href="https://godoc.org/github.com/ory/kratos">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a>
</h4>

<h4 align="center">
<a href="./README.md">English README</a>
</h4>

---

<p align="left">
    <a href="https://circleci.com/gh/ory/kratos/tree/master"><img src="https://circleci.com/gh/ory/kratos/tree/master.svg?style=shield" alt="Build Status"></a>
    <a href="https://coveralls.io/github/ory/kratos?branch=master"> <img src="https://coveralls.io/repos/ory/kratos/badge.svg?branch=master&service=github" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/ory/kratos"><img src="https://goreportcard.com/badge/github.com/ory/kratos" alt="Go Report Card"></a>
    <a href="https://bestpractices.coreinfrastructure.org/projects/364"><img src="https://bestpractices.coreinfrastructure.org/projects/364/badge" alt="CII Best Practices"></a>
    <a href="https://opencollective.com/ory" alt="sponsors on Open Collective"><img src="https://opencollective.com/ory/backers/badge.svg" /></a>
    <a href="https://opencollective.com/ory" alt="Sponsors on Open Collective"><img src="https://opencollective.com/ory/sponsors/badge.svg" /></a>
</p>

ORY Kratos 是当今世界上有且仅有的一个支持云原生的身份认证和用户管理系统。
终于让广大开发者不用一次又一次的实现用户登录功能了。

**内容目录：**

- [为什么是 ORY Kratos？](#为什么是-ORY-Kratos？)
- [谁正在使用 ORY Kratos？](#谁正在使用-ORY-Kratos？)
- [千里之行 始于足下](#千里之行-始于足下)
  - [快速开始](#快速开始)
  - [安装指南](#安装指南)
- [ORY Kratos的生态](#ORY-Kratos的生态)
  - [ORY Kratos：身份认证和用户管理系统](#ORY-Kratos：身份认证和用户管理系统)
  - [ORY Hydra：OAuth2 和 OpenID 联接的服务端](#ORY-Hydra：OAuth2-和-OpenID-联接的服务端)
  - [ORY OAuthkeeper：认证和接入代理](#ORY0OAuthkeeper：认证和接入代理)
  - [ORY Keto：将接入控制策略作为服务](#ORY-Keto：将接入控制策略作为服务)
- [安全](#安全)
  - [揭露漏洞](#揭露漏洞)
- [匿名数据](#匿名数据)
- [文档](#文档)
  - [指引](#指引)
  - [HTTP API 文档](#HTTP-API-文档)
  - [更新和变更日志](#更新和变更日志)
  - [命令行文档](#命令行文档)
  - [开发](#开发)
    - [依赖](#依赖)
    - [从源码安装](#从源码安装)
    - [格式化代码](#格式化代码)
    - [运行测试](#运行测试)
      - [简短测试](#简短测试)
      - [常规测试](#常规测试)
      - [端到端测试](#端到端测试)
    - [构建Docker](#构建Docker)
    - [文档测试](#文档测试)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## 为什么是 ORY Kratos？

ORY Kratos 是根据 [云架构最佳实践](https://www.ory.sh/docs/ecosystem/software-architecture-philosophy)
构建的API优先的身份和用户管理系统。
它实现了几乎每一个软件应用程序都需要实现的核心使用功能，如下：

- **登录和注册**：允许终端用户创建和登录其账户，使用用户名/电子邮件和密码组合、社交登录（“使用Google、GitHub登录”），
  无密码免验证登录等。
- **多条件身份验证（MFA/2FA）**：支持 TOTP（[RFC 6238](https://tools.ietf.org/html/rfc6238) ，
  [IETF RFC 4226](https://tools.ietf.org/html/rfc4226) 和更为人熟知的
  [Google Authenticator](https://en.wikipedia.org/wiki/Google_Authenticator)）等协议
- **帐户验证**：验证电子邮件地址、电话号码或物理地址是否确实属于该身份。
- **帐户恢复**：使用“忘记密码”工作流、安全代码（在 MFA 设备丢失的情况下）等恢复访问。
- **用户档案和帐户管理**：更新密码，个人资料，电子邮件地址，链接的社会档案使用安全流。
- **管理员级 API**：导入、更新、删除认证信息。

我们强烈建议你阅读 [ORY Kratos 简介文档](https://www.ory.sh/kratos/docs/) ，以了解有关 ORY Kratos 的背景，功能集合以及与其他产品的区别的更多信息。

## 谁正在使用 ORY Kratos？

<!--BEGIN ADOPTERS-->

ORY 社区站在个人，公司和维护者的肩膀上。
我们感谢参与其中的每个人，他们有的提交错误报告和功能请求，有的提供补丁程序，还有的赞助我们的工作。
我们的社区现在是拥有 1000 多人强大的群体，并且正在迅速发展。
ORY stack 每个月通过超过 250.000+ 个活动服务节点来保护 16.000.000.000+ 个 API 请求。
没有你们每个人，我们将永远无法实现这一目标！

以下列表代表了一直陪伴我们并为我们的生态系统做出杰出贡献的公司。如果你认为你的公司应该出现在这里，请立即联系<a href="mailto:office-muc@ory.sh">office-muc@ory.sh</a>！

**我们在
<a href="https://www.patreon.com/_ory">Patreon</a>
或 <a href="https://opencollective.com/ory">Open Collective</a> 上的开源作品，
请考虑通过成为我们的赞助商来回馈我们。**

<table>
    <thead>
        <tr>
          <th>类型</th>
          <th>名称</th>
          <th>LOGO</th>
          <th>网站</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>Sponsor</td>
            <td>Raspberry PI Foundation</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/raspi.svg" alt="Raspberry PI Foundation"></td>
            <td><a href="https://www.raspberrypi.org/">raspberrypi.org</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Kyma Project</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/kyma.svg" alt="Kyma Project"></td>
            <td><a href="https://kyma-project.io">kyma-project.io</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>ThoughtWorks</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/tw.svg" alt="ThoughtWorks"></td>
            <td><a href="https://www.thoughtworks.com/">thoughtworks.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Tulip</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/tulip.svg" alt="Tulip Retail"></td>
            <td><a href="https://tulip.com/">tulip.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Cashdeck / All My Funds</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/allmyfunds.svg" alt="All My Funds"></td>
            <td><a href="https://cashdeck.com.au/">cashdeck.com.au</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>3Rein</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/3R-horiz.svg" alt="3Rein"></td>
            <td><a href="https://3rein.com/">3rein.com</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Hootsuite</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/hootsuite.svg" alt="Hootsuite"></td>
            <td><a href="https://hootsuite.com/">hootsuite.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Segment</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/segment.svg" alt="Segment"></td>
            <td><a href="https://segment.com/">segment.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Arduino</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/arduino.svg" alt="Arduino"></td>
            <td><a href="https://www.arduino.cc/">arduino.cc</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>DataDetect</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/datadetect.svg" alt="Datadetect"></td>
            <td><a href="https://unifiedglobalarchiving.com/data-detect/">unifiedglobalarchiving.com/data-detect/</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>OrderMyGear</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/ordermygear.svg" alt="OrderMyGear"></td>
            <td><a href="https://www.ordermygear.com/">ordermygear.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Spiri.bo</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/spiribo.svg" alt="Spiri.bo"></td>
            <td><a href="https://spiri.bo/">spiri.bo</a></td>
        </tr>
    </tbody>
</table>

我们也非常感谢所有的个人贡献者

<a href="https://opencollective.com/ory" target="_blank"><img src="https://opencollective.com/ory/contributors.svg?width=890&button=false" /></a>

和我们所有的支持者。

<a href="https://opencollective.com/ory#backers" target="_blank"><img src="https://opencollective.com/ory/backers.svg?width=890"></a>

还有 [Patreon](https://www.patreon.com/_ory) 上过去和正在支持我们的人（按字母顺序）：Alexander Alimovs, Billy, Chancy
Kennedy, Drozzy, Edwin Trejos, Howard Edidin, Ken Adler Oz Haven, Stefan Hans,
TheCrealm。

<em>\* 表示在生产中使用ORY的主要项目之一。</em>

<!--END ADOPTERS-->


## 千里之行 始于足下

首先，请阅读 [ORY Kratos文档](https://www.ory.sh/kratos/docs) 。

### 快速开始

**[ORY Kratos快速入门](https://www.ory.sh/kratos/docs/quickstart)** 将教您ORY Kratos的基础知识，
并在不到五分钟的时间内建立一个基于 Docker Compose 的示例。

### 安装

请阅读 [ORY 开发人员文档](https://www.ory.sh/kratos/docs/install) ，以了解如何在 Linux，macOS，
Windows 和 Docker上安装ORY Kratos，以及如何从源代码构建 ORY Kratos。

### ORY Kratos的生态

<!--BEGIN ECOSYSTEM-->

当我们进行架构设计时，我们会基于以下几项指导原则来构建 Ory：

- 最小化依赖
- 处处可运行
- 扩大规模十分容易
- 最大限度地减少人为和网络错误的空间

ORY 架构设计的宗旨是能在**容器编排系统**（例如：Kubernetes，CloudFoundry，OpenShift 和类似项目）上最佳运行。
编译后的二进制可执行文件很小（5-15MB），
可用于所有流行的处理器类型（ARM，AMD64，i386）和操作系统（FreeBSD，Linux，macOS，Windows），
而且没有系统依赖性（Java，Node，Ruby，libxml等）。

### ORY Kratos：身份认证和用户管理系统

ORY Kratos 是根据 [云架构最佳实践](https://www.ory.sh/docs/ecosystem/software-architecture-philosophy)
构建的API优先的身份和用户管理系统。
它实现了几乎每一个软件应用程序都需要实现的核心使用功能，
登录和注册、登录和注册、多条件身份验证、
帐户验证和恢复、用户档案和帐户管理，
还有管理员级API。

### ORY Hydra：OAuth2 和 OpenID 联接的服务端

[ORY Hydra](https://github.com/ory/hydra) 是 OpenID 认证的 OAuth2 和 OpenID Connect 提供商，
可通过编写微型“桥”应用程序轻松连接到任何现有的身份系统。
绝对控制用户界面和用户体验流程。

### ORY OAuthkeeper：认证和接入代理

[ORY Oathkeeper](https://github.com/ory/oathkeeper) 是 BeyondCorp / Zero 信任身份和访问代理（IAP），
具有针对Web服务的可配置身份验证，授权和请求突变规则：验证JWT，访问令牌，API密钥，mTLS；
检查所包含的主题是否被允许执行请求；
将结果内容编码为自定义标头（X-User-ID），JSON Web令牌等等！
### ORY Keto：将接入控制策略作为服务
[ORY Keto](https://github.com/ory/keto) 是策略决策点。
它使用一组类似于AWS IAM策略的访问控制策略，
以确定是否授权某个主题（用户，应用程序，服务，汽车等）
对资源执行特定操作。

<!--END ECOSYSTEM-->

## 安全

运行身份认证基础设施需要[注意和了解威胁模型](https://www.ory.sh/kratos/docs/concepts/security) 。

### 揭露漏洞

如果您认为已发现安全漏洞，请不要在论坛，聊天室或GitHub上公开发布该漏洞，而应向我们发送电子邮件至 [hi@ory.am](mailto:hi@ory.sh) 。

## 匿名数据

对于 Ory 的服务收集汇总的匿名数据，你可以选择将其关闭。点击[这里](https://www.ory.sh/docs/ecosystem/sqa) 了解更多。

## 文档

### 指引

该指南可[在此](https://www.ory.sh/kratos/docs) 处获得。

### HTTP API 文档

HTTP API请访问[这里](https://www.ory.sh/kratos/docs/sdk/api) 。

### 更新和变更日志

新版本可能会引入重大更改。
为了帮助您识别和合并这些更改，
我们在 [UPGRADE.md](./UPGRADE.md)和 [CHANGELOG.md](./CHANGELOG.md) 中记录了这些更改。

### 命令行文档

运行 `kratos -h` 或 `kratos help`。

### 开发

我们鼓励所有贡献，并鼓励你阅读我们的[贡献准则](./CONTRIBUTING.md)。

#### 依赖

你需要使用 Go 1.13+ 并且 `GO111MODULE=on`，以正常编译和测试。

- Docker and Docker Compose
- Makefile
- NodeJS / npm

ORY Kratos 可以在 Windows PC 上开发，但请注意所有文档中的脚本均是类 UNIX 终端脚本，bash 或 zsh。

#### 从源码安装

```shell script
make install
```
#### 格式化代码

你可以运行 `make format` 命令来格式化代码，我们的 CI 也会自动检查你的代码格式是否正确。

#### 运行测试

你可以运行三种类型的测试：

- 简短测试（不需要像 PostgreSQL 这样的 SQL 数据库）
- 常规测试（需要 PostgreSQL，MySQL，CockroachDB）
- 端到端测试（需要数据库，并且将使用测试浏览器）

##### 短测试

简短测试运行得很快。你可以一次测试所有代码。

```shell script
go test -short -tags sqlite ./...
```
或者，仅测试指定的模块。

```shell script
cd client; go test -tags sqlite -short .
```

##### 常规测试

常规测试需要建立数据库。
我们的测试套件能够直接与 docker 一起使用（使用 [ory/dockertest](https://github.com/ory/dockertest) ），
但是我们鼓励使用 **Makefile** 代替。
使用 dockertest 可能会使系统上的 Docker 映像数量膨胀， 而且速度很慢。
相反，我们建议您执行以下操作：

<pre type="make/command">
make test
</pre>

请注意，每次运行 <code type="make/command">make test</code> 时，<code type="make/command">make test</code>
都会重新创建数据库。
如果您要修复非常具体的问题并需要一直进行数据库测试，这可能会很烦人。
在这种情况下，建议您使用以下方法初始化数据库：

<a type="make/command">

```shell script
make test-resetdb
export TEST_DATABASE_MYSQL='mysql://root:secret@(127.0.0.1:3444)/mysql?parseTime=true'
export TEST_DATABASE_POSTGRESQL='postgres://postgres:secret@127.0.0.1:3445/kratos?sslmode=disable'
export TEST_DATABASE_COCKROACHDB='cockroach://root@127.0.0.1:3446/defaultdb?sslmode=disable'
```

</a>

然后，你就可以运行 `go test` 命令了：

```shell script
go test -tags sqlite ./...

# 或者，在某个模块中:
cd client; go test  -tags sqlite  .
```
##### 端到端测试

我们使用 [Cypress](https://www.cypress.io) 来运行 e2e 测试。
你可以使用以下命令运行所有测试：

<pre type="make/command">
make test-e2e
</pre>

如果想要编写 e2e 测试，请运行以下命令以获取更多详细信息：

<pre type="make/command">
make docker
</pre>

#### 构建Docker

你可以使用以下命令构建开发Docker Image：

<pre type="make/command">
make docker
</pre>

#### 文档测试

要准备文档测试，请运行 `npm i` 以安装 [Text-Runner](https://github.com/kevgo/text-runner)。
- 测试所有文档：制作测试文档
- 测试单个文件：文本运行
