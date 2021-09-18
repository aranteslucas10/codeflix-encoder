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

func prepareJob() (*domain.Job, repositories.JobRepository, services.VideoService) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "video.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}
	repo.Insert(video)

	job, err := domain.NewJob(os.Getenv("PROCESSED_VIDEO_BUCKET"), "NEW", video)
	if err != nil {
		log.Fatalf("error creating jog: %v", err)
	}

	repoJob := repositories.JobRepositoryDb{Db: db}

	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	return job, repoJob, videoService
}

func TestJobService(t *testing.T) {

	job, repo, videoService := prepareJob()

	require.NotNil(t, job)
	require.NotNil(t, repo)
	require.NotNil(t, videoService)

	j := services.JobService{
		Job:           job,
		JobRepository: repo,
		VideoService:  videoService,
	}

	require.NotNil(t, j)

	require.Equal(t, job.ID, j.Job.ID)

	err := j.Start()

	require.Nil(t, err)
}
