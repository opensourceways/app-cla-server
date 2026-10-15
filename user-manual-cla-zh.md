# CLA 系统用户手册

基于 [app-cla-server](https://github.com/opensourceways/app-cla-server) 实现贡献者许可协议（CLA）的在线签署与管理，后端使用 Go + Beego 框架。

---

## 角色与资源总览

系统围绕四类角色运转，每类角色的操作范围不同：

| 角色 | 英文标识 | 主要职责 |
|------|----------|----------|
| 社区管理员 | `owner of org` | 创建/删除 Link，管理 CLA 模板，添加企业管理员 |
| 企业管理员 | `corporation administrator` | 代表企业签署 CCLA，管理员工签署状态，添加员工管理员 |
| 员工管理员 | `employee manager` | 审批/停用员工签署记录 |
| 贡献者 | —— | 签署个人 CLA（ICLA）或企业员工 CLA |

系统核心资源：

| 资源 | 说明 |
|------|------|
| Link | 将一个 GitHub/Gitee 组织或仓库与一套 CLA 绑定的配置单元，是所有签署操作的入口 |
| CLA 模板 | 附属于 Link 的协议文件（PDF），分 `individual`（个人）和 `corporation`（企业）两类 |
| 企业签署记录（Corp Signing） | 企业管理员代表企业完成签署后生成的记录 |
| 个人签署记录（Individual Signing） | 个人贡献者签署 ICLA 后生成的记录 |
| 员工签署记录（Employee Signing） | 企业员工签署员工 CLA 后生成的记录，需企业管理员审批激活 |

---

## 接入流程图

```
                           开始
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│ 步骤 1: 确定签署类型                                     │
│ 你：判断贡献者身份——个人贡献者还是企业员工               │
│ 验收：明确走 ICLA 流程还是 CCLA + 员工 CLA 流程         │
└──────────────────────────┬──────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              │                           │
              ▼                           ▼
     ┌─────────────────┐       ┌─────────────────────┐
     │  个人贡献者      │       │  企业员工            │
     └────────┬────────┘       └──────────┬──────────┘
              │                           │
              ▼                           ▼
┌─────────────────────────┐  ┌────────────────────────────┐
│ 步骤 2A: 签署 ICLA       │  │ 步骤 2B: 企业签署 CCLA      │
│ 你：访问签署页面，填写    │  │ 你（企业管理员）：访问签署   │
│     信息，提交验证码      │  │     页面，填写企业信息       │
│ 验收：签署记录创建成功    │  │ 验收：企业签署记录创建成功   │
└────────────┬────────────┘  └────────────┬───────────────┘
             │                            │
             ▼                            ▼
┌─────────────────────────┐  ┌────────────────────────────┐
│ 步骤 3A: 验证签署状态    │  │ 步骤 3B: 员工签署员工 CLA   │
│ 你：提交 PR，系统自动    │  │ 你（员工）：访问签署页面，   │
│     检查签署状态         │  │     填写信息，提交验证码     │
│ 验收：PR 通过 CLA 检查   │  │ 验收：员工签署记录待审批     │
└─────────────────────────┘  └────────────┬───────────────┘
                                           │
                                           ▼
                             ┌────────────────────────────┐
                             │ 步骤 4B: 企业管理员审批      │
                             │ 你（企业管理员）：登录后台，  │
                             │     激活员工签署记录         │
                             │ 验收：员工签署状态变为激活    │
                             └────────────┬───────────────┘
                                           │
                                           ▼
                             ┌────────────────────────────┐
                             │ 步骤 5B: 验证签署状态        │
                             │ 你：提交 PR，系统自动检查    │
                             │ 验收：PR 通过 CLA 检查       │
                             └────────────────────────────┘
```

---

## 社区管理员操作

社区管理员负责初始化 CLA 接入配置。需要持有 `owner of org` 权限的 token。

### 创建 Link

Link 是 CLA 系统的核心配置单元，将代码平台的组织/仓库与 CLA 协议绑定。

**你需要做什么**：

向 `POST /v1/link/` 发送请求，携带以下参数：

```json
{
  "platform":     "github",
  "org_id":       "your-org-name",
  "repo_id":      "",
  "org_alias":    "Your Org Display Name",
  "org_logo":     "https://example.com/logo.png",
  "project_url":  "https://github.com/your-org"
}
```

- `platform`：代码平台，支持 `github` / `gitee`
- `org_id`：组织名称（GitHub/Gitee 上的 org name）
- `repo_id`：留空表示组织级 Link；填写仓库名表示仓库级 Link
- `org_alias`：展示给签署者的组织名称
- `project_url`：项目主页 URL，会出现在签署邮件中

**怎么验证这步完成**：

- 接口返回 `201`
- `GET /v1/link/` 列表中出现新创建的 Link 及其 `link_id`

---

### 上传 CLA 模板

每个 Link 需要至少一个 CLA 模板（PDF 文件），分个人和企业两类。

**你需要做什么**：

向 `POST /v1/cla/:link_id` 发送 multipart/form-data 请求：

```
link_id:   <上一步获取的 link_id>
apply_to:  individual   （或 corporation）
file:      <CLA PDF 文件>
fields:    <JSON 数组，定义签署者需填写的字段>
```

`fields` 示例（个人 CLA 常用字段）：

```json
[
  {"id": "name",    "title": "姓名",  "type": "string",  "required": true},
  {"id": "email",   "title": "邮箱",  "type": "string",  "required": true},
  {"id": "company", "title": "所在公司", "type": "string", "required": false}
]
```

**怎么验证这步完成**：

- 接口返回 `201`
- `GET /v1/cla/:link_id` 返回列表中包含刚上传的 CLA 模板及其 `cla_id`

---

### 添加企业管理员

企业完成 CCLA 签署后，社区管理员需要为该企业指定管理员账号。

**你需要做什么**：

向 `POST /v1/corporation-manager/:link_id/:signing_id` 发送请求（无需 body）。

- `signing_id`：企业签署记录 ID，从 `GET /v1/corporation-signing/:link_id` 获取

**我们做什么**：

- 系统自动生成企业管理员账号（初始密码）并发送邮件通知企业联系人

**怎么验证这步完成**：

- 接口返回 `202`
- 企业联系人收到包含账号和初始密码的邮件

---

## 企业管理员操作

企业管理员代表企业签署 CCLA，并管理本企业员工的签署状态。

### 签署企业 CLA（CCLA）

**你需要做什么**：

1. 获取验证码：向 `POST /v1/corporation-signing/:link_id/code` 发送请求：

```json
{"email": "corp-contact@example.com"}
```

2. 提交签署：向 `POST /v1/corporation-signing/:link_id/` 发送请求：

```json
{
  "cla_id":            "<cla_id>",
  "verification_code": "<收到的验证码>",
  "email":             "corp-contact@example.com",
  "corporation_name":  "Example Corp Ltd.",
  "name":              "张三",
  "address":           "北京市朝阳区...",
  "date":              "2026-05-13"
}
```

**我们做什么**：

- 验证验证码有效性
- 创建企业签署记录
- 生成签署 PDF 并发送至企业联系邮箱

**怎么验证这步完成**：

- 接口返回 `201`
- 企业联系邮箱收到签署确认邮件及 PDF 附件

---

### 登录企业管理员账号

社区管理员添加企业管理员后，企业管理员使用邮件中的账号登录。

**你需要做什么**：

向 `POST /v1/corporation-manager/` 发送请求：

```json
{
  "account":  "corp-contact@example.com",
  "password": "<邮件中的初始密码>"
}
```

**怎么验证这步完成**：

- 接口返回 `200`，响应中包含 `link_id` 和账号基本信息
- Cookie 中写入 `access_token` 和 `csrf_token`，后续请求携带这两个 token

---

### 管理员工签署记录

员工签署员工 CLA 后，企业管理员需要审批激活。

**你需要做什么**：

查看待审批员工列表：`GET /v1/employee-signing/`

激活员工签署：向 `PUT /v1/employee-signing/:signing_id` 发送请求：

```json
{"enabled": true}
```

停用员工签署：将 `enabled` 改为 `false`。

**我们做什么**：

- 激活后，系统向员工发送"CLA 签署已激活"通知邮件
- 停用后，系统向员工发送"CLA 签署已停用"通知邮件

**怎么验证这步完成**：

- 接口返回 `202`
- 员工收到状态变更通知邮件
- `GET /v1/employee-signing/` 列表中该员工状态更新

---

### 添加员工管理员

企业管理员可以将部分员工提升为员工管理员，代为审批其他员工。

**你需要做什么**：

向 `POST /v1/employee-manager/` 发送请求：

```json
{
  "managers": [
    {"email": "manager@example.com", "name": "李四"}
  ]
}
```

**我们做什么**：

- 系统向新员工管理员发送授权通知邮件，包含登录账号和初始密码

**怎么验证这步完成**：

- 接口返回 `201`
- 新员工管理员收到授权邮件
- `GET /v1/employee-manager/` 列表中出现新管理员

---

## 贡献者操作

### 签署个人 CLA（ICLA）

适用于以个人身份贡献代码、不代表任何企业的贡献者。

**你需要做什么**：

1. 获取签署页面信息（CLA 内容和字段定义）：

```
GET /v1/link/:link_id/individual
```

2. 获取验证码：

```json
POST /v1/individual-signing/:link_id/code
{"email": "contributor@example.com"}
```

3. 提交签署：

```json
POST /v1/individual-signing/:link_id/
{
  "cla_id":            "<cla_id>",
  "verification_code": "<验证码>",
  "email":             "contributor@example.com",
  "name":              "王五",
  "date":              "2026-05-13"
}
```

**怎么验证这步完成**：

- 接口返回 `201`
- `GET /v1/individual-signing/:link_id?email=contributor@example.com` 返回 `{"signed": true}`

---

### 签署员工 CLA

适用于所在企业已签署 CCLA、以企业员工身份贡献代码的贡献者。

**你需要做什么**：

1. 确认所在企业已完成 CCLA 签署，获取企业的 `signing_id`：

```
GET /v1/corporation-signing/:link_id/corps/:corp_email_domain
```

2. 获取验证码：

```json
POST /v1/employee-signing/:link_id/:signing_id/code
{"email": "employee@example-corp.com"}
```

3. 提交签署：

```json
POST /v1/employee-signing/:link_id/
{
  "cla_id":            "<cla_id>",
  "verification_code": "<验证码>",
  "email":             "employee@example-corp.com",
  "name":              "赵六",
  "date":              "2026-05-13"
}
```

**我们做什么**：

- 系统向员工发送签署确认邮件
- 系统向企业管理员发送"有员工待审批"通知邮件

**怎么验证这步完成**：

- 接口返回 `201`
- 等待企业管理员审批激活后，PR 可通过 CLA 检查

---

## 常见问题

**Q1：签署时提示 `resigned`，是什么意思？**

该邮箱已经签署过此 Link 下的 CLA，不能重复签署。如需更新签署信息，联系社区管理员处理。

**Q2：验证码提示过期或错误怎么办？**

验证码有效期较短（通常 10 分钟）。重新调用 `/code` 接口获取新验证码，再提交签署。

**Q3：员工签署后 PR 仍然显示未通过 CLA 检查？**

员工签署记录需要企业管理员审批激活（`enabled: true`）后才生效。确认企业管理员已完成审批。

**Q4：企业管理员账号忘记密码怎么办？**

访问 `POST /v1/password-retrieval/` 接口，通过注册邮箱重置密码。

**Q5：如何查看某个邮箱是否已签署 ICLA？**

```
GET /v1/individual-signing/:link_id?email=<邮箱>
```

返回 `{"signed": true}` 表示已签署。

**Q6：企业 CLA 签署后 PDF 没有收到怎么办？**

社区管理员可以调用 `PUT /v1/corporation-signing/:link_id/:signing_id` 重新触发 PDF 生成和邮件发送。

**Q7：`link_id` 从哪里获取？**

社区管理员登录后调用 `GET /v1/link/` 获取所有 Link 列表，其中包含每个 Link 的 `link_id`。

---

## 反馈与支持

遇到问题请在 GitHub 提交 Issue：

[https://github.com/opensourceways/app-cla-server/issues](https://github.com/opensourceways/app-cla-server/issues)

提交时请提供：
- 操作角色（社区管理员 / 企业管理员 / 贡献者）
- 调用的接口路径和请求体
- 返回的错误码和错误信息
- `link_id`（如涉及）
