package main

import (
	. "github.com/wct-devops/cos-transmit/common"
	"path/filepath"
	"github.com/lxn/walk"
	"os"
	"time"
	"strings"
	mc "github.com/minio/mc/cmd"
	"context"
)

type DirectoryTreeModel struct {
	walk.TreeModelBase
	local bool
	roots []*Directory
}

func NewDirectoryTreeModel() *DirectoryTreeModel {
	model := new(DirectoryTreeModel)
	return model
}

func (m *DirectoryTreeModel) Init(local bool) {
	if local {
		drives, err := walk.DriveNames()
		if err != nil {
			panic(err)
		}
		for _, drive := range drives {
			switch drive {
			case "A:\\", "B:\\":
				continue
			}
			m.roots = append(m.roots, NewDirectory(drive, nil, nil))
		}
	} else {
		for _, server := range CONF.Servers {
			server.Endpoint = strings.TrimSuffix(server.Endpoint, "/")
			s3Config := NewS3Config(&server)
			s3Config.HostURL = s3Config.HostURL + "/"
			_, perr := mc.S3New(s3Config)
			if perr != nil {
				LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", s3Config.HostURL, perr))
				continue
			}
			m.roots = append(m.roots, NewDirectory(server.Endpoint, nil, s3Config))
		}
	}
}

func (*DirectoryTreeModel) LazyPopulation() bool {
	// We don't want to eagerly populate our tree view with the whole file system.
	return true
}

func (m *DirectoryTreeModel) RootCount() int {
	return len(m.roots)
}

func (m *DirectoryTreeModel) RootAt(index int) walk.TreeItem {
	return m.roots[index]
}

func NewS3Config(server *Server) *mc.Config {
	return &mc.Config{
		AccessKey: server.AccessKey,
		SecretKey: server.SecretKey,
		HostURL:   server.Endpoint,
		Signature: "S3v4",
	}
}

type Directory struct {
	name     string
	parent   *Directory
	children []*Directory
	s3Config *mc.Config
	connected bool
}

func NewDirectory(name string, parent *Directory, s3Config *mc.Config) *Directory {
	return &Directory{name: name, parent: parent, s3Config: s3Config}
}

var _ walk.TreeItem = new(Directory)

func (d *Directory) Text() string {
	return d.name
}

func (d *Directory) Parent() walk.TreeItem {
	if d.parent == nil {
		// We can't simply return d.parent in this case, because the interface
		// value then would not be nil.
		return nil
	}
	return d.parent
}

func (d *Directory) ChildCount() int {
	if d.children == nil {
		// It seems this is the first time our child count is checked, so we
		// use the opportunity to populate our direct children.
		if err := d.ResetChildren(); err != nil {
			LOGGER.Error(I18n.Sprintf("Get children list failed: %v", err))
		}
	}

	return len(d.children)
}

func (d *Directory) ChildAt(index int) walk.TreeItem {
	return d.children[index]
}

func (d *Directory) Image() interface{} {
	if d.s3Config == nil {
		return d.Path()
	} else {
		return "."
	}
}

func (d *Directory) ResetChildren() error {
	//open the tree after user activate it
	if d.parent == nil && d.connected == false && d.s3Config != nil {
		return nil
	}

	//TODo we add add refresh here
	if d.children != nil {
		return nil
	}

	d.children = make([]*Directory, 0)

	if d.s3Config != nil {
		return d.ResetChildrenS3()
	}

	dirPath := d.Path()

	if err := filepath.Walk(d.Path(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if info == nil {
				return filepath.SkipDir
			}
		}

		name := info.Name()

		if !info.IsDir() || path == dirPath {
			return nil
		}

		if shouldExclude(name) == false {
			d.children = append(d.children, NewDirectory(name, d, nil))
		}

		return filepath.SkipDir
	}); err != nil {
		return err
	}

	return nil
}

func S3ConfigClone(s3Config *mc.Config) *mc.Config {
	return &mc.Config{
		AccessKey:    s3Config.AccessKey,
		SecretKey:    s3Config.SecretKey,
		SessionToken: s3Config.SessionToken,
		Signature:    s3Config.Signature,
		HostURL:      s3Config.HostURL,
		AppName:      s3Config.AppName,
		AppVersion:   s3Config.AppVersion,
		Debug:        CONF.Trace,
		Insecure:     s3Config.Insecure,
		Lookup:       s3Config.Lookup,
		Transport:    s3Config.Transport,
	}
}

func (d *Directory) ResetChildrenS3() error {
	d.s3Config.HostURL = d.Path()
	s3Client, err := mc.S3New(d.s3Config)
	if err != nil {
		LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", d.s3Config.HostURL, err))
		return err.ToGoError()
	}

	for content := range s3Client.List(context.Background(), mc.ListOptions{}) {
		if content.Err != nil {
			LOGGER.Error(I18n.Sprintf("API response error: %v", content.Err))
			continue
		}

		if content.StorageClass == "GLACIER" {
			continue
		}

		if content.Type.IsDir(){
			fullPath := strings.TrimSuffix(content.URL.Path, "/")
			name := fullPath[strings.LastIndex(fullPath, "/")+1:]
			newS3Conf := S3ConfigClone(d.s3Config)
			d.children = append(d.children, NewDirectory(name, d, newS3Conf))
		}
	}

	return nil
}

func (d *Directory) Path() string {
	elems := []string{d.name}
	dir, _ := d.Parent().(*Directory)
	for dir != nil {
		elems = append([]string{dir.name}, elems...)
		dir, _ = dir.Parent().(*Directory)
	}
	if d.s3Config != nil {
		return strings.Join(elems, "/") + "/"
	} else {
		return filepath.Join(elems...)
	}
}

type FileInfo struct {
	Name     string
	Size     int64
	SSize    string
	Modified time.Time
	s3Config *mc.Config
	IsDir    bool
	VersionId string
}

type FileInfoModel struct {
	walk.SortedReflectTableModelBase
	dirPath string
	items   []*FileInfo
}

func NewFileInfoModel() *FileInfoModel {
	return new(FileInfoModel)
}

func (m *FileInfoModel) Items() interface{} {
	return m.items
}

func (m *FileInfoModel) SetDirPath(dirPath string, s3Config *mc.Config) error {
	m.dirPath = dirPath
	m.items = nil

	if s3Config != nil {
		return m.SetDirPathS3(dirPath, s3Config)
	}

	item := &FileInfo{
		Name: "..",
	}
	m.items = append(m.items, item)

	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if info == nil {
				return filepath.SkipDir
			}
		}
		name := info.Name()

		if path == dirPath {
			return nil
		}

		if !shouldExclude(name) {
			item := &FileInfo{
				Name:     name,
				Size:     info.Size(),
				SSize:    FormatByteSize(info.Size()),
				Modified: info.ModTime(),
				IsDir: info.IsDir(),
			}
			m.items = append(m.items, item)
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	}); err != nil {
		return err
	}

	m.PublishRowsReset()

	return nil
}

func (m *FileInfoModel) SetDirPathS3(dirPath string, s3Config *mc.Config) error {
	s3Config.HostURL = dirPath
	s3Client, err := mc.S3New(s3Config)
	if err != nil {
		LOGGER.Error(I18n.Sprintf("Client %v init failed with error: %v", s3Config.HostURL, err))
		return err.ToGoError()
	}
	for content := range s3Client.List(context.Background(), mc.ListOptions{}) {
		if content.Err != nil {
			LOGGER.Error(I18n.Sprintf("API response error: %v", content.Err))
			continue
		}

		if content.StorageClass == "GLACIER" {
			continue
		}

		fullPath := strings.TrimSuffix(content.URL.Path, "/")
		name := fullPath[strings.LastIndex(fullPath, "/")+1:]

		conf := S3ConfigClone(s3Config)
		item := &FileInfo{
			Name:     name,
			Size:     content.Size,
			SSize:    FormatByteSize(content.Size),
			Modified: content.Time,
			IsDir:    content.Type.IsDir(),
			s3Config: conf,
			VersionId: content.VersionID,
		}
		m.items = append(m.items, item)
	}

	m.PublishRowsReset()

	return nil
}

func (m *FileInfoModel) Image(row int) interface{} {
	f := m.items[row]
	if f.s3Config == nil {
		return filepath.Join(m.dirPath, f.Name)
	} else {
		if f.IsDir {
			return "."
		} else {
			return "cfg.yaml"
		}
	}
}

func shouldExclude(name string) bool {
	switch strings.ToUpper(name) {
	case "SYSTEM VOLUME INFORMATION", "PAGEFILE.SYS", "SWAPFILE.SYS", "$RECYCLE.BIN", ".DS_STORE", "LOST+FOUND":
		return true
	}
	return false
}

