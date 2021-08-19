package main
import (
	"github.com/lxn/walk"
	"strings"
	"os"
	mc "github.com/minio/mc/cmd"
	"context"
	. "github.com/wct-devops/cos-transmit/common"
	"path/filepath"
	"os/exec"
	"fmt"
)

type MyMainWindow struct {
	mainWindow  *walk.MainWindow
	LogTextEdit *walk.TextEdit

	LocalTreeView   *walk.TreeView
	LocalTreeModel  *DirectoryTreeModel
	LocalTableView  *walk.TableView
	LocalTableModel *FileInfoModel
	LocalPathEdit   *walk.LineEdit
	LocalPathname   string

	RemoteTreeView   *walk.TreeView
	RemoteTreeModel  *DirectoryTreeModel
	RemoteTableView  *walk.TableView
	RemoteTableModel *FileInfoModel
	RemotePathEdit   *walk.LineEdit
	RemotePathname   string

	TaskTableView  *walk.TableView
	TaskTableModel *TaskInfoModel
	TaskMenuStop   *walk.Action
	TaskMenuStart  *walk.Action
	TaskMenuRemove *walk.Action
}

func NewMyMainWindow(conf *YamlCfg) *MyMainWindow {
	mw := &MyMainWindow{
		LocalTreeModel:   NewDirectoryTreeModel(),
		LocalTableModel:  NewFileInfoModel(),
		RemoteTreeModel:  NewDirectoryTreeModel(),
		RemoteTableModel: NewFileInfoModel(),
	}
	mw.TaskTableModel = NewTaskInfoModel(mw)
	return mw
}

func (mw *MyMainWindow) LocalTreeOnCurrentItemChanged() {
	dir := mw.LocalTreeView.CurrentItem().(*Directory)
	if err := mw.LocalTableModel.SetDirPath(dir.Path(), nil); err != nil {
		walk.MsgBox(
			mw.mainWindow,
			"Error",
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
	}
	mw.LocalPathEdit.SetText(dir.Path())
}

func (mw *MyMainWindow) LocalTableOnItemActivated() {
	if index := mw.LocalTableView.CurrentIndex(); index > -1 {
		name := mw.LocalTableModel.items[index].Name
		dir := mw.LocalTreeView.CurrentItem().(*Directory)
		if name == ".." {
			if dir.parent != nil {
				mw.LocalTreeView.SetCurrentItem(dir.parent)
			}
		} else {
			mw.LocalTreeView.Expanded(dir)
			for _, d := range dir.children {
				if d.name == name {
					mw.LocalTreeView.SetCurrentItem(d)
					mw.LocalPathEdit.SetText(d.Path())
				}
			}
		}
	}
}

func (mw *MyMainWindow) LocalPathEditChanged() {
	pathname := strings.TrimSuffix( mw.LocalPathEdit.Text(), string(os.PathSeparator))
	if mw.LocalPathname == pathname {
		return
	}
	mw.LocalPathname = pathname
	mw.LocalMenuRefreshAction()

	stat, err := os.Stat(mw.LocalPathname)
	if  err != nil &&  os.IsNotExist(err) {
		walk.MsgBox(
			mw.mainWindow,
			"Error",
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	pathList := strings.Split(mw.LocalPathname, string(os.PathSeparator))
	if stat.IsDir() == false {
		pathList = pathList[:len(pathList)-1]
	}

	var last *Directory
	children := mw.LocalTreeModel.roots
	for _, i := range pathList {
		ok := false
		for _, c := range children {
			if strings.TrimSuffix(c.name, string(os.PathSeparator)) == i {
				mw.LocalTreeView.Expanded(c)
				c.ResetChildren()
				last = c
				children = c.children
				ok = true
				break;
			}
		}
		if ok == false {
			walk.MsgBox(
				mw.mainWindow,
				"Error",
				"Path not exits" + i,
				walk.MsgBoxOK|walk.MsgBoxIconError)
			return
		}
	}

	mw.LocalTreeView.SetCurrentItem(last)
}

func (mw *MyMainWindow) RemoteTreeOnCurrentItemChanged() {
	dir := mw.RemoteTreeView.CurrentItem().(*Directory)
	if dir.parent == nil && dir.connected == false {
		dir.connected = true
		LOGGER.Info(I18n.Sprintf("Try connecting %v", dir.s3Config.HostURL))
		dir.ResetChildren()
	}
	if err := mw.RemoteTableModel.SetDirPath(dir.Path(), dir.s3Config); err != nil {
		walk.MsgBox(
			mw.mainWindow,
			"Error",
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
	}
	mw.RemotePathEdit.SetText(dir.Path())
}

func (mw *MyMainWindow) RemoteTableOnItemActivated() {
	if index := mw.RemoteTableView.CurrentIndex(); index > -1 {
		name := mw.RemoteTableModel.items[index].Name
		dir := mw.RemoteTreeView.CurrentItem().(*Directory)
		if name == ".." {
			mw.RemoteTreeView.SetCurrentItem(dir.parent)
		} else {
			mw.RemoteTreeView.Expanded(dir)
			for _, d := range dir.children {
				if d.name == name {
					mw.RemoteTreeView.SetCurrentItem(d)
					mw.RemoteTableModel.SetDirPath(d.Path(), d.s3Config)
				}
			}
		}
	}
}

func (mw *MyMainWindow) RemotePathEditChanged() {
	pathname := strings.TrimSuffix( mw.RemotePathEdit.Text(), "/")
	if mw.RemotePathname == pathname {
		return
	}
	mw.RemotePathname = pathname
	mw.RemoteMenuRefreshAction()
	var s3Config *mc.Config
	var content *mc.ClientContent
	var pathList []string
	for _, server := range CONF.Servers {
		if strings.HasPrefix(pathname, server.Endpoint) {
			s3Config = NewS3Config(&server)
			s3Client, err := mc.S3New(s3Config)
			if err != nil {
				LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", s3Config.HostURL, err))
			} else {
				content, err = s3Client.Stat(context.Background(), mc.StatOptions{})
				if err != nil {
					LOGGER.Error(I18n.Sprintf("API response error: %v", err))
				}
			}
			pathList = strings.Split(strings.TrimPrefix(strings.TrimPrefix(pathname, server.Endpoint), "/"), "/")
			pathList = append([]string{server.Endpoint}, pathList...)
		}
	}

	if content == nil {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("Could not find matched endpoint for %v", pathname),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	if content.Err != nil {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("API response error: %v", content.Err),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	var last *Directory
	children := mw.RemoteTreeModel.roots
	for _, i := range pathList {
		if len(i) < 1 {
			continue
		}
		ok := false
		for _, c := range children {
			if strings.TrimSuffix(c.name, "/") == i {
				mw.RemoteTreeView.Expanded(c)
				c.ResetChildren()
				last = c
				children = c.children
				ok = true
				break;
			}
		}
		if ok == false {
			walk.MsgBox(
				mw.mainWindow,
				I18n.Sprintf("Error"),
				I18n.Sprintf( "Path [%v] does not exits in tree", i),
				walk.MsgBoxOK|walk.MsgBoxIconError)
			return
		}
	}

	mw.RemoteTreeView.SetCurrentItem(last)
}

func (mw *MyMainWindow) RemoteMenuAddTaskAction() {
	if len(mw.RemoteTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf( "Please select at least one object to download"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	for _, i := range mw.RemoteTableView.SelectedIndexes() {
		remoteFile := mw.RemoteTableModel.items[i]
		if mw.LocalTreeView.CurrentItem() == nil {
			walk.MsgBox(
				mw.mainWindow,
				I18n.Sprintf("Error"),
				I18n.Sprintf( "Please select the local path to save object(s)"),
				walk.MsgBoxOK|walk.MsgBoxIconError)
			return
		}
		LocalDir := mw.LocalTreeView.CurrentItem().(*Directory)

		from := mw.RemoteTreeView.CurrentItem().(*Directory).Path() +  remoteFile.Name
		to := LocalDir.Path()
		task := NewTaskInfo(from, to, false, remoteFile.IsDir, remoteFile.s3Config)
		mw.TaskTableModel.AddTask(task)
	}
	mw.TaskTableModel.PublishRowsReset()
}

func (mw *MyMainWindow) LocalMenuAddTaskAction() {
	if len(mw.LocalTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf( "Please select at least one object to upload"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	for _, i := range mw.LocalTableView.SelectedIndexes() {
		localFile := mw.LocalTableModel.items[i]

		if mw.RemoteTreeView.CurrentItem() == nil {
			walk.MsgBox(
				mw.mainWindow,
				I18n.Sprintf("Error"),
				I18n.Sprintf( "Please select the remote path to save object(s)"),
				walk.MsgBoxOK|walk.MsgBoxIconError)
			return
		}
		RemoteDir := mw.RemoteTreeView.CurrentItem().(*Directory)
		LocalDir := mw.LocalTreeView.CurrentItem().(*Directory)

		from := filepath.Join( LocalDir.Path(), localFile.Name)
		to := RemoteDir.s3Config.HostURL
		task := NewTaskInfo(from, to, true, localFile.IsDir, RemoteDir.s3Config)
		mw.TaskTableModel.AddTask(task)
	}
	mw.TaskTableModel.PublishRowsReset()
}

func (mw *MyMainWindow) LocalMenuRefreshAction() {
	current := mw.LocalTreeView.CurrentItem()
	if current == nil {
		return
	}
	dir := current.(*Directory)
	dir.children = nil
	dir.ResetChildren()
	mw.LocalTreeModel.PublishItemsReset(dir)
	mw.LocalTreeView.Expanded(dir)
	mw.LocalTreeView.ExpandedChanged()
	mw.LocalTableModel.SetDirPath(dir.Path(), dir.s3Config)
}

func (mw *MyMainWindow) LocalMenuDelAction() {
	if len(mw.LocalTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("Please select at least one object to delete"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	dir := mw.LocalTreeView.CurrentItem().(*Directory)
	for _, i := range mw.LocalTableView.SelectedIndexes() {
		file := mw.LocalTableModel.items[i]
		filepath := filepath.Join( dir.Path(), file.Name)
		var text string
		if file.IsDir {
			text = I18n.Sprintf("Are you sure to delete directory: %v ?", filepath)
		} else {
			text = I18n.Sprintf("Are you sure to delete file: %v ?", filepath)
		}
		if 1 == walk.MsgBox(mw.mainWindow, I18n.Sprintf("Delete confirm"), text, walk.MsgBoxOKCancel) {
			err := os.RemoveAll(filepath)
			if err != nil {
				LOGGER.Error(I18n.Sprintf("Delete %v failed with error: %v", filepath, err))
			} else {
				LOGGER.Info(I18n.Sprintf("Delete %v success", filepath))
			}

		}
	}
	mw.LocalMenuRefreshAction()
}

func (mw *MyMainWindow) LocalMenuOpenExplorerAction() {
	current := mw.LocalTreeView.CurrentItem()
	if current == nil {
		return
	}
	dir := current.(*Directory)
	exec.Command("cmd", "/c", "start", "explorer", dir.Path()).Start()
}

func (mw *MyMainWindow) RemoteMenuRefreshAction() {
	current := mw.RemoteTreeView.CurrentItem()
	if current == nil {
		return
	}
	dir := current.(*Directory)
	dir.children = nil
	dir.ResetChildren()
	mw.RemoteTreeModel.PublishItemsReset(dir)
	mw.RemoteTreeView.Expanded(dir)
	mw.RemoteTableModel.SetDirPath(dir.Path(), dir.s3Config)
}

func (mw *MyMainWindow) RemoteMenuDelAction() {
	if len(mw.RemoteTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("Please select at least one object to delete"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	dir := mw.RemoteTreeView.CurrentItem().(*Directory)
	for _, i := range mw.RemoteTableView.SelectedIndexes() {
		file := mw.RemoteTableModel.items[i]
		filepath := dir.Path() + file.Name
		var text string
		if file.IsDir {
			// add a slash, otherwise it will delete all dirs with prefix 'filepath'
			filepath = filepath + "/"
			text = I18n.Sprintf("Are you sure to delete directory: %v ?", filepath)
		} else {
			text = I18n.Sprintf("Are you sure to delete file: %v ?", filepath)
		}
		if 1 == walk.MsgBox(mw.mainWindow, I18n.Sprintf("Delete confirm"), text, walk.MsgBoxOKCancel) {
			newS3Conf := S3ConfigClone(file.s3Config)
			newS3Conf.HostURL = filepath
			s3Client, err := mc.S3New(newS3Conf)
			if err != nil {
				LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", newS3Conf.HostURL, err))
			} else {
				var err error
				for perr := range s3Client.Remove(context.Background(), false, false, false, s3Client.List(context.Background(), mc.ListOptions{Recursive: true}) ) {
					if perr != nil {
						LOGGER.Error(I18n.Sprintf("Delete %v failed with error: %v", filepath, perr))
						err = perr.ToGoError()
					}
				}
				if err == nil {
					LOGGER.Info(I18n.Sprintf("Delete %v success", filepath))
				}
			}
		}
	}
	mw.RemoteMenuRefreshAction()
}

func (mw *MyMainWindow) TaskMenuSuspendTaskAction() {
	if len(mw.TaskTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("Please select at least one task to suspend"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	for _, i := range mw.TaskTableView.SelectedIndexes() {
		task := mw.TaskTableModel.items[i]
		task.Suspend()
		mw.TaskTableModel.PublishRowChanged(i)
	}
}

func (mw *MyMainWindow) TaskMenuResumeTaskAction() {
	if len(mw.TaskTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("Please select at least one task to resume"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	for _, i := range mw.TaskTableView.SelectedIndexes() {
		task := mw.TaskTableModel.items[i]
		task.Resume()
		mw.TaskTableModel.PublishRowChanged(i)
	}
}


func (mw *MyMainWindow) TaskMenuDelTaskAction() {
	if len(mw.TaskTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf("Please select at least one task to delete"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	for i, rowIdx := range mw.TaskTableView.SelectedIndexes() {
		newIdx := rowIdx - i
		task := mw.TaskTableModel.items[newIdx]
		task.Del()
		for _, next := range mw.TaskTableModel.items[newIdx+1:] {
			next.RowId = next.RowId - 1
		}
		mw.TaskTableModel.items = append(mw.TaskTableModel.items[:newIdx], mw.TaskTableModel.items[newIdx+1:]...)
		mw.TaskTableModel.PublishRowsRemoved(newIdx, newIdx)
	}
}

func (mw *MyMainWindow) LocalMenuNewFolderAction() {
	folderName := NewFolderDiag(mw.mainWindow)
	if len(folderName) < 1 || len(mw.LocalPathEdit.Text()) <1 {
		return
	}
	fullFolderName := filepath.Join( mw.LocalPathEdit.Text(), folderName)
	LOGGER.Info(I18n.Sprintf( "Make Folder %v", fullFolderName  ))
	err := os.MkdirAll(fullFolderName, os.ModePerm)
	if err != nil {
		LOGGER.Error(I18n.Sprintf("Error: %v", err))
	}
	mw.LocalMenuRefreshAction()
}

func (mw *MyMainWindow) RemoteMenuNewFolderAction() {
	folderName := NewFolderDiag(mw.mainWindow)
	if len(folderName) < 1 || len(mw.RemotePathEdit.Text()) <1 {
		return
	}
	fullFolderName :=  mw.RemotePathEdit.Text() + folderName
	LOGGER.Info(I18n.Sprintf( "Make Folder %v", fullFolderName ))

	dir := mw.RemoteTreeView.CurrentItem().(*Directory)

	newS3Conf := S3ConfigClone(dir.s3Config)
	// use an empty file to keep the path
	newS3Conf.HostURL =  fullFolderName + "/.coskeep"
	s3Client, err := mc.S3New(newS3Conf)
	if err != nil {
		LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", newS3Conf.HostURL, err))
	} else {
		s3Client.GetURL()
		_, perr := s3Client.Put(context.Background(), nil, 0, nil, mc.PutOptions{})
		if perr != nil {
			LOGGER.Error(I18n.Sprintf("Error: %v", perr))
		}
	}
	mw.RemoteMenuRefreshAction()
}

func (mw *MyMainWindow) OnDropFiles(files []string) {
	RemoteDir := mw.RemoteTreeView.CurrentItem().(*Directory)
	for _,f := range files {
		stat, _ := os.Stat(f)
		to := RemoteDir.s3Config.HostURL
		task := NewTaskInfo(f, to, true, stat.IsDir(), RemoteDir.s3Config)
		mw.TaskTableModel.AddTask(task)
	}
	mw.TaskTableModel.PublishRowsReset()
}

func (mw *MyMainWindow) RemoteMenuCopyURLAction() {
	if len(mw.RemoteTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf( "Please select at least one object"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}
	var urls []string
	for _, i := range mw.RemoteTableView.SelectedIndexes() {
		remoteFile := mw.RemoteTableModel.items[i]
		url := mw.RemoteTreeView.CurrentItem().(*Directory).Path() +  remoteFile.Name
		urls = append(urls, url)
	}
	walk.Clipboard().SetText( strings.Join(urls, "\r\n"))
}


func (mw *MyMainWindow) RemoteMenuDownloadShareURLAction() {
	if len(mw.RemoteTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf( "Please select at least one object"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}
	var urls []string
	for _, i := range mw.RemoteTableView.SelectedIndexes() {
		remoteFile := mw.RemoteTableModel.items[i]
		// now we only support URLs for file
		if remoteFile.IsDir {
			continue
		}
		newS3Conf := S3ConfigClone(remoteFile.s3Config)
		newS3Conf.HostURL = mw.RemoteTreeView.CurrentItem().(*Directory).Path() +  remoteFile.Name
		s3Client, err := mc.S3New(newS3Conf)
		if err != nil {
			LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", newS3Conf.HostURL, err))
		} else {
			url, perr := s3Client.ShareDownload(context.Background(), remoteFile.VersionId, EXPIRY)
			if perr != nil {
				if strings.Contains(fmt.Sprintf("%v", perr), "Presigned URLs cannot be generated with anonymous credentials" ){
					LOGGER.Error(I18n.Sprintf("Share URLs cannot be generated with anonymous credentials"))
				} else {
					LOGGER.Error(I18n.Sprintf("Error: %v", perr))
				}
			} else {
				urls = append(urls, url)
			}
		}
	}
	walk.Clipboard().SetText( strings.Join(urls, "\r\n"))
}

func (mw *MyMainWindow) RemoteMenuUploadShareURLAction() {
	if len(mw.RemoteTableView.SelectedIndexes()) < 1 {
		walk.MsgBox(
			mw.mainWindow,
			I18n.Sprintf("Error"),
			I18n.Sprintf( "Please select at least one object"),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}
	var urls []string
	for _, i := range mw.RemoteTableView.SelectedIndexes() {
		remoteFile := mw.RemoteTableModel.items[i]
		newS3Conf := S3ConfigClone(remoteFile.s3Config)
		newS3Conf.HostURL = mw.RemoteTreeView.CurrentItem().(*Directory).Path() +  remoteFile.Name
		s3Client, err := mc.S3New(newS3Conf)
		if err != nil {
			LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", newS3Conf.HostURL, err))
		} else {
			shareURL, uploadInfo, perr := s3Client.ShareUpload(context.Background(), false, EXPIRY, "")
			if perr != nil {
				if strings.Contains(fmt.Sprintf("%v", perr), "Presigned operations are not supported for anonymous credentials" ){
					LOGGER.Error(I18n.Sprintf("Share URLs cannot be generated with anonymous credentials"))
				} else {
					LOGGER.Error(I18n.Sprintf("Error: %v", perr))
				}
			} else {
				objectURL := s3Client.GetURL().String()
				curlCmd := makeCurlCmd(objectURL, shareURL, false, uploadInfo)
				urls = append(urls, curlCmd)
			}
		}
	}
	walk.Clipboard().SetText( strings.Join(urls, "\r\n"))
}

func makeCurlCmd(key, postURL string, isRecursive bool, uploadInfo map[string]string) string {
	postURL += " "
	curlCommand := "curl " + postURL
	for k, v := range uploadInfo {
		if k == "key" {
			key = v
			continue
		}
		curlCommand += fmt.Sprintf("-F %s=%s ", k, v)
	}
	// If key starts with is enabled prefix it with the output.
	if isRecursive {
		curlCommand += fmt.Sprintf("-F key=%s<NAME> ", key) // Object name.
	} else {
		curlCommand += fmt.Sprintf("-F key=%s ", key) // Object name.
	}
	curlCommand += "-F file=@<FILE>" // File to upload.
	return curlCommand
}