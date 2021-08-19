module github.com/wct-devops/cos-transmit

go 1.16

require (
	github.com/cheggaaa/pb v1.0.29
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/frankban/quicktest v1.13.0 // indirect
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/lxn/walk v0.0.0-20210112085537-c389da54e794
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/minio/mc v0.0.0-20210716165100-bfafb966324c
	github.com/minio/minio-go/v7 v7.0.11-0.20210607181445-e162fdb8e584
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/text v0.3.6
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/h2non/filetype.v1 v1.0.5
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/minio/mc v0.0.0-20210716165100-bfafb966324c => github.com/wangyumu/mc v0.0.0-20210729053358-390a1c5a9845
