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
- 支持直接输入路径
- 支持中英文切换

由于我们使用了mc后端，因此mc的很多传输能力也都是继承过来的，比如既支持minio也支持通用的s3。

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
# lang:  # en_US/zh_CN # 默认会根据操作系统语言自动切换, 可以在这里强制指定
# trace: # true/false # 用于调试，默认关闭
# ignore: # true/fase # 是否自动跳过同名同大小的文件传输，默认打开
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

### 致谢 
主要使用的开源库  
https://github.com/Minio  
https://github.com/lxn/walk  
https://filezilla-project.org
