package services_test

import (
	"encoder/application/repositories"
	"encoder/application/services"
	"encoder/domain"
	"encoder/framework/database"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load(path.Join("..", "..", ".env"))
	if err != nil {
		log.Fatalf("error loading .env file")
	}
}

func prepareVideo() (*domain.Video, repositories.VideoRepository) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "video.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}
	repo.Insert(video)

	return video, repo
}

func TestVideoServiceUpload(t *testing.T) {

	video, repo := prepareVideo()

	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	err := videoService.Download(os.Getenv("NOT_PROCESSED_VIDEO_BUCKET"))
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	videoUpload := services.NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("PROCESSED_VIDEO_BUCKET")
	videoUpload.VideoPath = path.Join(os.Getenv("LOCAL_STORAGE_PATH"), video.ID, "")

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(50, doneUpload)

	result := <-doneUpload

	require.Equal(t, result, "upload completed")

	t.Logf(result)

	// time.Sleep(5 * time.Second)

	err = videoService.Finish()
	require.Nil(t, err)
}
