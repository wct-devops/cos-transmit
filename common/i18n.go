package common

import (
	"golang.org/x/text/message"
	"runtime"
	"os"
	"golang.org/x/text/language"
	"strings"
)

var I18n *message.Printer
var lang []string

func InitI18nPrinter(defaultLang string){
	if len(defaultLang) > 0 {
		lang = []string{defaultLang}
	} else {
		if runtime.GOOS != "windows" && len(os.Getenv("LANG")) > 0  {
			lang = append(lang, os.Getenv("LANG"))
		}
	}
	if len(lang) < 1 {
		lang = []string{"en_US"}
	}
	var matcher = language.NewMatcher([]language.Tag{ language.English, language.Chinese,})
	tag, _ := language.MatchStrings(matcher, strings.Join(lang,","))
	I18n = message.NewPrinter(tag)
}

func init() {
	message.SetString(language.Chinese, "Wait", "等待")
	message.SetString(language.Chinese, "Run", "运行")
	message.SetString(language.Chinese, "Suspend", "暂停")
	message.SetString(language.Chinese, "Done", "完成")
	message.SetString(language.Chinese, "Failed", "失败")
	message.SetString(language.Chinese, "Upload", "上传")
	message.SetString(language.Chinese, "Download", "下载")
	message.SetString(language.Chinese, "Task from [%v] to [%v] suspended", "从[%v]到[%v]的任务已暂停")
	message.SetString(language.Chinese, "Task from [%v] to [%v] resumed", "从[%v]到[%v]的任务已恢复")
	message.SetString(language.Chinese, "File from [%v] to [%v] start", "开始从[%v]到[%v]的文件传输")
	message.SetString(language.Chinese, "Done:%s|%s| Total %v %v/s", "完成:%s|%s| 汇总 %v %v/s")
	message.SetString(language.Chinese, "File from [%v] to [%v] finished,%v", "从[%v]到[%v]的文件传输完成,%v")
	message.SetString(language.Chinese, "Failed:%s|%s| Total %v %v/s", "失败:%s|%s| 汇总 %v %v/s")
	message.SetString(language.Chinese, "File from [%v] to [%v] failed,%v", "从[%v]到[%v]的文件传输失败,%v")
	message.SetString(language.Chinese, "Processing:%s|%s| Total %v %v/s", "正处理:%s|%s| 汇总 %v %v/s")
	message.SetString(language.Chinese, "API response error: %v", "接口调用报错: %v")
	message.SetString(language.Chinese, "Task from [%v] to [%v] start", "开始从[%v]到[%v]的传输任务")
	message.SetString(language.Chinese, "Total %v, %v/s", "汇总 %v, %v/s")
	message.SetString(language.Chinese, "Task from [%v] to [%v] failed with error: %v", "从[%v]到[%v]的任务传输失败,报错： %v")
	message.SetString(language.Chinese, "Task from [%v] to [%v] done", "从[%v]到[%v]的任务传输完成")
	message.SetString(language.Chinese, "Try connecting %v", "连接 %v")
	message.SetString(language.Chinese, "Client %v init failed with error: %v", "客户端%v初始化失败,报错: %v")
	message.SetString(language.Chinese, "Read cfg.yaml failed: %v", "读取cfg.yaml报错: %v")
	message.SetString(language.Chinese, "Configuration File Error", "配置文件错误")
	message.SetString(language.Chinese, "Configuration File cfg.yaml incorrect, for instruction visit github.com/wct-devops/cos-transmit", "配置文件错误, 使用说明可以咨询DevOps团队或者查看github.com/wct-devops/cos-transmit")
	message.SetString(language.Chinese, "Parse cfg.yaml file failed: %v, for instruction visit github.com/wct-devops/cos-transmit", "解析cfg.yaml配置文件失败，使用说明可以咨询DevOps团队或者查看github.com/wct-devops/cos-transmit")
	message.SetString(language.Chinese, "Cloud Object Storage Transmit-WhaleCloud DevOps Team", "COS对象存储传输工具-浩鲸DevOps团队")
	message.SetString(language.Chinese, "Name", "文件名")
	message.SetString(language.Chinese, "Size", "大小")
	message.SetString(language.Chinese, "Modified", "修改时间")
	message.SetString(language.Chinese, "Refresh", "刷新目录")
	message.SetString(language.Chinese, "Delete", "删除文件(夹)")
	message.SetString(language.Chinese, "Explorer", "文件浏览器")
	message.SetString(language.Chinese, "From", "从")
	message.SetString(language.Chinese, "To", "到")
	message.SetString(language.Chinese, "Status", "状态")
	message.SetString(language.Chinese, "TaskType", "类型")
	message.SetString(language.Chinese, "Progress", "进度")
	message.SetString(language.Chinese, "Suspend", "暂停")
	message.SetString(language.Chinese, "Resume", "恢复")
	message.SetString(language.Chinese, "Delete", "删除")
	message.SetString(language.Chinese, "Get children list failed: %v", "获取子列表失败: %v")
	message.SetString(language.Chinese, "Error", "错误")
	message.SetString(language.Chinese, "Could not find matched endpoint for %v", "%v无法找到匹配的存储")
	message.SetString(language.Chinese, "Path [%v] does not exits in tree", "路径[%v]无法在目录树中找到")
	message.SetString(language.Chinese, "Please select at least one object to download", "请至少选择一个对象来下载")
	message.SetString(language.Chinese, "Please select the local path to save object(s)", "请指定保存对象的本地地址")
	message.SetString(language.Chinese, "Please select at least one object to upload", "请至少选择一个对象来上传")
	message.SetString(language.Chinese, "Please select the remote path to save object(s)", "请指定保存对象的远程地址")
	message.SetString(language.Chinese, "Please select at least one object to delete", "请至少选择一个对象来删除")
	message.SetString(language.Chinese, "Are you sure to delete directory: %v ?", "是否确认删除目录 %v ?")
	message.SetString(language.Chinese, "Are you sure to delete file: %v ?", "是否确认删除文件 %v ?")
	message.SetString(language.Chinese, "Delete confirm", "删除确认")
	message.SetString(language.Chinese, "Delete %v failed with error: %v", "删除 %v 失败, 报错%v")
	message.SetString(language.Chinese, "Delete %v success", "删除 %v 成功")
	message.SetString(language.Chinese, "Please select at least one task to suspend", "请至少选择一个任务来暂停")
	message.SetString(language.Chinese, "Please select at least one task to resume", "请至少选择一个任务来继续")
	message.SetString(language.Chinese, "Please select at least one task to delete", "请至少选择一个任务来删除")
	message.SetString(language.Chinese, "Local Path: ", "本地路径: ")
	message.SetString(language.Chinese, "Remote Path: ", "远程路径: ")
	message.SetString(language.Chinese, "Make Folder %v", "创建目录 %v")
	message.SetString(language.Chinese, "Error: %v", "错误: %v")
	message.SetString(language.Chinese, "Please input the folder name", "请输入文件夹名称")
	message.SetString(language.Chinese, "Folder name:", "文件夹名:")
	message.SetString(language.Chinese, "OK", "确定")
	message.SetString(language.Chinese, "Cancel", "取消")
	message.SetString(language.Chinese, "New Folder", "新建文件夹")
	message.SetString(language.Chinese, "Please select at least one object", "请至少选择一个对象")
	message.SetString(language.Chinese, "VersionId", "版本号")
	message.SetString(language.Chinese, "Copy URL", "复制URL")
	message.SetString(language.Chinese, "Download Share", "下载分享URL")
	message.SetString(language.Chinese, "Upload Share", "上传分享URL")
	message.SetString(language.Chinese, "Share URLs cannot be generated with anonymous credentials", "此存储使用匿名访问，无法生成分享URL")
}
