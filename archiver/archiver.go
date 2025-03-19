package archiver

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"

	"github.com/XiaoMiku01/bilibili-archiver/internal"
)

type ArchiverUser struct {
	config internal.Config
	buser  internal.UserInfoStruct

	bapi       *internal.BApiClient
	firstRound bool // 添加标志，表示是否完成第一轮处理
}

func NewArchiverUser(config internal.Config) *ArchiverUser {
	return &ArchiverUser{
		config:     config,
		bapi:       internal.BApi,
		firstRound: true, // 初始化为false，表示第一轮未完成
	}
}

func (au *ArchiverUser) Init() error {
	err := au.bapi.SetCookieFile(au.config.User)
	if err != nil {
		return err
	}
	buser, err := au.bapi.GetUserInfo()
	if err != nil {
		return err
	}
	au.buser = buser

	log.Info().Msgf("用户: %s [UID: %d] 登录成功", au.buser.Uname, au.buser.Mid)

	// au.bapi.GetFavList(au.buser.Mid)
	return nil
}

// 关键词过滤收藏夹
func (au *ArchiverUser) FillerFavoriteList(fvl internal.FavListStruct, kw []string) internal.FavListStruct {
	if len(kw) == 0 {
		return fvl
	}
	var favList internal.FavListStruct
	for _, fav := range fvl.List {
		for _, k := range kw {
			if strings.Contains(fav.Title, k) {
				favList.List = append(favList.List, fav)
			}
		}
	}
	return favList
}

func (au *ArchiverUser) Run() {
	startTime := int(time.Now().Unix()) // 程序启动时间
	var lastRoundTime int = startTime   // 记录上一轮结束的时间
	go internal.DM.Run()                // 启动下载管理器
	go au.UpdateVideoMeta()
	for {
		// currentTime := lastRoundTime // 使用上一轮结束时间作为基准

		// 定义一个变量表示是否全量处理
		var isFull bool
		// 如果是第一轮，且不是增量模式，全量处理
		if au.firstRound && !au.config.Incremental {
			isFull = true
			log.Warn().Msg("全量处理中...")
		} else {
			isFull = false
		}
		au.bapi.InitGRPC()
		favs, err := au.bapi.GetFavList(au.buser.Mid)
		if err != nil {
			log.Error().Err(err).Msg("获取收藏夹列表失败")
			continue
		}
		// 过滤收藏夹
		favs = au.FillerFavoriteList(favs, au.config.Keywords)
		log.Info().Msgf("过滤后收藏夹数量: %d , 过滤关键词: %v ", len(favs.List), au.config.Keywords)
		for _, fav := range favs.List {
			log.Info().Msgf("开始处理收藏夹: %s", fav.Title)

			// 获取收藏夹投稿
			for pn := 1; pn <= fav.MediaCount/40+1; pn++ {
				// au.bapi.GetFavMediaList(fav)
				favMediaList, err := au.bapi.GetFavMediaList(fav.ID, pn)
				if err != nil {
					log.Error().Err(err).Msgf("获取收藏夹投稿: %s pn:%d 失败", fav.Title, pn)
					pn--
					time.Sleep(10 * time.Second)
					continue
				}

				if len(favMediaList.Medias) == 0 {
					// log.Debug().Msgf("收藏夹: %s pn:%d 失败", fav.Title, pn)
					if pn == 1 {
						log.Warn().Msgf("收藏夹: %s pn:%d 无投稿", fav.Title, pn)
					}
					break
				} else {
					// 如果不是全量处理，且最近的 FavTime 小于上一轮结束时间，跳出循环
					if !isFull && favMediaList.Medias[0].FavTime < lastRoundTime {
						log.Debug().Msgf("上一次处理时间: %s, 最近稿件时间: %s", internal.FormatTime(lastRoundTime), internal.FormatTime(favMediaList.Medias[0].FavTime))
						break
					}
				}

				for _, media := range favMediaList.Medias {
					// TODO: 过滤 PGC
					if media.Ugc.FirstCid == 0 {
						continue
					}

					if !isFull && media.FavTime < lastRoundTime {
						log.Debug().Msgf("上一次处理时间: %s, 当前稿件时间: %s, 退出遍历", internal.FormatTime(lastRoundTime), internal.FormatTime(favMediaList.Medias[0].FavTime))
						break
					}

					log.Info().Msgf("开始处理投稿: %s", media.Title)
					vinfo, err := au.bapi.GetView(&internal.ViewReq{
						Aid: media.ID,
					})
					if err != nil {
						log.Error().Err(err).Msgf("获取投稿信息失败: %s", media.Title)
						continue
					}

					// 路径模板替换

					// pdir := filepath.Dir(filepath.Join(au.config.SavePath, dirpath)) // 获取父目录 保存元数据
					au.downloadVideMeta(fav.Title, vinfo, media.FavTime) // 下载投稿元数据
					au.downloadVideo(fav.Title, vinfo, media)            // 下载投稿
					// time.Sleep(10 * time.Second)
				}
				// 获取分页 time.sleep
				log.Debug().Msgf("收藏夹: %s pn:%d 处理完成", fav.Title, pn)
				// time.Sleep(10 * time.Second)
				if isFull {
					time.Sleep(10 * time.Second)
				}
			}
			// 收藏夹处理完成 time.sleep
			log.Debug().Msgf("收藏夹: %s 处理完成", fav.Title)
			// time.Sleep(10 * time.Second)
		}

		// 最外层获取收藏夹循环 time.sleep
		log.Info().Msg("所有收藏夹处理完成, 休眠中...")
		// au.bapi.CloseGRPC()
		// 第一轮已完成，设置标志
		au.firstRound = false
		// 更新 lastRoundTime 为当前时间，这样下次循环会处理这段时间内的新投稿
		lastRoundTime = int(time.Now().Unix())
		time.Sleep(time.Duration(au.config.ScanInterval) * time.Minute)
		au.bapi.GetUserInfo() // 检查登录状态
		tinfo, _ := au.bapi.CheckToken()
		exTime := tinfo.ExpiresIn / 86400
		log.Debug().Msgf("Cookie 有效期: %d 天", exTime)
		if exTime < 7 {
			// au.bapi.RefreshToken()
			log.Warn().Msg("Cookie 即将过期，刷新中...")
			internal.RefreshToken(au.config.User)
		}
	}
}

func (au *ArchiverUser) downloadVideo(favname string, vinfo *internal.ViewReply, media internal.FavMediaStruct) error {
	groupID := vinfo.Bvid

	// 注册任务组，设置回调函数
	internal.DM.RegisterTaskGroup(groupID, len(vinfo.Pages), func(id, pdir string) {
		if len(vinfo.Pages) > 1 {
			log.Info().Msgf("%s 所有分P下载完成", vinfo.Arc.Title)
		}
		// 执行自定义脚本和通知
		if au.config.CustomScript != "" {
			go internal.ExecCommand(au.config.CustomScript, pdir)
		}

		// 通知
		if au.config.Notification != "" {
			msg := `%s-%s.%s (%dP)
已留档完成
%s
`
			msg = fmt.Sprintf(msg, vinfo.Bvid, vinfo.Arc.Title, vinfo.Arc.Author.Name, len(vinfo.Pages), internal.FormatTime(int(time.Now().Unix())))
			log.Info().Msg(msg)
			err := internal.SendNotification(au.config.Notification, msg, au.config.NotificationProxy)
			if err != nil {
				log.Error().Err(err).Msg("发送通知失败")
			} else {
				log.Info().Msg("发送通知成功")
			}
		}
	})

	for i, p := range vinfo.Pages {
		log.Info().Msgf("投稿信息: %s: P%d: cid: %d", vinfo.Bvid, i+1, p.Page.Cid)
		playInfo, err := au.bapi.GetPlayURL(vinfo.Arc.Aid, vinfo.Arc.FirstCid)
		if err != nil {
			log.Error().Err(err).Msgf("获取投稿播放信息失败: %s P%d", media.Title, i+1)
			continue
		}
		if len(playInfo.Dash.Video) == 0 || len(playInfo.Dash.Audio) == 0 {
			log.Error().Msgf("投稿播放信息为空: %s P%d", media.Title, i+1)
			continue
		}
		var quality int
		var vurls internal.DownloadUrls
		var aurls internal.DownloadUrls
		if len(playInfo.Dash.Video) != 0 && len(playInfo.Dash.Audio) != 0 {
			quality = int(playInfo.Dash.Video[0].ID)
			vurls.Url = playInfo.Dash.Video[0].BackupURL[0]
			aurls.Url = playInfo.Dash.Audio[0].BackupURL[0]
			if len(playInfo.Dash.Video[0].BackupURL) != 0 && len(playInfo.Dash.Audio[0].BackupURL) != 0 {
				vurls.BackupUrl = playInfo.Dash.Video[0].BackupURL[0]
				aurls.BackupUrl = playInfo.Dash.Audio[0].BackupURL[0]
			}
		}

		var qualityStr string = "画质未知"
		for _, d := range playInfo.SupportFormats {
			if d.Quality == quality {
				qualityStr = d.NewDescription
			}
		}
		dirpath := internal.FillTemplatePath(au.config.PathTemplate, map[string]string{
			"uname":       au.buser.Uname,
			"fav_name":    favname,
			"date":        internal.FormatDate(media.FavTime),
			"video_title": vinfo.Arc.Title,
			"bv":          vinfo.Bvid,
			"upper_name":  vinfo.Arc.Author.Name,
			"pn":          fmt.Sprintf("%d", i+1),
		})

		downloaderTask := internal.DownloadTask{
			GroupID:   groupID,
			Title:     fmt.Sprintf("[%s]%s P%d", qualityStr, media.Title, i+1),
			VideoUrls: vurls,
			AudioUrls: aurls,
			DirPath:   filepath.Join(au.config.SavePath, dirpath),
		}
		internal.DM.AddTask(&downloaderTask)

		// 下载弹幕
		if au.config.Danmaku {
			dmList := au.downloadDanmaku(p.Page.Cid)
			if len(dmList) != 0 {
				var danmakuXml internal.DanmakuXmlstruct
				danmakuXml.ChatServer = "chat.bilibili.com"
				danmakuXml.ChatID = p.Page.Cid
				danmakuXml.Mission = 0
				danmakuXml.MaxLimit = len(dmList)
				danmakuXml.State = 0
				danmakuXml.RealName = 0
				danmakuXml.Source = "k-v"
				danmakuXml.Danmaku = internal.DM2XmlD(dmList)
				xmlData, _ := xml.MarshalIndent(danmakuXml, "", "    ")
				danmakuPath := filepath.Join(au.config.SavePath, dirpath+"_danmaku.xml")
				f, err := os.Create(danmakuPath)
				if err != nil {
					log.Error().Err(err).Msgf("创建文件失败: %s", danmakuPath)
				}
				defer f.Close()
				f.WriteString(xml.Header)
				f.WriteString(string(xmlData))
				log.Info().Msgf("保存弹幕完成: %s P%d (%d)条", media.Title, i+1, len(dmList))

			} else {
				log.Warn().Msgf("尚未获取到弹幕: %s P%d", media.Title, i+1)
			}
		}
	}
	return nil
}

func (au *ArchiverUser) downloadVideMeta(favname string, vinfo *internal.ViewReply, favtime int) {
	dirpath := internal.FillTemplatePath(au.config.PathTemplate, map[string]string{
		"uname":       au.buser.Uname,
		"fav_name":    favname,
		"date":        internal.FormatDate(favtime),
		"video_title": vinfo.Arc.Title,
		"bv":          vinfo.Bvid,
		"upper_name":  vinfo.Arc.Author.Name,
		"pn":          fmt.Sprintf("%d", 1),
	})
	pdir := filepath.Dir(filepath.Join(au.config.SavePath, dirpath)) // 获取父目录 保存元数据
	err := os.MkdirAll(pdir, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msgf("创建目录失败: %s", pdir)
		return
	}
	filename := filepath.Join(au.config.SavePath, dirpath+"_meta.json")
	jsonData, _ := json.MarshalIndent(vinfo.Arc, "", "  ")
	f, err := os.Create(filename)
	if err != nil {
		log.Error().Err(err).Msgf("创建文件失败: %s", filename)
		return
	}
	defer f.Close()
	f.WriteString(string(jsonData))
	// 下载封面
	coverPath := filepath.Join(au.config.SavePath, dirpath+"_cover.jpg")
	_, err = req.SetOutputFile(coverPath).Get(vinfo.Arc.Pic)
	if err != nil {
		log.Error().Err(err).Msgf("下载封面失败: %s", vinfo.Arc.Pic)
	}
	log.Info().Msgf("保存投稿元数据完成: %s", vinfo.Arc.Title)
}

func (au *ArchiverUser) downloadDanmaku(cid int64) []*internal.DanmakuStruct {
	var segIndex int64 = 1
	var danmakuList []*internal.DanmakuStruct
	for {
		danmaku, err := au.bapi.GetDanmaku(&internal.DmSegMobileReq{
			Type:         1,
			Oid:          cid,
			SegmentIndex: segIndex,
		})
		if err != nil {
			// log.Error().Err(err).Msgf("获取弹幕失败: cid: %d, segIndex: %d", cid, segIndex)
			break
		}
		if len(danmaku.Elems) == 0 {
			break
		}
		danmakuList = append(danmakuList, danmaku.Elems...)
		segIndex++
	}
	return danmakuList
}
