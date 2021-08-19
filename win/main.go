package main

import (
	. "github.com/wct-devops/cos-transmit/common"
	"github.com/lxn/walk"
	"io/ioutil"
	"fmt"
	"time"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	log "github.com/cihub/seelog"
	. "github.com/lxn/walk/declarative"
	"encoding/base64"
	"bytes"
	"image"
	"image/png"
)

var (
	CONF     *YamlCfg
	LOGGER   CtxLogger
	HOME     = "data"
	MAIN_ICO = string("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")
	EXPIRY = time.Duration(604800) * time.Second
)

type Server struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accesskey"`
	SecretKey string `yaml:"secretkey"`
}

type YamlCfg struct {
	Servers []Server `yaml:"cos,omitempty"`
	Lang    string   `yaml:"lang,omitempty"`
	Ignore bool `yaml:"ignore"`
	Trace bool `yaml:"trace"`
	Expiry int64 `yaml:"expiry"`
}

func main() {
	InitI18nPrinter("")
	var loggerCfg []byte
	if _, err := os.Stat( "logCfg.xml"); err == nil {
		loggerCfg, _ = ioutil.ReadFile("logCfg.xml")
	} else if _, err := os.Stat( filepath.Join(HOME, "logCfg.xml")); err == nil {
		loggerCfg, _ = ioutil.ReadFile(filepath.Join(HOME, "logCfg.xml"))
	}
	InitLogger(loggerCfg)

	CONF = &YamlCfg{
		Ignore : true,
		Trace: false,
	}
	var cfgFile []byte
	_, err := os.Stat( "cfg.yaml")
	if err != nil && os.IsNotExist(err) {
		_, err = os.Stat( filepath.Join(HOME,"cfg.yaml"))
		if err != nil && os.IsNotExist(err) {
			log.Error(I18n.Sprintf("Read cfg.yaml failed: %v", err))
		} else {
			cfgFile, err = ioutil.ReadFile(filepath.Join(HOME,"cfg.yaml"))
			if err != nil {
				log.Error(I18n.Sprintf("Read cfg.yaml failed: %v", err))
			}
		}
	} else {
		cfgFile, err = ioutil.ReadFile("cfg.yaml")
		if err != nil {
			log.Error(I18n.Sprintf("Read cfg.yaml failed: %v", err))
		}
	}

	err = yaml.Unmarshal(cfgFile, CONF)
	mw := NewMyMainWindow(CONF)
	if len(CONF.Lang) >1 {
		InitI18nPrinter(CONF.Lang)
	}

	if err != nil {
		walk.MsgBox(nil,
			I18n.Sprintf("Configuration File Error"),
			fmt.Sprintf(I18n.Sprintf("Parse cfg.yaml file failed: %v, for instruction visit github.com/wct-devops/cos-transmit", err)),
			walk.MsgBoxIconStop)
		return
	}

	if len(CONF.Servers) < 1 {
		walk.MsgBox(nil,
			I18n.Sprintf("Configuration File Error"),
			I18n.Sprintf("Configuration File cfg.yaml incorrect, for instruction visit github.com/wct-devops/cos-transmit"),
			walk.MsgBoxIconStop)
		return
	}

	if CONF.Expiry > 0 {
		EXPIRY = time.Duration(CONF.Expiry) * time.Second
	}

	icon, _ := walk.NewIconFromImageForDPI(Base64ToImage(MAIN_ICO), 96)

	err = MainWindow {
		AssignTo: &mw.mainWindow,
		Icon:     icon,
		Title:    I18n.Sprintf("Cloud Object Storage Transmit-WhaleCloud DevOps Team"),
		MinSize:  Size{600, 400},
		Layout:   VBox{},
		OnDropFiles: mw.OnDropFiles,
		Children: []Widget{
			VSplitter{
				Children: []Widget{
					Composite{
						Layout:  VBox{MarginsZero: true},
						MaxSize: Size{0, 150},
						MinSize: Size{0, 50},
						Alignment: AlignHNearVNear,
						Children: []Widget{
							//Label{Text: I18n.Sprintf("Log Output: ") , Font: Font{ Bold: true}},
							TextEdit{
								AssignTo: &mw.LogTextEdit,
								MaxLength: 1000000,
								Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
								ReadOnly: true,
								VScroll: true,
							},
						},
					},
					HSplitter{
						MinSize: Size{0 , 250},
						Children: []Widget{
							VSplitter{
								Children: []Widget {
									Composite{
										Layout:  VBox{ MarginsZero: true },
										Children: []Widget{
											Composite{
												Background: SolidColorBrush{Color: walk.RGB(230, 230, 230)},
												Layout:  HBox{ MarginsZero: true },
												Children: []Widget{
													Label{ Text: I18n.Sprintf("Local Path: "), Font: Font{Bold: true} },
													LineEdit{ AssignTo: &mw.LocalPathEdit, OnEditingFinished: mw.LocalPathEditChanged },
												},
											},
											TreeView{
												AssignTo: &mw.LocalTreeView,
												OnCurrentItemChanged: mw.LocalTreeOnCurrentItemChanged,
											},
										},
									},
									TableView{
										StretchFactor: 2,
										Columns: []TableViewColumn{
											TableViewColumn{
												DataMember: "Name",
												Title: I18n.Sprintf("Name"),
												Width:      396,
											},
											TableViewColumn{
												DataMember: "SSize",
												Title: I18n.Sprintf("Size"),
												Width:      64,
											},
											TableViewColumn{
												DataMember: "Modified",
												Title: I18n.Sprintf("Modified"),
												Format:     "2006-01-02 15:04:05",
												Width:      120,
											},
										},
										AssignTo: &mw.LocalTableView,
										Model: mw.LocalTableModel,
										MultiSelection: true,
										OnItemActivated: mw.LocalTableOnItemActivated,
										ContextMenuItems:  []MenuItem{
											Action{Text: I18n.Sprintf("Upload"), OnTriggered: mw.LocalMenuAddTaskAction},
											Action{Text: I18n.Sprintf("New Folder"), OnTriggered: mw.LocalMenuNewFolderAction},
											Action{Text: I18n.Sprintf("Refresh"), OnTriggered: mw.LocalMenuRefreshAction},
											Action{Text: I18n.Sprintf("Delete"), OnTriggered: mw.LocalMenuDelAction},
											Action{Text: I18n.Sprintf("Explorer"), OnTriggered: mw.LocalMenuOpenExplorerAction},
										},
									},
								},
							},
							VSplitter{
								Children: []Widget {
									Composite{
										Layout:  VBox{ MarginsZero: true },
										Children: []Widget{
											Composite{
												Background: SolidColorBrush{Color: walk.RGB(230, 230, 230)},
												Layout:  HBox{ MarginsZero: true },
												Children: []Widget{
													Label{Text: I18n.Sprintf("Remote Path: "), Font: Font{Bold: true}},
													LineEdit{  AssignTo: &mw.RemotePathEdit, OnEditingFinished: mw.RemotePathEditChanged },
												},
											},
											TreeView{
												AssignTo: &mw.RemoteTreeView,
												OnCurrentItemChanged: mw.RemoteTreeOnCurrentItemChanged,
											},
										},
									},
									TableView{
										StretchFactor: 2,
										Columns: []TableViewColumn{
											TableViewColumn{
												DataMember: "Name",
												Title:      I18n.Sprintf("Name"),
												Width:      320,
											},
											TableViewColumn{
												DataMember: "SSize",
												Title:      I18n.Sprintf("Size"),
												Width:      64,
											},
											TableViewColumn{
												DataMember: "Modified",
												Title:      I18n.Sprintf("Modified"),
												Format:     "2006-01-02 15:04:05",
												Width:      120,
											},
											TableViewColumn{
												DataMember: "VersionId",
												Title:      I18n.Sprintf("VersionId"),
												Width:      64,
											},
										},
										AssignTo: &mw.RemoteTableView,
										Model: mw.RemoteTableModel,
										MultiSelection: true,
										OnItemActivated: mw.RemoteTableOnItemActivated,
										ContextMenuItems: []MenuItem{
											Action{Text: I18n.Sprintf("Download"), OnTriggered: mw.RemoteMenuAddTaskAction},
											Action{Text: I18n.Sprintf("New Folder"), OnTriggered: mw.RemoteMenuNewFolderAction},
											Action{Text: I18n.Sprintf("Refresh"), OnTriggered: mw.RemoteMenuRefreshAction},
											Action{Text: I18n.Sprintf("Delete"), OnTriggered: mw.RemoteMenuDelAction},
											Action{Text: I18n.Sprintf("Copy URL"), OnTriggered: mw.RemoteMenuCopyURLAction},
											Action{Text: I18n.Sprintf("Download Share"), OnTriggered: mw.RemoteMenuDownloadShareURLAction},
											Action{Text: I18n.Sprintf("Upload Share"), OnTriggered: mw.RemoteMenuUploadShareURLAction},
										},
									},
								},
							},
						},
					},
					TableView{
						MinSize: Size{0, 100},
						AssignTo: &mw.TaskTableView,
						Model: mw.TaskTableModel,
						MultiSelection:   true,
						Columns: []TableViewColumn{
							TableViewColumn{
								DataMember: "FromPath",
								Title: I18n.Sprintf("From"),
								Width:      256,
							},
							TableViewColumn{
								DataMember: "ToPath",
								Title: I18n.Sprintf("To"),
								Width:      256,
							},
							TableViewColumn{
								DataMember: "ViewStatus",
								Title: I18n.Sprintf("Status"),
								Width:      80,
							},
							TableViewColumn{
								DataMember: "ViewTaskType",
								Title: I18n.Sprintf("TaskType"),
								Width:      80,
							},
							TableViewColumn{
								DataMember: "Progress",
								Title: I18n.Sprintf("Progress"),
								Width:      512,
							},
						},
						ContextMenuItems: []MenuItem{
							Action{Text: I18n.Sprintf("Suspend"), OnTriggered: mw.TaskMenuSuspendTaskAction },
							Action{Text:  I18n.Sprintf("Resume"), OnTriggered: mw.TaskMenuResumeTaskAction },
							Action{Text:  I18n.Sprintf("Delete"), OnTriggered: mw.TaskMenuDelTaskAction },
						},
					},
				},
			},
		},
	}.Create()

	if err != nil {
		panic(err)
	}

	/*
	os.Stderr = os.NewFile(uintptr(syscall.Stderr), "data/stderr.txt")
	os.Stdout = os.NewFile(uintptr(syscall.Stdout), "data/stdout.txt")

	go func(){
		mw.LogTextEdit.AppendText(time.Now().Format("[2006-01-02 15:04:05]") + " Start \r\n" )
		buf := make([]byte, 1024)
		r, err := os.OpenFile("data/console.txt", os.O_RDONLY, 0666)
		if err != nil {
			panic(err)
		}
		for {
			f.Sync()
			n, err := r.Read(buf)
			fmt.Println(n)
			if err != nil {
				panic(err)
			}
			if n > 0 {
				mw.LogTextEdit.AppendText( time.Now().Format("[2006-01-02 15:04:05]") + " " + string(buf) + "\r\n" )
			}
			buf = make([]byte, 1024)
			time.Sleep(time.Millisecond * 50)
		}
	}()
	*/

	LOGGER = newGuiLogger(mw.LogTextEdit)
	mw.LocalTreeModel.Init(true)
	mw.LocalTreeView.SetModel(mw.LocalTreeModel)
	mw.RemoteTreeModel.Init( false)
	mw.RemoteTreeView.SetModel(mw.RemoteTreeModel)
	mw.mainWindow.Show()
	mw.mainWindow.Run()
}

func NewFolderDiag(mw *walk.MainWindow) string {
	var folderName string

	var dlg *walk.Dialog
	var le *walk.LineEdit
	var dialog = Dialog{}
	dialog.Title = I18n.Sprintf("Please input the folder name")
	dialog.MinSize = Size{350, 80}
	dialog.Layout = VBox{}
	dialog.AssignTo = &dlg

	childrens := []Widget{
		Composite{
			Layout: HBox{},
			Children: []Widget{
				HSpacer{},
				Label{
					Text: I18n.Sprintf("Folder name:"),
				},
				LineEdit{
					AssignTo: &le,
					MaxLength: 300,
				},
			},
		},
		Composite{
			Layout: HBox{},
			Children: []Widget{
				HSpacer{},
				PushButton{
					Text:     I18n.Sprintf("OK"),
					OnClicked: func() {
						folderName = le.Text()
						dlg.Accept()
					},
				},
				PushButton{
					Text:     I18n.Sprintf( "Cancel"),
					OnClicked: func() {
						dlg.Cancel() },
				},
			},
		},
	}
	dialog.Children = childrens
	dialog.Run(mw)

	return folderName
}

func Base64ToImage(str string) image.Image {
	ddd, _ := base64.StdEncoding.DecodeString(str)
	bbb := bytes.NewBuffer(ddd)
	m, _, _ := image.Decode(bbb)
	png.Encode(bbb, m)
	return m
}

type GuiLogger struct {
	te	*walk.TextEdit
	logChan    chan int
}

func newGuiLogger(te *walk.TextEdit) CtxLogger {
	return &GuiLogger{
		te: te ,
		logChan: make(chan int, 1),
	}
}

func (l *GuiLogger) Info(logStr string) {
	l.logChan <- 1
	defer func() {
		<-l.logChan
	}()
	l.te.AppendText( time.Now().Format("[2006-01-02 15:04:05]") + " " + logStr + "\r\n" )
	log.Info(logStr)
}

func (l *GuiLogger) Error(logStr string) {
	l.logChan <- 1
	defer func() {
		<-l.logChan
	}()
	l.te.AppendText( time.Now().Format("[2006-01-02 15:04:05]") + " " + logStr + "\r\n" )
	log.Error(logStr)
}

func (l *GuiLogger) Debug(logStr string) {
	log.Debug(fmt.Sprint(logStr))
}

func (l *GuiLogger) Errorf(format string, args ...interface{}) error {
	l.logChan <- 1
	defer func() {
		<-l.logChan
	}()
	errStr := fmt.Sprintf(format, args)
	l.te.AppendText( time.Now().Format("[2006-01-02 15:04:05]<ERROR>") + " " + fmt.Sprintf(format, args...) + "\r\n")
	log.Error(errStr)
	return fmt.Errorf(format, args...)
}