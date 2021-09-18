package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"cloud.google.com/go/storage"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(bucketName string) error {

	ctx := context.Background()

	client, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(v.Video.FilePath)

	reader, err := obj.NewReader(ctx)

	if err != nil {
		return err
	}

	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID+".mp4"))
	if err != nil {
		return err
	}

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	defer file.Close()

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {

	err := os.Mkdir(path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID), os.ModePerm)
	if err != nil {
		return err
	}

	source := path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID+".mp4")
	target := path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID+".frag")

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID+".frag"))
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID))
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")

	cmd := exec.Command("mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Finish() error {
	err := os.Remove(path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID+".mp4"))
	if err != nil {
		log.Println("error removing mp4 ", v.Video.ID, ".mp4")
		return err
	}

	err = os.Remove(path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID+".frag"))
	if err != nil {
		log.Println("error removing mp4 ", v.Video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(path.Join(os.Getenv("LOCAL_STORAGE_PATH"), v.Video.ID))
	if err != nil {
		log.Println("error removing folder ", v.Video.ID)
		return err
	}

	log.Println("files have been removed ", v.Video.ID)

	return nil
}

func (v *VideoService) InsertVideo() error {
	_, err := v.VideoRepository.Insert(v.Video)

	if err != nil {
		return err
	}

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("=====> Output: %s\n", string(out))
	}
}
