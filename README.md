# cos-transmit

## 背景
我们基于[Minio](https://github.com/minio/)搭建了对象存储服务(cloud object storage service), 为云雀研发云用户提供标准的S3对象存储服务，并且在门户中提供了文件浏览、上传和下载等功能。但是我们很多用户在日常使用中，需要频繁从cos上下载或者同步本地目录，但是通过web页面方式无法直接传输文件夹，而且大文件传输性能也受限。用户可以使用[mc](https://github.com/minio/mc/)来解决这个问题，但是很多用户并不习惯(喜欢)使用命令行。目前能找到的方案中，很多都要收费，还有一些免费版有非常多的限制，比如S3Browser。

我们参考流行的ftp/sftp客户端工具[FileZilla](https://filezilla-project.org)的界面，利用[mc](https://github.com/minio/mc/)的后端能力，开发了一个的界面化客户端工具。

## 功能介绍
支持的功能如下：
- 按文件或者文件夹上传、下载
- 支持多选，批量添加任务
- 支持实时显示速度等汇总信息
- 支持随时暂停、继续、删除任务
- 支持自动跳过相同的文件(根据文件名和文件大小)
- 支持直接输入路径快速查找
- 支持拖动文件或者文件夹到窗口来上传
- 支持生成上传或者下载分享链接
- 支持创建文件夹
- 支持打开文件浏览器
- 支持中英文界面自动切换

由于使用了mc作为后端，因此mc的很多传输能力也都是继承过来的，比如既支持minio也支持通用的s3。

## 安装和配置
1. 下载  
本程序无需安装，直接从release中下载最新的zip包，解压到任意目录即可，zip包包含两个文件: cos-transmit.exe和cfg.yaml，直接双击运行即可。
3. 配置  
在使用之前请配置cfg.yaml,各个字段的配置说明见下
```yaml
cos:
- endpoint: "http://172.16.24.24:7000/public"  # 注意从minio内置页面复制过来的地址会是http://172.16.24.24:7000/minio/public，需要去掉中间的minio后配置到这里
  accesskey:  # 如果是公开的cos服务，这里留空即可。否则填写分配的ak和sk信息
  secretkey:  #
- endpoint: "http://172.16.24.24:7000/public/software"  # 可以为某个常用的cos地址配置一个单独的cos服务，比如这个点击后可以直接打开public这个bucket的software目录
  accesskey:  # 
  secretkey:  #
- endpoint: "http://172.16.24.24:7000"  # 如果拥有相应权限，可以直接配置根路径，访问所有bucket
  accesskey:  # 
  secretkey:  #
- endpoint: "https://play.minio.io" # 这个是官方的演示地址
  accesskey: Q3AM3UQ867SPQQA43P2F
  secretkey: zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG
#lang:  # en_US/zh_CN # 默认会根据操作系统语言自动切换, 可以在这里强制指定
#trace: # true/false # 用于调试，默认关闭
#ignore: # true/fase # 是否自动跳过同名同大小的文件传输，默认打开
#expiry: # 604800 # 生成分享链接的时，分享的超时时间，以秒为单位，默认值为604800，即7天
```

## 界面使用介绍

整体界面功能如下：
![image](https://user-images.githubusercontent.com/11539396/127418865-91380348-06aa-4f60-916e-a12585702cbd.png)

## 操作方法:
- 首先选中要传输的文件后，右键弹出相应菜单，执行对应操作即可  
![image](https://user-images.githubusercontent.com/11539396/127419740-662349d4-fd1c-4b6c-93da-3b5c623f4448.png)
- 当任务执行时可以选择相应的任务，右键弹出相应的菜单进行管理操作  
![image](https://user-images.githubusercontent.com/11539396/127420226-60443f89-bc59-4f8d-b586-428d70116f45.png)
- 可以直接在路径输入框输入具体地址后回车，快速打开相应目录
- 可以将需要上传的文件，直接拖动到程序窗口中来实现上传 

## 分享链接的说明:
- 当用户点击生成分享链接后，会自动将信息拷贝到粘贴板中
- 用户可以对某个文件生成一个下载分享链接，通过这个链接可以无须认证直接下载，格式如:  
`http://172.16.24.24:7000/public/software/oracle/instantclient-basiclite-linux.x64-21.1.0.0.0.zip?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=zcm-upload%2F20210806%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20210806T153846Z&X-Amz-Expires=604800&X-Amz-SignedHeaders=host&X-Amz-Signature=1b0a10ae023023eebb94dd299b2d5716fd197226a0070393d35dfca90dc1b2b9`
- 用户可以为目录或者文件生成上传分享链接，粘贴板自动复制了相应的curl命令，用户可以直接执行这个命令来实现无认证的上传(注意替换一下@<FILE>路径)，格式如:  
`curl http://172.16.24.24:7000/public/ -F bucket=public -F policy=eyJleHBpcmF0aW9uIjoiMjAyMS0wOC0xM1QxNTozODoxMC42MDlaIiwiY29uZGl0aW9ucyI6W1siZXEiLCIkYnVja2V0IiwicHVibGljIl0sWyJlcSIsIiRrZXkiLCJzb2Z0d2FyZSJdLFsiZXEiLCIkeC1hbXotZGF0ZSIsIjIwMjEwODA2VDE1MzgxMFoiXSxbImVxIiwiJHgtYW16LWFsZ29yaXRobSIsIkFXUzQtSE1BQy1TSEEyNTYiXSxbImVxIiwiJHgtYW16LWNyZWRlbnRpYWwiLCJ6Y20tdXBsb2FkLzIwMjEwODA2L3VzLWVhc3QtMS9zMy9hd3M0X3JlcXVlc3QiXV19 -F x-amz-algorithm=AWS4-HMAC-SHA256 -F x-amz-credential=zcm-upload/20210806/us-east-1/s3/aws4_request -F x-amz-date=20210806T153810Z -F x-amz-signature=cbf2d99ac92e21935a1f6149be1bc60c00bf73f9ca6338c01ee4ab17856cb1c3 -F key=software -F file=@<FILE>`

## 已知问题
- 使用everything搜索文件后，直接点击打开，会出现无法通过拖动文件来进行上传的问题，原因未知，怀疑是everything有什么冲突，可以先打开目录，再启动来规避。
- 目前上传文件时，无法自动探测content-type，目前提了issue到mc，等PutOptions可以引用后再增加支持。

### 致谢 
主要使用的开源库  
https://github.com/Minio  
https://github.com/lxn/walk  
https://filezilla-project.org
