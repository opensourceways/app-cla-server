---
title: "Sign CLA"
weight: 6
description: |
  An overview about CLA.
---

### 检查一个PR是否签署了CLA的原理

机器人检查CLA的粒度是commit。一个PR包含1个或多个commit。当PR的每个commit都通过了检查时，PR也通过了CLA检查。

将通过两个步骤来检查单个commit是否签署了CLA。第一步，获取commit作者的邮箱。第二步，在CLA系统中验证该email地址，如果验证通过，则证明该commit签署了CLA，否则未签署。


### 查看PR的commit信息

请访问此[API](https://docs.github.com/en/rest/reference/pulls#list-commits-on-a-pull-request)查看一个PR的所有commit，以及每个commit作者的邮箱。

commit作者的邮箱见下图。

![commit-author-email](commit_author_email.png)


### 当某个commit未通过CLA检查时，怎么处理

首先，请确保您完成了CLA签署。如果您签署的是员工CLA，请确保您的CLA Manager批准了您的CLA签署

其次，请确保每个commit中作者的邮箱是正确的，且该邮箱跟您签署时填写的邮箱是一致的

最后，如果您是如下的几种场景之一，请参考对应的方法处理。

#### 1. commit中配置的邮件地址是不正确的

不正确的邮件地址包括，邮件地址非法，邮件地址中少了一些字符等。如果您是通过`git push`命令提交的commit，可以参考如下方法进行修改。

假设一个PR的commit信息如下。

``` sh
# git log
commit ca82a6dff817ec66f44342007202690a93763949
Author: Scott Chacon <schacon@gee-mail.com>
Date:   Mon Mar 17 21:52:11 2008 -0700

    Change version number

commit 085bb3bcb608e1e8451d4b2432f8ecbe6306e7e7
Author: Scott Chacon <schacon@gee-mail.cm>
Date:   Sat Mar 15 16:40:33 2008 -0700

    Remove unnecessary test
```

在第二个commit中，其邮件地址的域名中少了一个字符o。可以通过如下命令修改第二个commit作者的邮箱

``` sh
git rebase -i HEAD~2

git commit --amend --author="Scott Chacon <schacon@gee-mail.com>" --no-edit

git rebase --continue
```

修改完后，请不要忘记使用`git push`命令提交commit。


#### 2. commit中配置的邮箱未签署CLA

请签署cla，签署时请填写此邮箱


#### 3. 通过Github网页提交的commit，其作者的邮箱是匿名邮箱

这是因为commit作者未公开其github 邮箱。请登陆您的Github账号，依次访问 setting -> Emails，配置"Primary email address"，不要勾选"Keep my email addresses private"。

针对这种commit，一种修改方法是，将PR拉取到本地，并参考第一种场景的方法修改commit的邮箱。

拉取PR到本地的方法是：

``` sh
git clone ${REPOSITORY_URL}
git fetch origin pull/${PULL_NUMBER}/head:${PULL_BASE_BRANCH}-${PULL_NUMBER}
```

PULL_BASE_BRANCH: PR目标分支的名称


### 怎么设置本地开发环境

在开发前请按如下方式配置git。这里的邮箱必须是签署CLA时填写的邮箱。

``` sh
git config user.name [GITHUB ID]

git config user.email [EMAIL]
```

如果要了解详细的配置本地开发环境的方法，这里是一个很好的[例子](https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md)
