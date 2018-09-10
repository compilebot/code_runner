package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/docker/docker/client"
)

// NewBuild is the function to create a new container and run the code passed into it
func NewBuild(code, lang string) string {
	rand.Seed(time.Now().Unix())
	id := generateID(lang)
	err := writeCodeToFile(code, lang)

	if err != nil {
		panic(err)
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	buildImage(cli, id, lang)
	x := buildContainer(cli, id)

	time.Sleep(5 * time.Second)

	y := getLogs(x, cli)
	fmt.Println(y)

	cleanup(cli, id)

	return y

}

func buildImage(cli *client.Client, id, lang string) {

	config := types.ImageBuildOptions{Tags: []string{id}}
	createBuildContext(lang)
	buildContext, err := os.Open(fmt.Sprintf("/tmp/code_runner/%s.tar", lang))
	if err != nil {
		panic(err)
	}
	defer buildContext.Close()

	br, err := cli.ImageBuild(context.Background(), buildContext, config)
	if err != nil {
		panic(err)
	}
	defer br.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(br.Body)
	s := buf.String()
	fmt.Println(s) //do not remove

}

func createBuildContext(lang string) {
	file, err := os.Create(fmt.Sprintf("/tmp/code_runner/%s.tar", lang))
	if err != nil {
		panic(err)
	}

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tw := tar.NewWriter(gzipWriter)
	defer tw.Close()

	dkfile := fmt.Sprintf("Dockerfile-%s", lang)
	err = copyFile("/tmp/code_runner/Dockerfile", "./dockerfiles/"+dkfile)

	files := map[string][]byte{"Dockerfile": nil, "main.go": nil}
	for k, v := range files {
		a, err := ioutil.ReadFile("/tmp/code_runner/" + k)
		v = a
		if err != nil {
			panic(err)
		}
		hdr := &tar.Header{Name: k, Mode: 0600, Size: int64(len(v))}

		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}

		if _, err := tw.Write([]byte(v)); err != nil {
			panic(err)
		}

	}

}

func cleanup(cli *client.Client, id string) error {
	config := types.ContainerRemoveOptions{Force: true}
	ctx := context.Background()

	err := cli.ContainerRemove(ctx, id, config)

	if err != nil {
		return err
	}

	_, err = cli.ImageRemove(ctx, id, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}

func writeCodeToFile(code, lang string) error {
	path := fmt.Sprintf("/tmp/code_runner/main.%s", lang)
	os.Mkdir("/tmp/code_runner", 0777)
	os.Create(path)
	f, err := os.OpenFile(path, os.O_RDWR, 0777)

	if err != nil {
		return err
	}

	if _, err := f.WriteString(code); err != nil {
		return err
	}

	return nil
}

func buildContainer(cli *client.Client, id string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := container.Config{Image: id}
	hostconfig := container.HostConfig{}
	netconfig := network.NetworkingConfig{}

	c, err := cli.ContainerCreate(ctx, &config, &hostconfig, &netconfig, id)
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return c.ID
}

func getLogs(id string, cli *client.Client) string {
	var b strings.Builder

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reader, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(&b, reader)
	if err != nil && err != io.EOF {
		fmt.Println(err)
	}

	return b.String()
}
