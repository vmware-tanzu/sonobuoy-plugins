package main

import (
	"flag"
	"log"
	"strings"

	"github.com/vladimirvivien/echo"
)

var(
	tasks = map[string]func(e *echo.Echo){
		"all": all,
		"docker_build": dockerBuild,
		"docker_push": dockerPush,
	}
)

// make.go used for CI builds
// From project's root directory, run the followings for info
// $ go run .ci/* --help
func main() {
	var version string
	var targets string
	flag.StringVar(&version, "version", "0.6.5", "kube-bench image version to build")
	flag.StringVar(&targets, "targets", "all", "comma-separated list of build targets")
	flag.Parse()

	e := echo.New()
	e.Conf.SetPanicOnErr(true)
	e.SetVar("tag", version)
	e.SetVar("docker_org", "sonobuoy")
	e.SetVar("image_name", "kube-bench")
	e.SetVar("repository", "${docker_org}/${image_name}:${tag}")
	e.SetVar("repository_latest", "${docker_org}/${image_name}:latest")

	for _, target := range strings.Split(targets,",") {
		if task, ok := tasks[strings.TrimSpace(target)]; ok {
			task(e)
		}
	}
}

func all(e *echo.Echo) {
	dockerBuild(e)
	dockerPush(e)
}

func dockerBuild(e *echo.Echo) {
	log.Println(e.Eval("Building docker image: ${repository}, ${repository_latest}"))
	e.Runout("docker build -t ${repository} -t ${repository_latest} -f Dockerfile .")
}

func dockerPush(e *echo.Echo) {
	log.Println(e.Eval("Pushing docker image: ${repository}"))
	e.Run("docker push ${repository}")

	log.Println(e.Eval("Pushing docker image: ${repository_latest}"))
	e.Run("docker push ${repository_latest}")
}
