package main

import (
	. "github.com/wct-devops/cos-transmit/common"
	"github.com/lxn/walk"
	mc "github.com/minio/mc/cmd"
	"context"
	"time"
	"os"
	"github.com/cheggaaa/pb"
	"io"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/pkg/mimedb"
	"gopkg.in/h2non/filetype.v1"
	"strings"
	"path"
	"github.com/minio/minio-go/v7"
	"path/filepath"
	"fmt"
)

const (
	TaskWait    = "Wait"
	TaskRun     = "Run"
	TaskSuspend = "Suspend"
	TaskDone    = "Done"
	TaskFail    = "Failed"

	TaskUpload   = "Upload"
	TaskDownload = "Download"
)

type TaskInfo struct {
	FromPath string
	ToPath   string
	SSize    string
	ViewStatus string
	ViewTaskType string
	Progress string
	RowId    int

	table      *TaskInfoModel
	taskType   string
	startTime  time.Time
	totalBytes int64
	isDir      bool
	size       int64
	s3Config   *mc.Config
	speed      int64
	status     string
	statusCh   chan int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewTaskInfo(from string, to string, isUpload bool, isDir bool, s3Config *mc.Config) *TaskInfo {
	ctx, cancel := context.WithCancel(context.Background())

	taskType := TaskDownload
	if isUpload {
		taskType = TaskUpload
	}

	return &TaskInfo{
		FromPath:     from,
		ToPath:       to,
		taskType:     taskType,
		ViewTaskType: I18n.Sprintf(taskType),
		isDir:        isDir,
		s3Config:     s3Config,
		status:       TaskWait,
		ViewStatus:   I18n.Sprintf(TaskWait),
		statusCh:     make(chan int, 1),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (t *TaskInfo) Do() error {
	var err error
	t.startTime = time.Now()
	if t.taskType == TaskUpload {
		err = t.Upload()
	} else {
		err = t.Download()
	}

	// if cancelled, do not set status
	if t.ctx.Err() != nil {
		return err
	}

	if err != nil {
		t.SetStatus(TaskFail)
	} else {
		t.SetStatus(TaskDone)
	}
	t.ctx.Done()
	return err
}

func (t *TaskInfo) Suspend() {
	t.SetStatus(TaskSuspend)
	t.cancel()
	LOGGER.Info(I18n.Sprintf("Task from [%v] to [%v] suspended", t.FromPath, t.ToPath))
}

func (t *TaskInfo) Resume() {
	ctx, cancel := context.WithCancel(context.Background())
	t.SetStatus(TaskWait)
	t.ctx = ctx
	t.cancel = cancel
	LOGGER.Info(I18n.Sprintf("Task from [%v] to [%v] resumed", t.FromPath, t.ToPath))
}

func (t *TaskInfo) Del() {
	if t.GetStatus() == TaskRun {
		t.cancel()
	}
}

func (t *TaskInfo) GetStatus() string{
	t.statusCh <- 1
	defer func() {
		<-t.statusCh
	}()
	return t.status
}

func (t *TaskInfo) SetStatus(status string){
	t.statusCh <- 1
	defer func() {
		<-t.statusCh
	}()
	t.status = status
	t.ViewStatus = I18n.Sprintf(status)
}

func (t *TaskInfo) Upload() error {
	if !t.isDir {
		return t.UploadOne(t.FromPath, t.ToPath, nil)
	} else {
		pathName := t.FromPath[strings.LastIndex(t.FromPath, string(os.PathSeparator))+1:]
		err := filepath.Walk(t.FromPath, func(fullPath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			relativePath := strings.TrimPrefix(fullPath, t.FromPath)
			relativePath = strings.TrimPrefix(relativePath, string(os.PathSeparator))
			relativePath = strings.TrimSuffix(relativePath, info.Name())
			return t.UploadOne(fullPath, t.ToPath + pathName + "/" + filepath.ToSlash(relativePath), info )
		})
		return err
	}
	return nil
}

func (t *TaskInfo) UploadOne(src string, dst string, info os.FileInfo) error {
	var err error
	LOGGER.Info(I18n.Sprintf("File from [%v] to [%v] start", src, dst))

	if info == nil {
		info, err = os.Stat(src)
		if err != nil {
			return err
		}
	}
	dst = dst + info.Name()

	newS3Conf := S3ConfigClone(t.s3Config)
	newS3Conf.HostURL = dst
	s3Client, perr := mc.S3New(newS3Conf)
	if perr != nil {
		return perr.ToGoError()
	}

	remoteStat, perr := s3Client.Stat(t.ctx, mc.StatOptions{})
	if perr == nil {
		if remoteStat.Err == nil {
			if remoteStat.Size == info.Size() && CONF.Ignore {
				LOGGER.Info(I18n.Sprintf("Same file found, %v skipped", dst))
				return nil
			}
		}
	}

	progress := pb.New64(info.Size()).SetUnits(pb.U_BYTES).SetWidth(20)
	progress.ShowSpeed = true
	progress.ShowBar = false
	progress.NotPrint = true
	progress.Start()
	wait := make(chan int, 1)
	done := make(chan int, 1)
	fail := make(chan int, 1)

	go func() {
		for {
			select {
			case <- wait :
				progress.Finish()
				t.totalBytes = t.totalBytes + info.Size()
				t.Progress = I18n.Sprintf("Done:%s|%s| Total %v %v/s",info.Name(), progress.String(), FormatByteSize(t.totalBytes), FormatByteSize(int64(float64(t.totalBytes)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
				t.table.PublishRowChanged(t.RowId)
				LOGGER.Info(I18n.Sprintf("File from [%v] to [%v] finished,%v", src, dst, progress.String()))
				done <- 1
				return
			case <- fail:
				progress.Finish()
				t.Progress = I18n.Sprintf("Failed:%s|%s| Total %v %v/s", info.Name(), progress.String(), FormatByteSize(t.totalBytes), FormatByteSize(int64(float64(t.totalBytes)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
				t.table.PublishRowChanged(t.RowId)
				LOGGER.Error(I18n.Sprintf("File from [%v] to [%v] failed,%v", src, dst, progress.String()))
				done <- 1
				return
			default:
				totalNow := t.totalBytes + progress.Get()
				t.Progress = I18n.Sprintf("Processing:%s|%s| Total %v %v/s",info.Name(), progress.String(), FormatByteSize(totalNow), FormatByteSize(int64(float64(totalNow)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
				t.table.PublishRowChanged(t.RowId)
				time.Sleep( time.Millisecond * 200)
			}
		}
	}()

	file, err := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	defer file.Close()

	if err != nil {
		return err
	}

	metadata := make(map[string]string)
	contentType := mimedb.TypeByExtension(filepath.Ext(src))
	if contentType == "application/octet-stream" {
		contentType, perr = probeContentType(file)
		if perr != nil {
			return perr.ToGoError()
		}
	}
	metadata["Content-Type"] = contentType

	fmt.Println(metadata)

	_, perr = s3Client.Put(t.ctx, file, info.Size(), progress, mc.PutOptions{Metadata:metadata })

	/*
	if n != info.Size() {
		return errors.New(I18n.Sprintf("Task uploaded size mismatch with source: %v : %v", n, info.Size() ))
	}
	*/

	if perr != nil {
		fail <- 1
		<- done
		return perr.ToGoError()
	}
	wait <- 1
	<- done
	return nil
}

func (t *TaskInfo) Download() error {
	if !t.isDir {
		return t.DownloadOne(t.FromPath, t.ToPath )
	} else {
		newS3Conf := S3ConfigClone(t.s3Config)
		newS3Conf.HostURL = t.FromPath
		s3Client, perr := mc.S3New(newS3Conf)
		if perr != nil {
			return perr.ToGoError()
		}

		currentPath := t.FromPath[strings.LastIndex(t.FromPath, "/")+1:]
		for content := range s3Client.List(context.Background(), mc.ListOptions{ Recursive: true } ) {
			if content.Err != nil {
				LOGGER.Error(I18n.Sprintf("API response error: %v", content.Err))
				continue
			}

			if content.StorageClass == "GLACIER" {
				continue
			}

			if content.Type.IsDir(){
				// never here
			} else {
				relativeName := strings.TrimPrefix(content.URL.Path, s3Client.GetURL().Path)
				fileNameWithSlash := relativeName[strings.LastIndex(relativeName, "/") :]
				relativePath := strings.TrimSuffix(relativeName, fileNameWithSlash)
				localPath := filepath.Join(t.ToPath, currentPath, filepath.FromSlash(relativePath))
				os.MkdirAll(localPath, os.ModePerm)
				t.DownloadOne( t.FromPath + relativeName , filepath.Join(t.ToPath, currentPath, relativePath)   )
			}
		}
	}
	return nil
}

func (t *TaskInfo) DownloadOne(src string, dst string) error {
	LOGGER.Info(I18n.Sprintf("File from [%v] to [%v] start", src, dst))
	fileName := src[strings.LastIndex(src, "/")+1:]
	dst = path.Join(dst, fileName)

	newS3Conf := S3ConfigClone(t.s3Config)
	newS3Conf.HostURL = src
	s3Client, perr := mc.S3New(newS3Conf)
	if perr != nil {
		return perr.ToGoError()
	}

	reader, perr := s3Client.Get(t.ctx, mc.GetOptions{})
	if perr != nil {
		return perr.ToGoError()
	}
	defer reader.Close()
	object, err := reader.(*minio.Object).Stat()
	if err != nil {
		return err
	}

	info, err := os.Stat(dst)
	if err == nil {
		if info.Size() == object.Size && CONF.Ignore {
			LOGGER.Info(I18n.Sprintf("Same file found, %v skipped", dst))
			return nil
		}
	}

	progress := pb.New64(object.Size).SetUnits(pb.U_BYTES).SetWidth(20)
	progress.ShowSpeed = true
	progress.ShowBar = false
	progress.NotPrint = true
	progress.Start()
	pgReader := progress.NewProxyReader(reader)

	wait := make (chan int, 1)
	done := make (chan int, 1)
	fail := make (chan int, 1)

	go func() {
		for {
			select {
			case <-wait:
				progress.Finish()
				t.totalBytes = t.totalBytes + object.Size
				t.Progress = I18n.Sprintf("Done:%s|%s| Total %v %v/s", fileName, progress.String(), FormatByteSize(t.totalBytes), FormatByteSize(int64(float64(t.totalBytes)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
				t.table.PublishRowChanged(t.RowId)
				LOGGER.Info(I18n.Sprintf("File from [%v] to [%v] finished,%v", src, dst, progress.String()))
				done <- 1
				return
			case <-fail:
				progress.Finish()
				t.Progress = I18n.Sprintf("Failed:%s|%s| Total %v %v/s", fileName, progress.String(), FormatByteSize(t.totalBytes), FormatByteSize(int64(float64(t.totalBytes)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
				t.table.PublishRowChanged(t.RowId)
				LOGGER.Error(I18n.Sprintf("File from [%v] to [%v] failed,%v", src, dst, progress.String()))
				done <- 1
				return
			default:
				totalNow := t.totalBytes + progress.Get()
				t.Progress = I18n.Sprintf("Processing:%s|%s| Total %v %v/s", fileName, progress.String(), FormatByteSize(totalNow), FormatByteSize(int64(float64(totalNow)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
				t.table.PublishRowChanged(t.RowId)
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()

	file, err := os.Create(dst)
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(file, pgReader)
	if err != nil {
		fail <- 1
		<- done
		return err
	}
	wait <- 1
	<- done
	return nil
}

func (t *TaskInfo) Stop() error {
	t.cancel()
	t.SetStatus(TaskSuspend)
	return nil
}

type TaskInfoModel struct {
	walk.SortedReflectTableModelBase
	items   []*TaskInfo
}

func NewTaskInfoModel(mw *MyMainWindow) *TaskInfoModel {
	taskTable := new(TaskInfoModel)
	go func(){
		for {
			for _, t := range taskTable.items {
				if t.GetStatus() == TaskWait {
					LOGGER.Info(I18n.Sprintf("Task from [%v] to [%v] start", t.FromPath, t.ToPath))
					t.SetStatus(TaskRun)
					err := t.Do()
					t.Progress = I18n.Sprintf("Total %v, %v/s", FormatByteSize(t.totalBytes), FormatByteSize(int64(float64(t.totalBytes)/(float64(time.Now().Sub(t.startTime))/float64(time.Second)))))
					t.table.PublishRowChanged(t.RowId)
					if err != nil {
						LOGGER.Error(I18n.Sprintf("Task from [%v] to [%v] failed with error: %v", t.FromPath, t.ToPath, err))
					} else {
						LOGGER.Info(I18n.Sprintf("Task from [%v] to [%v] done", t.FromPath, t.ToPath))
					}
					if t.taskType == TaskDownload {
						mw.LocalMenuRefreshAction()
					} else {
						mw.RemoteMenuRefreshAction()
					}
				}
			}
			time.Sleep( time.Millisecond * 50)
		}
	}()

	return taskTable
}

func (m *TaskInfoModel) RefreshRow(row int) {
	m.TableModelBase.PublishRowChanged(row)
}

func (m *TaskInfoModel) AddTask(info *TaskInfo) {
	info.RowId = len(m.items)
	info.table = m
	m.items = append(m.items, info)
}

func (m *TaskInfoModel) Items() interface{} {
	return m.items
}

func (m *TaskInfoModel) Image(row int) interface{} {
	f := m.items[row]
	if f.isDir {
		return "."
	} else {
		return "cfg.yaml"
	}
}

func probeContentType(reader io.Reader) (ctype string, err *probe.Error) {
	ctype = "application/octet-stream"
	// Read a chunk to decide between utf-8 text and binary
	if s, ok := reader.(io.Seeker); ok {
		var buf [512]byte
		n, _ := io.ReadFull(reader, buf[:])
		if n <= 0 {
			return ctype, nil
		}

		kind, e := filetype.Match(buf[:n])
		if e != nil {
			return ctype, probe.NewError(e)
		}
		// rewind to output whole file
		if _, e = s.Seek(0, io.SeekStart); e != nil {
			return ctype, probe.NewError(e)
		}

		if kind.MIME.Value != "" {
			ctype = kind.MIME.Value
		}
	}
	return ctype, nil
}