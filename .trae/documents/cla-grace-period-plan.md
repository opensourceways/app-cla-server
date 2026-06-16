# CLA 协议变更平滑过渡方案

## 一、问题分析

### 当前流程的问题

```
CLA 协议更新 → 版本号递增 → 已签署员工 PR Check 返回 VersionMatched: false
                                    ↓
                              员工提交 PR 被阻塞 (cla-no 错误)
                                    ↓
                              员工找基础设施团队 → 基础设施团队联系企业法人
                                    ↓
                              企业法人可能未看到邮件通知 → 长时间阻塞
```

**核心痛点**：
1. CLA 更新后，企业员工的代码贡献流程被**立即阻塞**
2. 企业法人可能**忽略邮件通知**，导致阻塞时间不可控
3. 管理的社区数量多，基础设施团队需要**逐个排查和催促**，运维成本高

### 当前 Check API 逻辑

[individual_signing.go#L123-L164](file:///home/wuhejun/projects/github/app-cla-server/signing/app/individual_signing.go#L123-L164) 中，Check 方法对员工签署的版本匹配检查：

```go
// 员工签署检查
v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
if v.Enabled {
    dto.Signed = true
    dto.Type = "corp"
    dto.VersionMatched = s.cla.ContainsCla(cmd.LinkId, v.ClaId)  // ← 版本不匹配返回 false
    return
}
```

当 `VersionMatched == false` 时，下游 PR 检查服务会报 `cla-no` 错误，阻塞 PR 合并。

---

## 二、方案设计

### 核心思路：**非阻塞式版本检查 + 宽限期机制**

1. **CLA 更新后不立即阻塞任何人**：无论是企业员工还是个人签署者，签署过旧版本 CLA 仍然视为"版本匹配"，PR 可以正常通过
2. **记录待同意状态**：在 `corp_signing` 文档中新增 `pending_cla_id` 字段，标记企业法人尚未同意的新版本
3. **持续通知**：增强通知机制，定期提醒企业法人和个人签署者同意新版本
4. **可配置的宽限期**：支持按 Link 配置宽限期天数，宽限期过后可选择是否恢复阻塞
5. **提供管理面板**：基础设施团队可以查看所有待同意的企业列表，主动联系

### 方案架构图

```
CLA 协议更新
    │
    ├── 1. 更新 CLA 版本（现有逻辑不变）
    │
    ├── 2. 批量设置所有 corp_signing 的 pending_cla_id = 新版本 CLA ID
    │       └── 标记该 Link 下所有企业签署为"待同意新版本"
    │
    ├── 3. 发送通知（现有逻辑增强）
    │       ├── 企业法人：周期性邮件提醒
    │       └── 个人签署者：发送邮件提醒（新增）
    │
    └── 4. PR Check 逻辑变更
            │
            ├── 个人签署：版本不匹配
            │       │
            │       ├── 宽限期内 → VersionMatched: true（不阻塞 PR）
            │       │
            │       └── 超出宽限期 → 根据配置决定是否阻塞
            │
            └── 企业员工签署：版本不匹配
                    │
                    ├── 宽限期内 → VersionMatched: true（不阻塞 PR）
                    │
                    └── 超出宽限期 → 根据配置决定是否阻塞
```

---

## 三、详细实施步骤

### 步骤 1：数据库 Schema 变更

#### 1.1 `corp_signing` 集合新增字段

在 [corp_signing_do.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/repositoryimpl/corp_signing_do.go) 的 `corpSigningDO` 结构体中新增：

```go
type corpSigningDO struct {
    // ... 现有字段保持不变 ...
    
    PendingCLAId string `bson:"pending_cla_id"` // 待企业法人同意的新 CLA 版本 ID，为空表示无待同意版本
}
```

#### 1.2 `link` 集合新增字段

在 [link_do.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/repositoryimpl/link_do.go) 的 `linkDO` 结构体中新增：

```go
type linkDO struct {
    // ... 现有字段保持不变 ...
    
    GracePeriodDays int `bson:"grace_period_days"` // CLA 更新宽限期天数，0 表示永不阻塞，-1 表示使用全局默认值
}
```

#### 1.3 全局配置新增

在配置文件中新增全局默认宽限期配置：

```yaml
cla:
  default_grace_period_days: 30  # 默认宽限期 30 天，0 表示永不阻塞
```

### 步骤 2：领域模型变更

#### 2.1 `CorpSigning` 领域对象

在 [corp_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/corp_signing.go) 中：

- 新增 `PendingCLAId` 字段
- 新增 `SetPendingCLA(newClaId string)` 方法：设置待同意版本
- 新增 `ClearPendingCLA()` 方法：清除待同意状态（法人同意后调用）
- 新增 `HasPendingCLA() bool` 方法：判断是否有待同意版本
- 修改 `SetLatestClaId()` 方法：同意新版本时同时清除 `PendingCLAId`

#### 2.2 `Link` 领域对象

在 [link.go](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/link.go) 中：

- 新增 `GracePeriodDays` 字段
- 新增 `IsGracePeriodEnabled() bool` 方法：判断是否启用宽限期
- 新增 `GetEffectiveGracePeriodDays(defaultDays int) int` 方法：获取生效的宽限期天数

#### 2.3 `IndividualSignedDTO` 响应结构

在 [individual_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/app/individual_signing.go) 中：

```go
type IndividualSignedDTO struct {
    Type           string `json:"type"`
    Signed         bool   `json:"signed"`
    VersionMatched bool   `json:"version_matched"`
    PendingVersion bool   `json:"pending_version"` // 新增：是否存在待同意的新版本
}
```

### 步骤 3：CLA 更新流程变更

#### 3.1 更新 CLA 时批量设置 pending_cla_id

在 [service.go](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/claservice/service.go) 的 `Update` 方法中，CLA 更新成功后：

1. 发送 `CLAUpdatedMsg` 事件（现有逻辑）
2. 新增：调用 `CorpSigning` 仓储的 `SetPendingCLAForLink(linkId, newClaId)` 方法，批量将该 Link 下所有企业签署的 `pending_cla_id` 设置为新版本 ID

#### 3.2 仓储层新增方法

在 [repository/corp_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/repository/corp_signing.go) 接口中新增：

```go
type CorpSigning interface {
    // ... 现有方法 ...
    
    SetPendingCLAForLink(linkId, newClaId string) error  // 批量设置待同意版本
    FindPendingAgreements(linkId string) ([]CorpSigningSummary, error)  // 查询待同意列表
}
```

在 [corp_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/repositoryimpl/corp_signing.go) 中实现：

```go
func (impl *corpSigning) SetPendingCLAForLink(linkId, newClaId string) error {
    filter := bson.M{
        fieldLinkId: linkId,
        fieldHasPDF: true,  // 只更新已上传 PDF 的企业（正式签署的）
    }
    update := bson.M{
        "$set": bson.M{fieldPendingCLAId: newClaId},
    }
    return impl.dao.UpdateDocsWithoutVersion(filter, update)
}
```

### 步骤 4：Check API 逻辑变更

#### 4.1 核心变更

修改 [individual_signing.go#L123-L164](file:///home/wuhejun/projects/github/app-cla-server/signing/app/individual_signing.go#L123-L164) 的 `Check` 方法：

```go
func (s *individualSigningService) Check(cmd *CmdToCheckSinging) (dto IndividualSignedDTO, err error) {
    // 个人签署检查
    claId, err := s.repo.FindSignedCLA(cmd.LinkId, cmd.EmailAddr)
    if claId != "" {
        dto.Signed = true
        dto.Type = dp.CLATypeIndividual.CLAType()
        versionMatched := s.cla.ContainsCla(cmd.LinkId, claId)
        
        if versionMatched {
            dto.VersionMatched = true
        } else {
            // 个人签署者版本不匹配时，同样走宽限期逻辑
            if s.isInGracePeriod(cmd.LinkId) {
                dto.VersionMatched = true  // 宽限期内不阻塞
                dto.PendingVersion = true
            } else {
                dto.VersionMatched = false // 超出宽限期，阻塞
                dto.PendingVersion = true
            }
        }
        return
    }

    // 员工签署检查
    v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
    if v.Enabled {
        dto.Signed = true
        dto.Type = dp.CLATypeCorp.CLAType()
        
        versionMatched := s.cla.ContainsCla(cmd.LinkId, v.ClaId)
        
        if versionMatched {
            dto.VersionMatched = true
        } else {
            if s.isInGracePeriod(cmd.LinkId) {
                dto.VersionMatched = true
                dto.PendingVersion = true
            } else {
                dto.VersionMatched = false
                dto.PendingVersion = true
            }
        }
        return
    }
    
    // ... 未签署时的逻辑不变 ...
}
```

> **设计说明**：个人签署者和企业员工签署者共享同一套宽限期逻辑。两者在宽限期内都不会被阻塞，超出宽限期后的行为也一致。这样设计的原因是：
> 1. 个人签署者和企业员工都是社区的贡献者，应该享受同等的待遇
> 2. 个人签署者没有"企业法人"这个中间层，他们自己就是决策者，通知他们比通知企业法人更直接
> 3. 统一的逻辑降低代码复杂度和维护成本

#### 4.2 宽限期判断逻辑

```go
func (s *individualSigningService) isInGracePeriod(linkId string) bool {
    link, err := s.linkRepo.Find(linkId)
    if err != nil {
        return true // 查询失败时默认不阻塞，安全优先
    }
    
    days := link.GetEffectiveGracePeriodDays(s.defaultGracePeriodDays)
    if days == 0 {
        return true // 0 表示永不阻塞
    }
    
    // 检查 CLA 最后更新时间是否在宽限期内
    lastUpdateTime := s.cla.GetLastUpdateTime(linkId)
    return time.Since(lastUpdateTime) < time.Duration(days)*24*time.Hour
}
```

### 步骤 5：通知机制增强

#### 5.1 企业法人通知频率增强

在 [notify_corp_admin.go](file:///home/wuhejun/projects/github/app-cla-server/signing/watch/notify_corp_admin.go) 中：

- 当前：只通知一次（通过 `cla_notify` 字段去重）
- 增强：支持**周期性重复通知**，例如每 7 天提醒一次

新增字段 `cla_notify_count` 和 `cla_notify_time`：

```go
type corpSigningDO struct {
    // ...
    ClaNotify      string `bson:"cla_notify"`       // 已通知的 CLA 版本
    ClaNotifyCount int    `bson:"cla_notify_count"`  // 通知次数
    ClaNotifyTime  int64  `bson:"cla_notify_time"`   // 上次通知时间戳
}
```

#### 5.2 企业法人通知逻辑变更

```go
func (impl *notifyAdminWatchImpl) handleCorpSigning(link *repository.LinkCLA, corp *repository.CorpSigningSummary) {
    // 检查是否有待同意版本
    if corp.PendingCLAId == "" {
        return
    }
    
    // 检查是否需要重新通知（超过 7 天）
    if corp.CLANotify == corp.PendingCLAId {
        if time.Since(time.Unix(corp.CLANotifyTime, 0)) < 7*24*time.Hour {
            return // 7 天内已通知过，跳过
        }
    }
    
    // 发送邮件通知
    impl.handleSendEmail(link, corp)
    
    // 更新通知记录
    corp.CLANotify = corp.PendingCLAId
    corp.CLANotifyCount += 1
    corp.CLANotifyTime = time.Now().Unix()
    impl.corpSigningRepo.UpdateCLANotify(corp)
}
```

#### 5.3 个人签署者通知机制（新增）

**背景**：个人签署者没有企业法人这个中间层，他们自己就是决策者。CLA 更新后，需要通知个人签署者去同意新版本。当前系统**完全没有**对个人签署者的 CLA 更新通知。

**实现方案**：复用现有的 `notifyAdminWatchImpl` 机制，在 `handleNotifyJob` 中增加对个人签署者的通知逻辑。

**新增仓储方法**：

在 [repository/individual_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/repository/individual_signing.go) 接口中新增：

```go
type IndividualSigning interface {
    // ... 现有方法 ...
    
    FindAllWithPagination(linkId string, offset, limit int) ([]domain.IndividualSigning, error)
    UpdateCLANotify(linkId, email string, claId string, count int, notifyTime int64) error
}
```

**新增 `individual_signing` DO 字段**：

在 [individual_signing_do.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/repositoryimpl/individual_signing_do.go) 中：

```go
type individualSigningDO struct {
    // ... 现有字段保持不变 ...
    
    ClaNotify      string `bson:"cla_notify"`       // 新增：已通知的 CLA 版本
    ClaNotifyCount int    `bson:"cla_notify_count"`  // 新增：通知次数
    ClaNotifyTime  int64  `bson:"cla_notify_time"`   // 新增：上次通知时间戳
}
```

**通知逻辑**（在 `notifyCorpAdmin` goroutine 中新增）：

```go
func (impl *notifyAdminWatchImpl) handleNotifyJob() {
    // ... 现有企业通知逻辑 ...
    
    // 新增：个人签署者通知
    for i := range links {
        link := &links[i]
        impl.notifyIndividualSigners(link)
    }
}

func (impl *notifyAdminWatchImpl) notifyIndividualSigners(link *repository.LinkCLA) {
    offset := 0
    limit := 100
    
    for {
        signers, err := impl.individualSigningRepo.FindAllWithPagination(link.Id, offset, limit)
        if err != nil || len(signers) == 0 {
            break
        }
        
        for _, signer := range signers {
            // 版本已匹配，跳过
            if impl.cla.ContainsCla(link.Id, signer.Link.CLAId) {
                continue
            }
            
            // 已通知过且未到重复提醒时间，跳过
            if signer.ClaNotify == impl.getLatestClaId(link) {
                if time.Since(time.Unix(signer.ClaNotifyTime, 0)) < 7*24*time.Hour {
                    continue
                }
            }
            
            // 发送邮件
            impl.sendIndividualSignerEmail(link, &signer)
            
            // 更新通知记录
            impl.individualSigningRepo.UpdateCLANotify(
                link.Id, signer.Rep.EmailAddr.EmailAddr(),
                impl.getLatestClaId(link),
                signer.ClaNotifyCount + 1,
                time.Now().Unix(),
            )
        }
        
        offset += limit
    }
}
```

**个人签署者邮件模板**：新增 `conf/email-template/cla-updated-individual.tmpl`：

```
Subject: {{.Org}} CLA 协议已更新 - 无需立即操作，您的贡献不受影响

Dear {{.Name}},

您好！{{.Org}} 社区的贡献者许可协议（CLA）已于 {{.UpdateDate}} 进行了更新。

📌 这对您意味着什么？

  您此前已签署过 CLA，目前仍然可以正常提交代码，贡献流程不会受到任何影响。
  我们采用了平滑过渡机制，确保不会阻塞您的日常开发工作。

📋 您需要做什么？

  在您方便的时候，请登录 CLA 管理系统查看并确认新的协议版本。
  这不是紧急事项，但我们建议您在 30 天内完成确认，以确保协议信息保持最新。

  👉 CLA 管理系统地址：{{.URLOfCLAPlatform}}

💡 如有任何疑问，直接回复此邮件即可，{{.Org}} 社区支持团队将很乐意为您提供帮助。

感谢您对 {{.Org}} 社区的支持与贡献！

---

{{.Org}} Community Support Team
{{.ProjectURL}}
```

**邮件模板数据结构**（在 [template.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/emailtmpl/template.go) 中新增）：

```go
type IndividualCLAUpdated struct {
    Org              string
    Name             string
    UpdateDate       string
    ProjectURL       string
    URLOfCLAPlatform string
}

func (data *IndividualCLAUpdated) GenEmailMsg() (EmailMessage, error) {
    return genEmailMsg(TmplIndividualCLAUpdated, data)
}
```

#### 5.4 个人签署者同意新版本流程（现有逻辑不变）

个人签署者已有的 `AgreeNewCLA` API 保持不变：

```
PUT /v1/individual-signing/:link_id/
```

当个人签署者同意新版本时：
1. 更新 `cla_id` 为最新版本（现有逻辑，见 [individual_signing.go#L77-L97](file:///home/wuhejun/projects/github/app-cla-server/signing/app/individual_signing.go#L77-L97)）
2. 追加 agree 日志（现有逻辑）
3. 重置 `cla_notify` 相关字段（新增逻辑）

### 步骤 6：新增管理 API

#### 6.1 查询待同意列表

```
GET /v1/corporation-signing/pending-agreements?link_id=xxx
```

返回该 Link 下所有 `pending_cla_id` 不为空的企业签署记录，方便基础设施团队排查。

**Controller**: 在 [corporation_signing.go](file:///home/wuhejun/projects/github/app-cla-server/controllers/corporation_signing.go) 中新增方法。

**响应格式**：
```json
{
  "data": [
    {
      "signing_id": "xxx",
      "corp_name": "示例公司",
      "admin_email": "admin@example.com",
      "signed_cla_version": "3",
      "pending_cla_version": "5",
      "notify_count": 2,
      "last_notify_time": "2026-06-10T10:00:00Z"
    }
  ]
}
```

#### 6.2 更新 Link 宽限期配置

```
PUT /v1/link/:link_id/grace-period
Body: { "grace_period_days": 30 }
```

允许社区管理员配置该 Link 的宽限期天数。

### 步骤 7：企业法人同意流程变更

现有的 `AgreeWithLatestCLA` API 保持不变：

```
PUT /v1/corporation-signing/cla/agree
```

当企业法人同意新版本时：
1. 更新 `cla_id` 为最新版本（现有逻辑）
2. 清除 `pending_cla_id`（新增逻辑）
3. 重置 `cla_notify` 相关字段（新增逻辑）
4. **追加审计日志**（新增，见步骤 9）

### 步骤 8：邮件模板优化

#### 8.1 当前邮件模板的问题

当前 CLA 更新通知邮件 [cla-updated.tmpl](file:///home/wuhejun/projects/github/app-cla-server/conf/email-template/cla-updated.tmpl) 内容：

```
Dear {{.AdminName}},

A new version of CLA is published. Please login to the CLA management system to see the details.

The CLA management system login URL is {{.URLOfCLAPlatform}}.

Have questions or need help? Just reply to this email and the {{.Org}} Community Support Team will help you sort it out.

[1]. {{.ProjectURL}}
```

**存在的问题**：
- 语气生硬，"A new version of CLA is published" 像系统通知，缺乏人情味
- 没有说明**为什么**需要企业法人关注这件事
- 没有说明**不处理会有什么影响**（宽限期机制下实际不影响员工，但需要告知）
- 没有说明**变更了什么**，需要法人自己去系统里查看
- 没有体现对法人时间的尊重和感谢

#### 8.2 优化后的邮件模板

**新增模板文件** `conf/email-template/cla-updated.tmpl`（覆盖原文件）：

```
Subject: {{.Org}} CLA 协议已更新 - 无需立即操作，您的团队成员贡献不受影响

Dear {{.AdminName}},

您好！{{.Org}} 社区的贡献者许可协议（CLA）已于 {{.UpdateDate}} 进行了更新。

📌 这对您和您的团队意味着什么？

  您的企业「{{.CorpName}}」此前已签署过 CLA，团队成员目前仍然可以正常提交代码，
  贡献流程不会受到任何影响。我们采用了平滑过渡机制，确保不会阻塞您团队的日常开发工作。

📋 您需要做什么？

  在您方便的时候，请登录 CLA 管理系统查看并确认新的协议版本。
  这不是紧急事项，但我们建议您在 30 天内完成确认，以确保协议信息保持最新。

  👉 CLA 管理系统地址：{{.URLOfCLAPlatform}}

💡 如有任何疑问，直接回复此邮件即可，{{.Org}} 社区支持团队将很乐意为您提供帮助。

感谢您对 {{.Org}} 社区的支持与贡献！

---

{{.Org}} Community Support Team
{{.ProjectURL}}
```

#### 8.3 邮件模板数据结构变更

在 [template.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/emailtmpl/template.go) 中，`CLAUpdated` 结构体新增字段：

```go
type CLAUpdated struct {
    Org              string
    CorpName         string  // 新增：企业名称
    AdminName        string
    UpdateDate       string  // 新增：CLA 更新日期
    ProjectURL       string
    URLOfCLAPlatform string
}
```

#### 8.4 邮件发送逻辑变更

在 [notify_corp_admin.go](file:///home/wuhejun/projects/github/app-cla-server/signing/watch/notify_corp_admin.go) 的 `handleSendEmail` 方法中，传入更多上下文数据：

```go
func (impl *notifyAdminWatchImpl) handleSendEmail(link *repository.LinkCLA, corp *repository.CorpSigningSummary) error {
    if corp.Admin.Name == nil {
        return fmt.Errorf("failed to send email msg: admin name is null: %s", link.Id)
    }
    builder := emailtmpl.CLAUpdated{
        Org:              link.Org.Alias,
        CorpName:         corp.Corp.Name,          // 新增：企业名称
        AdminName:        corp.Admin.Name.Name(),
        UpdateDate:       time.Now().Format("2006-01-02"), // 新增：更新日期
        ProjectURL:       link.Org.ProjectURL,
        URLOfCLAPlatform: impl.claPlatformURL + link.Id,
    }
    emailMsg, err := builder.GenEmailMsg()
    if err != nil {
        return err
    }

    emailMsg.From = link.Email.Addr.EmailAddr()
    emailMsg.To = []string{corp.Admin.EmailAddr.EmailAddr()}
    emailMsg.Subject = fmt.Sprintf("%s CLA 协议已更新", link.Org.Alias)

    worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

    time.Sleep(impl.config.genSendEmailInterval())

    return nil
}
```

#### 8.5 邮件模板设计原则

| 原则 | 说明 |
|------|------|
| **消除焦虑** | 开头就告知"团队成员贡献不受影响"，避免法人看到邮件后紧张 |
| **降低行动门槛** | 明确说"不是紧急事项"、"在您方便的时候"，尊重法人的时间 |
| **告知价值** | 说明为什么需要确认（保持协议信息最新），而不是命令式要求 |
| **人性化语气** | 使用"您好"、"感谢"等礼貌用语，加入 emoji 增加亲和力 |
| **清晰的行动指引** | 提供明确的链接和操作说明 |
| **提供支持渠道** | 告知可以直接回复邮件获取帮助 |

### 步骤 9：法律合规审计日志（新增）

#### 9.1 当前审计日志现状

| 签署类型 | 审计日志 | 状态 |
|----------|----------|------|
| **个人签署者** | `logs` 数组，记录 `sign`/`agree` 动作、日期、版本号 | ✅ 已有 |
| **企业签署者** | 无 | ❌ **完全缺失！** |

**当前企业签署者 `SetLatestClaId` 的问题** — [corp_signing.go#L252-L260](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/corp_signing.go#L252-L260)：

```go
func (cs *CorpSigning) SetLatestClaId(latestClaId string) error {
    if cs.HasSignedCLA(latestClaId) {
        return NewDomainError(ErrorCodeCorpSigningCLAIsLatest)
    }
    cs.Link.CLAId = latestClaId  // 直接覆盖，没有任何日志记录！
    return nil
}
```

**法律风险**：如果未来发生法律纠纷，无法证明企业法人是在什么时间、同意了哪个版本的 CLA 协议变更。`cla_id` 字段被直接覆盖，历史版本信息丢失。

#### 9.2 审计日志设计方案

参照个人签署者已有的 `logs` 数组设计，为企业签署者也增加相同的审计日志机制。

**`corp_signing` DO 新增字段**：

```go
type corpSigningDO struct {
    // ... 现有字段保持不变 ...
    
    PendingCLAId string `bson:"pending_cla_id"` // 待同意版本
    ClaNotify      string `bson:"cla_notify"`
    ClaNotifyCount int    `bson:"cla_notify_count"`
    ClaNotifyTime  int64  `bson:"cla_notify_time"`
    
    Logs []corpSigningLogDO `bson:"logs"` // 新增：审计日志
}
```

**审计日志子文档结构**：

```go
type corpSigningLogDO struct {
    Date   string `bson:"date"   json:"date"   required:"true"` // 操作时间，格式 "2006-01-02"
    CLAId  string `bson:"cla_id" json:"cla_id" required:"true"` // 操作的 CLA 版本 ID
    Action string `bson:"action" json:"action" required:"true"` // 操作类型
}
```

**操作类型常量**：

```go
const (
    corpSigningActionSign  = "sign"  // 首次签署
    corpSigningActionAgree = "agree" // 同意新版本
)
```

#### 9.3 领域层变更

在 [corp_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/domain/corp_signing.go) 中：

```go
type CorpSigning struct {
    // ... 现有字段保持不变 ...
    PendingCLAId string
    
    Logs []CorpSigningLog  // 新增：审计日志
}

type CorpSigningLog struct {
    Date   string
    CLAId  string
    Action string
}

// 修改 SetLatestClaId，追加审计日志
func (cs *CorpSigning) SetLatestClaId(latestClaId string) error {
    if cs.HasSignedCLA(latestClaId) {
        return NewDomainError(ErrorCodeCorpSigningCLAIsLatest)
    }

    cs.Link.CLAId = latestClaId
    cs.PendingCLAId = ""  // 清除待同意状态
    
    // 追加审计日志
    cs.Logs = append(cs.Logs, CorpSigningLog{
        Date:   util.Date(),
        CLAId:  latestClaId,
        Action: corpSigningActionAgree,
    })

    return nil
}

// 首次签署时初始化日志
func NewCorpSigning(...) CorpSigning {
    cs := CorpSigning{...}
    cs.Logs = []CorpSigningLog{
        {
            Date:   cs.Date,
            CLAId:  cs.Link.CLAId,
            Action: corpSigningActionSign,
        },
    }
    return cs
}
```

#### 9.4 仓储层变更

**DO 转换** — [corp_signing_do.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/repositoryimpl/corp_signing_do.go)：

```go
// 新增日志 DO
type CorpSigningLogsDO struct {
    Logs []corpSigningLogDO `bson:"logs" json:"logs"`
}

type corpSigningLogDO struct {
    Date   string `bson:"date"   json:"date"`
    CLAId  string `bson:"cla_id" json:"cla_id"`
    Action string `bson:"action" json:"action"`
}

// toCorpSigningDO 转换时包含日志
func toCorpSigningDO(v *domain.CorpSigning) corpSigningDO {
    return corpSigningDO{
        // ... 现有字段 ...
        Logs: toCorpSigningLogsDO(v.Logs),
    }
}

// toCorpSigning 转换时还原日志
func (do *corpSigningDO) toCorpSigning() domain.CorpSigning {
    return domain.CorpSigning{
        // ... 现有字段 ...
        Logs: do.toCorpSigningLogs(),
    }
}
```

**UpdateClaId 仓储方法变更** — [corp_signing.go](file:///home/wuhejun/projects/github/app-cla-server/signing/infrastructure/repositoryimpl/corp_signing.go)：

```go
func (impl *corpSigning) UpdateClaId(cs *domain.CorpSigning) error {
    filter, _ := impl.toCorpSigningIndex(cs.Id)
    
    // 更新 cla_id、pending_cla_id、logs
    update := bson.M{
        "$set": bson.M{
            fieldCLAId:       cs.Link.CLAId,
            fieldPendingCLAId: "",        // 清除待同意状态
        },
        "$push": bson.M{
            fieldLogs: toCorpSigningLogDO(corpSigningLogDO{
                Date:   util.Date(),
                CLAId:  cs.Link.CLAId,
                Action: "agree",
            }),
        },
    }
    
    return impl.dao.UpdateDoc(filter, update, cs.Version)
}
```

#### 9.5 个人签署者日志增强

个人签署者已有日志机制，但需要在 `AgreeNewCLA` 时同时重置通知状态：

```go
func (i *IndividualSigning) AgreeNewCLA(claId string) error {
    if i.Link.CLAId == claId {
        return NewDomainError(ErrorCodeIndividualSigningCLAIsLatest)
    }

    i.Link.CLAId = claId
    i.ClaNotify = ""       // 新增：重置通知状态
    i.ClaNotifyCount = 0   // 新增：重置通知计数
    i.ClaNotifyTime = 0    // 新增：重置通知时间
    
    i.addLogOfAgreeingNewCLA()  // 现有：追加 agree 日志
    
    return nil
}
```

#### 9.6 审计日志查询

在 `CorpSigningInfoDTO` 中返回审计日志，方便前端展示：

```go
type CorpSigningInfoDTO struct {
    // ... 现有字段 ...
    Logs []CorpSigningLogDTO `json:"logs"` // 新增：审计日志
}

type CorpSigningLogDTO struct {
    Date   string `json:"date"`
    CLAId  string `json:"cla_id"`
    Action string `json:"action"`
}
```

#### 9.7 审计日志示例

MongoDB 中 `corp_signing` 文档的 `logs` 数组示例：

```json
{
  "logs": [
    {
      "date": "2025-01-15",
      "cla_id": "3",
      "action": "sign"
    },
    {
      "date": "2026-03-20",
      "cla_id": "4",
      "action": "agree"
    },
    {
      "date": "2026-06-10",
      "cla_id": "5",
      "action": "agree"
    }
  ]
}
```

每条日志清晰记录了：
- **何时**（`date`）：操作发生的日期
- **哪个版本**（`cla_id`）：同意的是哪个 CLA 版本
- **什么操作**（`action`）：`sign`（首次签署）或 `agree`（同意变更）

这为未来可能的法律纠纷提供了完整的审计追踪链。

---

## 四、实施文件清单

| 序号 | 文件 | 变更类型 | 说明 |
|------|------|----------|------|
| 1 | `signing/infrastructure/repositoryimpl/corp_signing_do.go` | 修改 | 新增 `pending_cla_id`、`cla_notify_count`、`cla_notify_time`、`logs` 字段 |
| 2 | `signing/infrastructure/repositoryimpl/link_do.go` | 修改 | 新增 `grace_period_days` 字段 |
| 3 | `signing/infrastructure/repositoryimpl/individual_signing_do.go` | 修改 | 新增 `cla_notify`、`cla_notify_count`、`cla_notify_time` 字段 |
| 4 | `signing/domain/corp_signing.go` | 修改 | 新增 `PendingCLAId`、`Logs` 审计日志及相关方法；`SetLatestClaId` 追加日志 |
| 5 | `signing/domain/link.go` | 修改 | 新增 `GracePeriodDays` 及相关方法 |
| 6 | `signing/domain/individual_signing.go` | 修改 | 新增 `ClaNotify` 相关字段，AgreeNewCLA 时重置通知状态 |
| 7 | `signing/domain/repository/corp_signing.go` | 修改 | 新增仓储接口方法 |
| 8 | `signing/domain/repository/individual_signing.go` | 修改 | 新增 `UpdateCLANotify` 仓储接口方法 |
| 9 | `signing/infrastructure/repositoryimpl/corp_signing.go` | 修改 | 实现新增仓储方法；`UpdateClaId` 追加审计日志 |
| 10 | `signing/infrastructure/repositoryimpl/individual_signing.go` | 修改 | 实现 `UpdateCLANotify` 仓储方法 |
| 11 | `signing/domain/claservice/service.go` | 修改 | CLA 更新时批量设置 pending_cla_id |
| 12 | `signing/app/individual_signing.go` | 修改 | Check 方法增加宽限期逻辑（个人+企业）；AgreeNewCLA 重置通知状态 |
| 13 | `signing/app/corp_signing.go` | 修改 | AgreeWithLatestCLA 清除 pending_cla_id + 追加审计日志；DTO 返回 logs |
| 14 | `signing/watch/notify_corp_admin.go` | 修改 | 增强通知机制：周期性提醒 + 个人签署者通知 + 邮件数据丰富 |
| 15 | `controllers/corporation_signing.go` | 修改 | 新增待同意列表查询 API |
| 16 | `controllers/link.go` | 修改 | 新增宽限期配置 API |
| 17 | `config/config.go` | 修改 | 新增全局默认宽限期配置 |
| 18 | `signing/adapter/individual_signing.go` | 修改 | 适配 Check 方法的新返回值 |
| 19 | `signing/adapter/corp_signing.go` | 修改 | 适配新增 API |
| 20 | `signing/infrastructure/emailtmpl/template.go` | 修改 | `CLAUpdated` 新增 `CorpName`、`UpdateDate`；新增 `IndividualCLAUpdated` 结构体 |
| 21 | `conf/email-template/cla-updated.tmpl` | 修改 | 重写企业法人邮件模板，语气更柔和、信息更丰富 |
| 22 | `conf/email-template/cla-updated-individual.tmpl` | 新增 | 个人签署者 CLA 更新通知邮件模板 |

---

## 五、配置示例

```yaml
cla:
  default_grace_period_days: 30  # 默认宽限期 30 天
  notify:
    interval_seconds: 1200       # 通知检查间隔（现有）
    remind_interval_days: 7      # 重复提醒间隔（新增）
```

---

## 六、风险与注意事项

1. **向后兼容**：新增字段均设置默认值，不影响现有数据
2. **性能考虑**：
   - `SetPendingCLAForLink` 批量更新操作需要考虑大数据量场景，建议使用 MongoDB 的 `updateMany`
   - 个人签署者通知采用分页遍历（每页 100 条），避免一次性加载大量数据导致内存问题
   - 邮件发送有间隔控制（`genSendEmailInterval`），防止邮件服务器拒绝服务
3. **缓存失效**：修改 `corp_signing` 后需要使 Redis 缓存失效
4. **个人签署者通知**：
   - 个人签署者数量可能很大，通知逻辑需要分页处理
   - 个人签署者通知与企业法人通知共享同一个 goroutine，注意不要相互阻塞
   - 个人签署者同意新版本后，需要重置 `cla_notify` 相关字段，避免继续收到重复通知
5. **宽限期过期后的行为**：建议默认配置 `grace_period_days: 0`（永不阻塞），让各社区按需配置
6. **通知去重**：`cla_notify` 字段存储的是最新 CLA 版本 ID，当 CLA 再次更新时，即使上次通知的版本不同，也会重新触发通知
