package archiver

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/XiaoMiku01/bilibili-archiver/internal"
	"github.com/rs/zerolog/log"
)

// 用于更新已下载投稿的元数据和弹幕

type VideoMetaPath struct {
	Path string
	Meta internal.VideoMetaStruct
}

func (au *ArchiverUser) getAllMetaFiles() []string {
	// 获取所有下载的投稿的元数据路径 _meta.json
	var metaPaths []string
	filepath.Walk(au.config.SavePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// 只添加以 _meta.json 结尾的文件
		if strings.HasSuffix(path, "_meta.json") {
			metaPaths = append(metaPaths, path)
		}
		return nil
	})
	return metaPaths
}

func (au *ArchiverUser) UpdateVideoMeta() {
	for {
		time.Sleep(time.Duration(au.config.UpdateInterval) * time.Minute)
		au.bapi.InitGRPC()
		metaPaths := au.getAllMetaFiles()
		var vmetas []VideoMetaPath
		for _, metaPath := range metaPaths {
			var meta internal.VideoMetaStruct
			metaFile, err := os.Open(metaPath)
			if err != nil {
				log.Error().Err(err).Msgf("打开元数据文件失败: %s", metaPath)
				continue
			}
			err = json.NewDecoder(metaFile).Decode(&meta)
			if err != nil {
				log.Error().Err(err).Msgf("解析元数据文件失败: %s", metaPath)
				continue
			}
			// 当前时间前 n 天 用于判断是否需要更新投稿信息
			t := int(time.Now().AddDate(0, 0, -au.config.UpdateDL).Unix())
			if meta.Ctime > t {
				vmetas = append(vmetas, VideoMetaPath{
					Path: metaPath,
					Meta: meta,
				})

			}
		}

		// time.Sleep(time.Duration(au.config.UpdateInterval) * time.Minute)
		log.Info().Msgf("开始更新元数据, 共 %d 个投稿在更新范围", len(vmetas))
		for _, vmeta := range vmetas {
			// log.Debug().Msgf("更新元数据: %s", vmeta.Path)
			vinfo, err := au.bapi.GetView(&internal.ViewReq{
				Aid: vmeta.Meta.Aid,
			})
			if err != nil {
				log.Error().Err(err).Msgf("获取投稿信息失败: %s", vmeta.Path)
				continue
			}
			// 更新元数据
			jsonData, _ := json.MarshalIndent(vinfo.Arc, "", "  ")
			f, err := os.Create(vmeta.Path)
			if err != nil {
				log.Error().Err(err).Msgf("创建文件失败: %s", vmeta.Path)
				return
			}
			defer f.Close()
			f.WriteString(string(jsonData))
			log.Debug().Msgf("更新投稿元数据完成: %s", vinfo.Arc.Title)
			// 更新弹幕
			if au.config.Danmaku {
				au.updateDanmaku(vmeta.Path)
			}

			if au.config.RunAfterUpdate != "" {
				pdir := filepath.Dir(vmeta.Path)
				internal.ExecCommand(au.config.RunAfterUpdate, pdir)
			}
		}
		log.Info().Msg("元数据更新完成")
	}
}

func (au *ArchiverUser) updateDanmaku(vpath string) {
	// 更新弹幕
	pdir := filepath.Dir(vpath)
	// 获取 pdir 目录下所有 _danmaku.xml 文件
	var danmakuPaths []string
	filepath.Walk(pdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_danmaku.xml") {
			danmakuPaths = append(danmakuPaths, path)
		}
		return nil
	})

	for _, danmakuPath := range danmakuPaths {
		var originalDanmaku internal.DanmakuXmlstruct
		var originalNum int
		danmakuFile, err := os.Open(danmakuPath)
		if err != nil {
			log.Error().Err(err).Msgf("打开弹幕文件失败: %s", danmakuPath)
			continue
		}
		err = xml.NewDecoder(danmakuFile).Decode(&originalDanmaku)
		danmakuFile.Close() // 读取完毕后立即关闭文件
		if err != nil {
			log.Error().Err(err).Msgf("解析弹幕文件失败: %s", danmakuPath)
			continue
		}
		originalNum = originalDanmaku.MaxLimit
		// 获取最新的弹幕
		dmList := au.downloadDanmaku(originalDanmaku.ChatID)
		if len(dmList) == 0 {
			continue
		}
		dm2 := internal.DM2XmlD(dmList)
		latestDmList := internal.MergeDMList(originalDanmaku.Danmaku, dm2)
		// 合并弹幕
		originalDanmaku.Danmaku = latestDmList
		originalDanmaku.MaxLimit = len(latestDmList)

		// 保存弹幕
		xmlData, _ := xml.MarshalIndent(originalDanmaku, "", "    ")
		f, err := os.Create(danmakuPath)
		if err != nil {
			log.Error().Err(err).Msgf("创建文件失败: %s", danmakuPath)
			continue // 添加continue，避免在文件创建失败时尝试写入
		}
		_, err = f.WriteString(xml.Header + string(xmlData))
		if err != nil {
			log.Error().Err(err).Msgf("写入弹幕文件失败: %s", danmakuPath)
		}
		f.Close() // 写入完毕后关闭文件
		log.Debug().Msgf("更新弹幕完成: %s (+%d)条", danmakuPath, len(latestDmList)-originalNum)
	}
}
