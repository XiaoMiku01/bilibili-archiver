package internal

import (
	"fmt"
	"math"

	"os"
	"os/exec"
	"path/filepath"

	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/imroc/req/v3"

	"github.com/rs/zerolog/log"
)

type DownloadUrls struct {
	Url       string // 主下载链接
	BackupUrl string // 备用下载链接
}

type DownloadTask struct {
	GroupID   string       // 任务组ID
	Title     string       // 稿件标题（分p）
	VideoUrls DownloadUrls // 视频下载链接
	AudioUrls DownloadUrls // 音频下载链接
	DirPath   string       // 保存路径
}

type TaskGroup struct {
	TotalTasks     int
	CompletedTasks int
	OnComplete     func(groupID, pdir string)
	mutex          sync.Mutex
}

type DownloaderManager struct {
	taskChan    chan *DownloadTask
	client      *req.Client
	concurrency int
	downloadSem chan struct{} // 用于限制并发下载任务数

	taskGroups map[string]*TaskGroup
	groupMutex sync.RWMutex
}

func NewDownloaderManager() *DownloaderManager {
	taskChan := make(chan *DownloadTask, 10)
	client := req.C().SetCommonHeaders(
		map[string]string{
			"Referer": "https://www.bilibili.com/",
		},
	).SetCommonRetryCount(3).SetTimeout(10 * time.Second).SetCommonRetryFixedInterval(1 * time.Second).
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			// 如果有网络错误或其他HTTP错误，进行重试
			if err != nil {
				log.Warn().Stack().Err(err).Msg("下载失败, 进行重试")
				return true
			}

			return false
		})
	// p := mpb.New(mpb.WithRefreshRate(100 * time.Millisecond))
	return &DownloaderManager{
		taskChan:    taskChan,
		client:      client,
		concurrency: 10,
		downloadSem: make(chan struct{}, 5), // 限制最多5个并发任务
		taskGroups:  make(map[string]*TaskGroup),
	}
}

// 注册任务组并设置完成回调
func (dm *DownloaderManager) RegisterTaskGroup(groupID string, taskCount int, callback func(string, string)) {
	dm.groupMutex.Lock()
	defer dm.groupMutex.Unlock()

	dm.taskGroups[groupID] = &TaskGroup{
		TotalTasks:     taskCount,
		CompletedTasks: 0,
		OnComplete:     callback,
	}
}

// 通知任务组完成情况
func (dm *DownloaderManager) notifyTaskGroupCompletion(groupID, pdir string) {
	if groupID == "" {
		return
	}

	dm.groupMutex.RLock()
	group, exists := dm.taskGroups[groupID]
	dm.groupMutex.RUnlock()

	if !exists {
		return
	}

	group.mutex.Lock()
	group.CompletedTasks++
	completed := group.CompletedTasks >= group.TotalTasks
	group.mutex.Unlock()

	if completed && group.OnComplete != nil {
		// 所有任务已完成，调用回调并移除任务组
		group.OnComplete(groupID, pdir)

		dm.groupMutex.Lock()
		delete(dm.taskGroups, groupID)
		dm.groupMutex.Unlock()
	}
}

// 使用 HEAD 测试 URL 是否有效 如果有效返回 文件大小
func (dm *DownloaderManager) TestUrl(url string) (int64, int, bool) {
	if url == "" {
		return 0, 111, false
	}
	resp := dm.client.Head(url).Do()
	log.Debug().Msgf("%s  %d", url, resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0, resp.StatusCode, false
	}
	contentLength := resp.ContentLength
	return contentLength, resp.StatusCode, true
}

// 检查文件/目录是否存在 如果不存在则创建目录
func (dm *DownloaderManager) CheckPath(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	// 检查文件是否存在 如果存在就返回错误
	if _, err := os.Stat(path); err == nil {
		return os.ErrExist
	}
	return nil
}

func (dm *DownloaderManager) AddTask(task *DownloadTask) {
	dm.taskChan <- task
}

func (dm *DownloaderManager) Run() {
	log.Info().Msg("下载管理器已启动")
	for task := range dm.taskChan {
		// 获取信号量，限制最多5个任务同时下载
		dm.downloadSem <- struct{}{}

		videoPath := task.DirPath + ".mp4.1"
		audioPath := task.DirPath + ".mp3.1"
		outPath := task.DirPath + ".mp4"

		var wg sync.WaitGroup
		wg.Add(2)

		videoDone := make(chan error, 1)
		audioDone := make(chan error, 1)

		go func() {
			err := dm.download(task.VideoUrls, videoPath)
			videoDone <- err
			wg.Done()
		}()

		go func() {
			err := dm.download(task.AudioUrls, audioPath)
			audioDone <- err
			wg.Done()
		}()

		go func() {
			wg.Wait()
			videoErr := <-videoDone
			audioErr := <-audioDone

			defer func() {
				os.Remove(videoPath)
				os.Remove(audioPath)
				// 任务完成后释放信号量
				<-dm.downloadSem
				// log.Info().Msgf("删除临时文件: %s", task.Title)
				// 无论下载成功还是失败，都通知任务组完成状态
				dm.notifyTaskGroupCompletion(task.GroupID, filepath.Dir(outPath))
			}()

			if videoErr != nil || audioErr != nil {
				log.Error().
					Err(videoErr).
					Msgf("视频或音频下载失败，无法合并: %s", task.Title)
				return
			}
			err := dm.merge(videoPath, audioPath, outPath)
			if err != nil {
				log.Error().Err(err).Msgf("合并失败: %s", task.Title)
				return
			}
			log.Info().Msgf("下载并合并完成: %s", task.Title)
		}()
	}
}

func (dm *DownloaderManager) download(durl DownloadUrls, filePath string) error {
	// 1) 分别对 url 和 backupURL 做一次 HEAD 检查
	fileSize, rcode, urlOk := dm.TestUrl(durl.Url)
	fileSize2, rcode2, backupUrlOk := dm.TestUrl(durl.BackupUrl)
	fileBaseName := filepath.Base(filePath)
	if !urlOk && !backupUrlOk && fileSize != fileSize2 {
		return fmt.Errorf("无效的下载链接: %s Code : %d/%d", fileBaseName, rcode, rcode2)
	}

	if err := dm.CheckPath(filePath); err != nil {
		return fmt.Errorf("保存路径检查失败: %s , %v", filePath, err)
	}

	log.Debug().Msgf("开始下载: %s [%d MB]", fileBaseName, fileSize/1024/1024)
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %s , %v", filePath, err)
	}
	defer f.Close()
	if err = f.Truncate(fileSize); err != nil {
		return fmt.Errorf("创建文件失败: %s , %v", filePath, err)
	}

	var downloadErr error
	chunkSize := int64(math.Ceil(float64(fileSize) / float64(dm.concurrency)))
	var wg sync.WaitGroup
	errChan := make(chan error, dm.concurrency)

	restyClient := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetTimeout(10 * time.Second).AddRetryHook(func(resp *resty.Response, err error) {
		log.Warn().Err(err).Msgf("分块下载请求失败: %s, 将在 1s 后重试", fileBaseName)
	})

	for i := range dm.concurrency {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}
		if start >= fileSize {
			break
		}
		// 如果两个 url 都可用，就交替使用；如果只有一个可用，就只用那个
		var url string
		if urlOk && backupUrlOk {
			if i%2 == 0 {
				url = durl.Url
			} else {
				url = durl.BackupUrl
			}
		} else if urlOk {
			url = durl.Url
		} else if backupUrlOk {
			url = durl.BackupUrl
		}
		wg.Add(1)
		go func(start, end int64, dURL string) {
			defer wg.Done()

			var resp *resty.Response
			var err error
			maxRetries := 3
			resp, err = restyClient.R().
				SetHeader("Referer", "https://www.bilibili.com/").
				SetHeader("Range", fmt.Sprintf("bytes=%d-%d", start, end)).
				SetDoNotParseResponse(true).
				Get(dURL)
			if err != nil {
				errChan <- fmt.Errorf("分块下载失败(已重试%d次): %s, %v", maxRetries, fileBaseName, err)
				return
			}

			defer resp.RawBody().Close()
			buf := make([]byte, 1024*1024)
			offset := start
			for {
				n, err := resp.RawBody().Read(buf)
				if n > 0 {
					if _, err = f.WriteAt(buf[:n], offset); err != nil {
						errChan <- err
						break
					}
					offset += int64(n)
				}
				if err != nil {
					break
				}
			}
		}(start, end, url)
	}

	// 等待所有下载协程完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			downloadErr = err
			break
		}
	}

	if downloadErr != nil {
		return downloadErr
	}
	log.Debug().Msgf("下载完成: %s", fileBaseName)
	return nil
}

func (dm *DownloaderManager) merge(videoPath, audioPath, outPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "copy",
		"-y",
		outPath,
	)
	return cmd.Run()
}

var DM = NewDownloaderManager()
