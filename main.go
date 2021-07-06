package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

const remoteOrigin = "origin"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()
	e.GET("/build", build)
	e.GET("/hello-world", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "up & running"})
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func build(c echo.Context) error {
	var err error
	text := c.QueryParam("text")
	params := strings.Split(text, " ")
	if len(params) < 3 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "insufficient params"})
	}

	var tokenProject = map[string]string{
		"hello-world": os.Getenv("TOKEN_HELLO_WORLD"),
		"freestyle":   os.Getenv("TOKEN_FREESTYLE"),
	}

	jenkinsJob := params[0]
	repo := params[1]
	branch := params[2]
	commitHash := params[3]

	if err = TagExecCmd(repo, branch, commitHash); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	jenkinsToken := tokenProject[jenkinsJob]
	jenkinsURL := os.Getenv("JENKINS_URL")

	baseURL := fmt.Sprintf(
		"%s/buildByToken/buildWithParameters?job=%s&token=%s&text=%s",
		jenkinsURL,
		jenkinsJob,
		jenkinsToken,
		url.QueryEscape(branch))

	jenkinsResponse, err := http.Get(baseURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	defer jenkinsResponse.Body.Close()

	if jenkinsResponse.StatusCode >= http.StatusInternalServerError {
		err = fmt.Errorf("error while fetch data from server with code %d", jenkinsResponse.StatusCode)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"text": text, "url": baseURL, "jenkinsResponse": jenkinsResponse.StatusCode})
}

func TagExecCmd(repo, branch, commitHash string) error {
	var err error

	username := os.Getenv("GITHUB_USER")
	ghToken := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")

	cmdGitInit := fmt.Sprintf("git init")
	if _, err = CmdExec(cmdGitInit); err != nil {
		return err
	}

	// get remote
	cmdGetRemote := fmt.Sprintf("git remote -v | awk '{print $1;}' ")
	strOutput, err := CmdExec(cmdGetRemote)
	if err != nil && strOutput != remoteOrigin {
		return err
	}

	// init
	if strOutput == remoteOrigin {
		cmdSetRmRemove := fmt.Sprintf("git remote rm %s", remoteOrigin)
		if _, err = CmdExec(cmdSetRmRemove); err != nil {
			return err
		}
	}

	cmdSetRemote := fmt.Sprintf("git remote add %s https://%s:%s@github.com/%s/%s.git", remoteOrigin, username, ghToken, owner, repo)
	if _, err = CmdExec(cmdSetRemote); err != nil {
		return err
	}

	cmdFetch := fmt.Sprintf("git fetch --all")
	if _, err = CmdExec(cmdFetch); err != nil {
		_ = fmt.Sprintf("git remote rm %s", remoteOrigin)
		return err
	}

	cmdSetTag := fmt.Sprintf("git tag -f %s %s", branch, commitHash)
	if _, err = CmdExec(cmdSetTag); err != nil {
		_ = fmt.Sprintf("git remote rm %s", remoteOrigin)
		return err
	}

	cmdPushTag := fmt.Sprintf("git push -f origin %s", branch)
	if _, err = CmdExec(cmdPushTag); err != nil {
		_ = fmt.Sprintf("git remote rm %s", remoteOrigin)
		return err
	}

	cmdRmRemote := fmt.Sprintf("git remote rm %s", remoteOrigin)
	if _, err = CmdExec(cmdRmRemote); err != nil {
		return err
	}

	return nil

}

func CmdExec(cmdLine string) (string, error) {
	out := new(strings.Builder)
	var stderr bytes.Buffer

	command := exec.Command("sh", "-c", cmdLine)
	command.Stdout = out
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	if out.String() != "" {
		getOrigin := strings.Split(out.String(), "\n")
		for _, v := range getOrigin {
			if v == remoteOrigin {
				return v, nil
			}
		}
	}

	return "", nil
}

func CreateTagGo(token, branch, username, repo, commitHash string) (*github.Tag, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	gitObjectCommit := "commit"
	gitObject := github.GitObject{
		Type: &gitObjectCommit,
		SHA:  &commitHash,
	}
	Tag := github.Tag{
		Tag:     &branch,
		SHA:     &commitHash,
		Object:  &gitObject,
		Message: &branch,
	}
	tagResponse, _, err := client.Git.CreateTag(ctx, username, repo, &Tag)
	if err != nil {
		return nil, err
	}

	return tagResponse, nil
}
